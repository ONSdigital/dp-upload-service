package files

import (
	"context"
	"errors"
	"net/http"
	"strings"

	filesAPI "github.com/ONSdigital/dp-api-clients-go/v2/files"
	filesAPITypes "github.com/ONSdigital/dp-files-api/files"
	filesSDK "github.com/ONSdigital/dp-files-api/sdk"
	s3client "github.com/ONSdigital/dp-s3/v3"
	"github.com/ONSdigital/dp-upload-service/aws"
	"github.com/ONSdigital/dp-upload-service/config"
	"github.com/ONSdigital/log.go/v2/log"
)

//go:generate moq -out mock/files.go -pkg mock_files . FilesClienter

var (
	ErrFilesAPIDuplicateFile    = errors.New("files API already contains a file with this path")
	ErrFilesAPINotFound         = errors.New("cannot find a file with this path")
	ErrFileAPICreateInvalidData = errors.New("invalid data sent to Files API")
	ErrS3Upload                 = errors.New("uploading part failed")
	ErrS3Download               = errors.New("downloading file failed")
	ErrS3Head                   = errors.New("getting file info failed")
	ErrChunkTooSmall            = errors.New("chunk size below minimum 5MB")
	ErrFilesServer              = errors.New("file api returning internal server errors")
	ErrFilesUnauthorised        = errors.New("access unauthorised")
)

// FileMetadataWithContentItem extends the files API metadata with content_item
type FileMetadataWithContentItem struct {
	filesAPI.FileMetaData
	ContentItem *ContentItem `json:"content_item,omitempty"`
}

type ContentItem struct {
	DatasetID string `json:"dataset_id"`
	Edition   string `json:"edition"`
	Version   string `json:"version"`
}

type FilesClienter interface {
	GetFile(ctx context.Context, path string, headers filesSDK.Headers) (*filesAPITypes.StoredRegisteredMetaData, error)
	RegisterFile(ctx context.Context, metadata filesAPITypes.StoredRegisteredMetaData, headers filesSDK.Headers) error
	MarkFilePublished(ctx context.Context, path string, headers filesSDK.Headers) error
	MarkFileUploaded(ctx context.Context, path string, etag string, headers filesSDK.Headers) error
}

type Store struct {
	files  FilesClienter
	bucket *aws.Bucket
	cfg    *config.Config
}

type Resumable struct {
	FileName     string `schema:"resumableFilename"`
	Type         string `schema:"resumableType"`
	CurrentChunk int32  `schema:"resumableChunkNumber"`
	TotalChunks  int    `schema:"resumableTotalChunks"`
}

type StatusMessage struct {
	Value bool   `json:"valid"`
	Err   string `json:"error,omitempty"`
}

type Status struct {
	Metadata    filesAPITypes.FileMetaData `json:"metadata"`
	FileContent StatusMessage              `json:"file_content"`
}

func NewStore(files FilesClienter, bucket *aws.Bucket, cfg *config.Config) Store {
	return Store{files, bucket, cfg}
}

func (s Store) Status(ctx context.Context, path string) (*Status, error) {
	headers := filesSDK.Headers{
		Authorization: s.cfg.ServiceAuthToken,
	}

	//metadata
	storedMetadata, err := s.files.GetFile(ctx, path, headers)
	if err != nil {
		log.Error(ctx, "failed to get file metadata", err, log.Data{"path": path})
		return nil, ErrFilesAPINotFound
	}

	metadata := filesAPITypes.FileMetaData{
		Path:          storedMetadata.Path,
		IsPublishable: storedMetadata.IsPublishable,
		CollectionID:  storedMetadata.CollectionID,
		Title:         storedMetadata.Title,
		SizeInBytes:   storedMetadata.SizeInBytes,
		Type:          storedMetadata.Type,
		Licence:       storedMetadata.Licence,
		LicenceURL:    storedMetadata.LicenceURL,
		State:         storedMetadata.State,
		Etag:          storedMetadata.Etag,
	}

	//file content
	head, err := s.bucket.Head(ctx, path)
	fileContent := newStatusMessage(head != nil && head.ContentLength != nil && *head.ContentLength > 0, err)

	return &Status{
		Metadata:    metadata,
		FileContent: fileContent,
	}, nil
}

// registerFileWithContentItem uses the dp-files-api SDK to register file with content_item
func (s Store) registerFileWithContentItem(ctx context.Context, metadata FileMetadataWithContentItem) error {
	var contentItem *filesAPITypes.StoredContentItem
	if metadata.ContentItem != nil {
		contentItem = &filesAPITypes.StoredContentItem{
			DatasetID: metadata.ContentItem.DatasetID,
			Edition:   metadata.ContentItem.Edition,
			Version:   metadata.ContentItem.Version,
		}
	}

	storedMetadata := filesAPITypes.StoredRegisteredMetaData{
		Path:          metadata.Path,
		IsPublishable: metadata.IsPublishable,
		CollectionID:  metadata.CollectionID,
		BundleID:      metadata.BundleID,
		Title:         metadata.Title,
		SizeInBytes:   metadata.SizeInBytes,
		Type:          metadata.Type,
		Licence:       metadata.Licence,
		LicenceURL:    metadata.LicenceUrl,
		ContentItem:   contentItem,
	}

	headers := filesSDK.Headers{
		Authorization: s.cfg.ServiceAuthToken,
	}

	err := s.files.RegisterFile(ctx, storedMetadata, headers)

	if err != nil {
		if apiErr, ok := err.(*filesSDK.APIError); ok {
			switch apiErr.StatusCode {
			case http.StatusConflict:
				return filesAPI.ErrFileAlreadyRegistered
			case http.StatusBadRequest:
				return ErrFileAPICreateInvalidData
			case http.StatusForbidden:
				return ErrFilesUnauthorised
			case http.StatusInternalServerError:
				return ErrFilesServer
			}
		}
	}

	return err
}

func (s Store) UploadFile(ctx context.Context, metadata FileMetadataWithContentItem, resumable Resumable, content []byte) (bool, error) {
	baseMetadata := metadata.FileMetaData

	part := generateUploadPart(baseMetadata, resumable)
	response, err := s.bucket.UploadPart(ctx, part, content)
	if err != nil {
		log.Error(ctx, "failed to write chunk to s3", err, log.Data{"s3-upload-part": part})
		if _, ok := err.(*s3client.ErrChunkTooSmall); ok {
			return false, ErrChunkTooSmall
		}
		return false, ErrS3Upload
	}

	if response.AllPartsUploaded {
		if err = s.registerFileWithContentItem(ctx, metadata); err != nil {
			log.Error(ctx, "failed to register file metadata with dp-files-api", err, log.Data{"metadata": metadata})
			return false, err
		}

		head, err := s.bucket.Head(ctx, baseMetadata.Path)
		if err != nil {
			log.Error(ctx, "failed to get completed file info from s3", err, log.Data{"key": baseMetadata.Path})
			return false, ErrS3Head
		}
		if head.ETag == nil {
			log.Error(ctx, "failed to get completed file etag from s3", err, log.Data{"key": baseMetadata.Path})
			return false, ErrS3Head
		}

		headers := filesSDK.Headers{
			Authorization: s.cfg.ServiceAuthToken,
		}
		return true, s.files.MarkFileUploaded(ctx, baseMetadata.Path, strings.Trim(*head.ETag, "\""), headers)
	}

	return false, nil
}

func generateUploadPart(metadata filesAPI.FileMetaData, resumable Resumable) *s3client.UploadPartRequest {
	return &s3client.UploadPartRequest{
		UploadKey:   metadata.Path,
		Type:        resumable.Type,
		ChunkNumber: resumable.CurrentChunk,
		TotalChunks: resumable.TotalChunks,
		FileName:    resumable.FileName,
	}
}

func newStatusMessage(val bool, err error) StatusMessage {
	msg := StatusMessage{
		Value: val && err == nil,
	}
	if err != nil {
		msg.Err = err.Error()
	}
	return msg
}
