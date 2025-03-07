package files

import (
	"context"
	"errors"
	"strings"

	"github.com/ONSdigital/dp-api-clients-go/v2/files"
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

type FilesClienter interface {
	GetFile(ctx context.Context, path string, authToken string) (files.FileMetaData, error)
	RegisterFile(ctx context.Context, metadata files.FileMetaData) error
	MarkFileUploaded(ctx context.Context, path string, etag string) error
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
	Metadata    files.FileMetaData `json:"metadata"`
	FileContent StatusMessage      `json:"file_content"`
}

func NewStore(files FilesClienter, bucket *aws.Bucket, cfg *config.Config) Store {
	return Store{files, bucket, cfg}
}

func (s Store) Status(ctx context.Context, path string) (*Status, error) {
	//metadata
	metadata, err := s.files.GetFile(ctx, path, s.cfg.ServiceAuthToken)
	if err != nil {
		log.Error(ctx, "failed to get file metadata", err, log.Data{"path": path})
		return nil, ErrFilesAPINotFound
	}

	//file content
	head, err := s.bucket.Head(ctx, path)
	fileContent := newStatusMessage(head != nil && head.ContentLength != nil && *head.ContentLength > 0, err)

	return &Status{
		Metadata:    metadata,
		FileContent: fileContent,
	}, nil
}

func (s Store) UploadFile(ctx context.Context, metadata files.FileMetaData, resumable Resumable, content []byte) (bool, error) {
	part := generateUploadPart(metadata, resumable)
	response, err := s.bucket.UploadPart(ctx, part, content)
	if err != nil {
		log.Error(ctx, "failed to write chunk to s3", err, log.Data{"s3-upload-part": part})
		if _, ok := err.(*s3client.ErrChunkTooSmall); ok {
			return false, ErrChunkTooSmall
		}
		return false, ErrS3Upload
	}

	if response.AllPartsUploaded {
		var err error
		if err = s.files.RegisterFile(ctx, metadata); err != nil {
			log.Error(ctx, "failed to register file metadata with dp-files-api", err, log.Data{"metadata": metadata})
			return false, err
		}

		head, err := s.bucket.Head(ctx, metadata.Path)
		if err != nil {
			log.Error(ctx, "failed to get completed file info from s3", err, log.Data{"key": metadata.Path})
			return false, ErrS3Head
		}
		if head.ETag == nil {
			log.Error(ctx, "failed to get completed file etag from s3", err, log.Data{"key": metadata.Path})
			return false, ErrS3Head
		}

		return true, s.files.MarkFileUploaded(ctx, metadata.Path, strings.Trim(*head.ETag, "\""))
	}

	return false, nil
}

func generateUploadPart(metadata files.FileMetaData, resumable Resumable) *s3client.UploadPartRequest {
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
