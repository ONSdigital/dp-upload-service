package files

import (
	"context"
	"errors"
	"strings"

	"github.com/ONSdigital/dp-api-clients-go/v2/files"
	s3client "github.com/ONSdigital/dp-s3/v2"
	"github.com/ONSdigital/dp-upload-service/encryption"
	"github.com/ONSdigital/dp-upload-service/upload"
	"github.com/ONSdigital/log.go/v2/log"
)

var (
	ErrFilesAPIDuplicateFile    = errors.New("files API already contains a file with this path")
	ErrFileAPICreateInvalidData = errors.New("invalid data sent to Files API")
	ErrS3Upload                 = errors.New("uploading part failed")
	ErrChunkTooSmall            = errors.New("chunk size below minimum 5MB")
	ErrFilesServer              = errors.New("file api returning internal server errors")
	ErrFilesUnauthorised        = errors.New("access unauthorised")
)

type Store struct {
	files upload.FilesClienter
	s3    upload.S3Clienter
	vault *encryption.Vault
}

func NewStore(files upload.FilesClienter, s3 upload.S3Clienter, vault *encryption.Vault) Store {
	return Store{files, s3, vault}
}

func (s Store) UploadFile(ctx context.Context, metadata files.FileMetaData, resumable Resumable, content []byte) (bool, error) {
	var encryptionKey []byte
	var err error

	if firstChunk(resumable.CurrentChunk) {
		if err = s.files.RegisterFile(ctx, metadata); err != nil {
			log.Error(ctx, "failed to register file metadata with dp-files-api", err, log.Data{"metadata": metadata})
			return false, err
		}

		encryptionKey, err = s.vault.GenerateEncryptionKey(ctx, metadata.Path)
		if err != nil {
			return false, err
		}
	} else {
		encryptionKey, err = s.vault.EncryptionKey(ctx, metadata.Path)
		if err != nil {
			return false, err
		}
	}

	part := s.generateUploadPart(metadata, resumable)
	response, err := s.s3.UploadPartWithPsk(ctx, part, content, encryptionKey)
	if err != nil {
		log.Error(ctx, "failed to write chunk to s3", err, log.Data{"s3-upload-part": part})
		if _, ok := err.(*s3client.ErrChunkTooSmall); ok {
			return false, ErrChunkTooSmall
		}
		return false, ErrS3Upload
	}

	if response.AllPartsUploaded {
		return true, s.files.MarkFileUploaded(ctx, metadata.Path, strings.Trim(response.Etag, "\""))
	}

	return false, nil
}

func firstChunk(currentChunk int64) bool { return currentChunk == 1 }

func (s Store) generateUploadPart(metadata files.FileMetaData, resumable Resumable) *s3client.UploadPartRequest {
	return &s3client.UploadPartRequest{
		UploadKey:   metadata.Path,
		Type:        resumable.Type,
		ChunkNumber: resumable.CurrentChunk,
		TotalChunks: resumable.TotalChunks,
		FileName:    resumable.FileName,
	}
}

type Resumable struct {
	FileName     string `schema:"resumableFilename"`
	Type         string `schema:"resumableType"`
	CurrentChunk int64  `schema:"resumableChunkNumber"`
	TotalChunks  int    `schema:"resumableTotalChunks"`
}
