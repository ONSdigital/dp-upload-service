package files

import (
	"context"
	"encoding/hex"
	"errors"
	"fmt"
	"strings"

	"github.com/ONSdigital/log.go/v2/log"

	"github.com/ONSdigital/dp-upload-service/encryption"

	s3client "github.com/ONSdigital/dp-s3/v2"

	"github.com/ONSdigital/dp-upload-service/upload"

	"github.com/ONSdigital/dp-api-clients-go/v2/files"
)

const (
	StateUploaded = "UPLOADED"
	vaultKey      = "key"
)

var (
	ErrFilesAPIDuplicateFile    = errors.New("files API already contains a file with this path")
	ErrFileAPICreateInvalidData = errors.New("invalid data sent to Files API")
	ErrUnknownError             = errors.New("unknown error")
	ErrConnectingToFilesApi     = errors.New("could not connect to files API")
	ErrS3Upload                 = errors.New("uploading part failed")
	ErrFileNotFound             = errors.New("file not found")
	ErrFileStateConflict        = errors.New("file was not in the expected state")
	ErrChunkTooSmall            = errors.New("chunk size below minimum 5MB")
	ErrVaultWrite               = errors.New("failed to write to vault")
	ErrVaultRead                = errors.New("failed to read from vault")
	ErrInvalidEncryptionKey     = errors.New("encryption key invalid")
	ErrFilesServer              = errors.New("file api returning internal server errors")
	ErrFilesUnauthorised        = errors.New("access unauthorised")
)

type ContextKey string

type Store struct {
	files        upload.FilesClienter
	s3           upload.S3Clienter
	keyGenerator encryption.GenerateKey
	vault        upload.VaultClienter
	vaultPath    string
}

func NewStore(files upload.FilesClienter, s3 upload.S3Clienter, keyGenerator encryption.GenerateKey, vault upload.VaultClienter, vaultPath string) Store {
	return Store{files, s3, keyGenerator, vault, vaultPath}
}

func (s Store) UploadFile(ctx context.Context, metadata files.FileMetaData, resumable Resumable, content []byte) (bool, error) {
	var encryptionKey []byte
	var err error

	if firstChunk(resumable.CurrentChunk) {
		if err = s.files.RegisterFile(ctx, metadata); err != nil {
			log.Error(ctx, "failed to register file metadata with dp-files-api", err, log.Data{"metadata": metadata})
			return false, err
		}

		encryptionKey, err = s.generateEncryptionKey(ctx, metadata.Path)
		if err != nil {
			return false, err
		}
	} else {
		encryptionKey, err = s.getEncryptionKey(ctx, metadata.Path)
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

func (s Store) generateEncryptionKey(ctx context.Context, filepath string) ([]byte, error) {
	encryptionKey := s.keyGenerator()
	if err := s.vault.WriteKey(s.getVaultPath(filepath), vaultKey, hex.EncodeToString(encryptionKey)); err != nil {
		log.Error(ctx, "failed to write encryption encryptionKey to vault", err, log.Data{"vault-path": s.getVaultPath(filepath), "vault-encryptionKey": vaultKey})
		return nil, ErrVaultWrite
	}

	return encryptionKey, nil
}

func (s Store) getEncryptionKey(ctx context.Context, filepath string) ([]byte, error) {
	strKey, err := s.vault.ReadKey(s.getVaultPath(filepath), vaultKey)
	if err != nil {
		log.Error(ctx, "failed to read encryption encryptionkey from vault", err, log.Data{"vault-path": s.getVaultPath(filepath), "vault-encryptionkey": vaultKey})
		return nil, ErrVaultRead
	}

	encryptionKey, err := hex.DecodeString(strKey)
	if err != nil {
		log.Error(ctx, "encryption key contains non-hexadecimal characters", err, log.Data{"vault-path": s.getVaultPath(filepath), "vault-encryptionkey": vaultKey})
		return nil, ErrInvalidEncryptionKey
	}
	return encryptionKey, nil
}

func (s Store) getVaultPath(filepath string) string {
	return fmt.Sprintf("%s/%s", s.vaultPath, filepath)
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
