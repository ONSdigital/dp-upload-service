package files

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"

	"github.com/ONSdigital/dp-upload-service/encryption"

	s3client "github.com/ONSdigital/dp-s3/v2"

	"github.com/ONSdigital/dp-upload-service/upload"
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
)

type Store struct {
	hostname     string
	s3           upload.S3Clienter
	keyGenerator encryption.GenerateKey
	vault        upload.VaultClienter
	vaultPath    string
}

func NewStore(hostname string, s3 upload.S3Clienter, keyGenerator encryption.GenerateKey, vault upload.VaultClienter, vaultPath string) Store {
	return Store{hostname, s3, keyGenerator, vault, vaultPath}
}

type Metadata struct {
	Path          string `schema:"path" json:"path" validate:"required,aws-upload-key"`
	IsPublishable bool   `schema:"isPublishable" json:"is_publishable" validate:"required"`
	CollectionId  string `schema:"collectionId" json:"collection_id" validate:"required"`
	Title         string `schema:"title" json:"title"`
	SizeInBytes   int    `schema:"sizeInBytes" json:"size_in_bytes" validate:"required"`
	Type          string `schema:"type" json:"type" validate:"required,mime-type"`
	Licence       string `schema:"licence" json:"licence" validate:"required"`
	LicenceUrl    string `schema:"licenceUrl" json:"licence_url" validate:"required"`
}

type Resumable struct {
	Path         string `schema:"path"`
	Type         string `schema:"type"`
	CurrentChunk int64  `schema:"currentChunk"`
	TotalChunks  int    `schema:"totalChunks"`
}

type uploadComplete struct {
	Path string `json:"path"`
	ETag string `json:"etag"`
}

type jsonError struct {
	Code        string `json:"code"`
	Description string `json:"description"`
}

type jsonErrors struct {
	Error []jsonError `json:"errors"`
}

func firstChunk(currentChunk int64) bool { return currentChunk == 1 }

func (s Store) UploadFile(ctx context.Context, metadata Metadata, resumable Resumable, content []byte) (bool, error) {

	var key []byte
	if firstChunk(resumable.CurrentChunk) {
		key = s.keyGenerator()
		s.vault.WriteKey(fmt.Sprintf("%s/%s", s.vaultPath, metadata.Path), "key", string(key)) // nolint error handling in next ticket
		if err := s.registerFileUpload(metadata); err != nil {
			return false, err
		}
	} else {
		strKey, _ := s.vault.ReadKey(fmt.Sprintf("%s/%s", s.vaultPath, metadata.Path), "key") // notlint error handling in next ticket
		key = []byte(strKey)
	}

	upr := s3client.UploadPartRequest{
		UploadKey:   resumable.Path,
		Type:        resumable.Type,
		ChunkNumber: resumable.CurrentChunk,
		TotalChunks: resumable.TotalChunks,
		FileName:    resumable.Path,
	}

	response, err := s.s3.UploadPartWithPsk(ctx, &upr, content, key)
	if err != nil {
		if _, ok := err.(*s3client.ErrChunkTooSmall); ok {
			return false, ErrChunkTooSmall
		}
		return false, ErrS3Upload
	}

	uc := uploadComplete{
		Path: metadata.Path,
		ETag: strings.Trim(response.Etag, "\""),
	}

	if response.AllPartsUploaded {
		if err := s.markUploadComplete(uc); err != nil {
			return true, err
		}
	}

	return response.AllPartsUploaded, nil
}

func (s Store) markUploadComplete(uc uploadComplete) error {
	resp, err := http.Post(fmt.Sprintf("%s/v1/files/upload-complete", s.hostname), "application/json", jsonEncode(uc))
	if err != nil {
		return ErrConnectingToFilesApi
	}

	if resp.StatusCode == http.StatusNotFound {
		return ErrFileNotFound
	} else if resp.StatusCode == http.StatusConflict {
		return ErrFileStateConflict
	} else if resp.StatusCode != http.StatusCreated {
		return ErrUnknownError
	}

	return nil
}

func (s Store) registerFileUpload(metadata Metadata) error {
	resp, err := http.Post(fmt.Sprintf("%s/v1/files/register", s.hostname), "application/json", jsonEncode(metadata))
	if err != nil {
		return ErrConnectingToFilesApi
	}

	if resp.StatusCode == http.StatusBadRequest {
		var err error

		jsonErrors := jsonErrors{}
		err = json.NewDecoder(resp.Body).Decode(&jsonErrors)
		if err != nil {
			return err
		}

		switch jsonErrors.Error[0].Code {
		case "DuplicateFileError":
			err = ErrFilesAPIDuplicateFile
		case "ValidationError":
			err = ErrFileAPICreateInvalidData
		default:
			err = ErrUnknownError
		}

		return err
	}
	return nil
}

func jsonEncode(data interface{}) *bytes.Buffer {
	b := &bytes.Buffer{}
	json.NewEncoder(b).Encode(data) // nolint // Only fails due to coding error
	return b
}
