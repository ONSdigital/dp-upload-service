package files

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/ONSdigital/dp-upload-service/api"
	"github.com/ONSdigital/dp-upload-service/encryption"
	"net/http"
	"strings"

	dphttp "github.com/ONSdigital/dp-net/v2/http"
	"github.com/ONSdigital/dp-net/v2/request"

	"github.com/ONSdigital/log.go/v2/log"

	s3client "github.com/ONSdigital/dp-s3/v2"

	"github.com/ONSdigital/dp-upload-service/upload"
)

const (
	StateUploaded = "UPLOADED"
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
	ErrFilesServer              = errors.New("file api returning internal server errors")
	ErrFilesUnauthorised        = errors.New("access unauthorised")
)

type ContextKey string

type Store struct {
	hostname string
	s3       upload.S3Clienter
	vault    *encryption.Vault
}

func NewStore(hostname string, s3 upload.S3Clienter, vault *encryption.Vault) Store {
	return Store{hostname, s3, vault}
}

func (s Store) UploadFile(ctx context.Context, metadata StoreMetadata, resumable Resumable, content []byte) (bool, error) {
	var encryptionKey []byte
	var err error

	if firstChunk(resumable.CurrentChunk) {
		if err = s.registerFileUpload(ctx, metadata); err != nil {
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
		return true, s.markUploadComplete(ctx, metadata.Path, response.Etag)
	}

	return false, nil
}

func firstChunk(currentChunk int64) bool { return currentChunk == 1 }

func (s Store) generateUploadPart(metadata StoreMetadata, resumable Resumable) *s3client.UploadPartRequest {
	return &s3client.UploadPartRequest{
		UploadKey:   metadata.Path,
		Type:        resumable.Type,
		ChunkNumber: resumable.CurrentChunk,
		TotalChunks: resumable.TotalChunks,
		FileName:    resumable.FileName,
	}
}

func (s Store) markUploadComplete(ctx context.Context, path, etag string) error {
	uc := uploadComplete{StateUploaded, strings.Trim(etag, "\"")}

	req, _ := http.NewRequest(http.MethodPatch, fmt.Sprintf("%s/files/%s", s.hostname, path), jsonEncode(uc))
	req.Header.Set("Content-Type", "application/json")
	s.setAuthHeader(ctx, req)

	resp, err := dphttp.NewClient().Do(ctx, req)

	logData := log.Data{"upload-complete": uc, "request": req, "response": resp}
	if err != nil {
		log.Error(ctx, fmt.Sprintf("making patch request to mark file %s", StateUploaded), err, logData)
		return ErrConnectingToFilesApi
	}

	if resp.StatusCode == http.StatusOK {
		return nil
	}

	switch resp.StatusCode {
	case http.StatusNotFound:
		log.Error(ctx, "could not file file to mark uploaded", ErrFileNotFound, logData)
		return ErrFileNotFound
	case http.StatusConflict:
		log.Error(ctx, "file in wrong state to be marked uploaded", ErrFileStateConflict, logData)
		return ErrFileStateConflict
	case http.StatusInternalServerError:
		err := ErrFilesServer
		log.Error(ctx, "file api returning internal server errors", err, logData)
		return err
	case http.StatusForbidden:
		log.Error(ctx, "unauthorised access", ErrFilesUnauthorised, logData)
		return ErrFilesUnauthorised
	default:
		log.Error(ctx, "unexpected error morning file uploaded", ErrUnknownError, logData)
		return ErrUnknownError
	}
}

func (s Store) registerFileUpload(ctx context.Context, metadata StoreMetadata) error {
	log.Info(ctx, "Register files API Call", log.Data{"hostname": s.hostname})

	req, _ := http.NewRequest(http.MethodPost, fmt.Sprintf("%s/files", s.hostname), jsonEncode(metadata))
	req.Header.Set("Content-Type", "application/json")
	s.setAuthHeader(ctx, req)

	resp, err := dphttp.NewClient().Do(ctx, req)

	if err != nil {
		log.Error(ctx, "failed to connect to files API", err, log.Data{"hostname": s.hostname})
		return ErrConnectingToFilesApi
	}

	if resp.StatusCode == http.StatusCreated {
		return nil
	}

	switch resp.StatusCode {
	case http.StatusInternalServerError:
		return ErrFilesServer
	case http.StatusForbidden:
		return ErrFilesUnauthorised
	case http.StatusBadRequest:
		return s.handleBadRequestResponse(resp)
	default:
		return ErrUnknownError
	}
}

func (s Store) setAuthHeader(ctx context.Context, req *http.Request) {
	if authHeaderValue, ok := ctx.Value(api.AuthContextKey).(string); ok {
		req.Header.Set(request.AuthHeaderKey, authHeaderValue)
	} else {
		log.Info(ctx, fmt.Sprintf("no %s set in context, this may cause auth issues", request.AuthHeaderKey), log.Data{"Request:": req})
	}
}

func (s Store) handleBadRequestResponse(resp *http.Response) error {
	jsonErrors := jsonErrors{}
	if err := json.NewDecoder(resp.Body).Decode(&jsonErrors); err != nil {
		return err
	}

	switch jsonErrors.Error[0].Code {
	case "DuplicateFileError":
		return ErrFilesAPIDuplicateFile
	case "ValidationError":
		return ErrFileAPICreateInvalidData
	default:
		return ErrUnknownError
	}
}

func jsonEncode(data interface{}) *bytes.Buffer {
	b := &bytes.Buffer{}
	json.NewEncoder(b).Encode(data) // nolint // Only fails due to coding error
	return b
}

type StoreMetadata struct {
	Path          string  `json:"path"`
	IsPublishable bool    `json:"is_publishable"`
	CollectionId  *string `json:"collection_id,omitempty"`
	Title         string  `json:"title"`
	SizeInBytes   int     `json:"size_in_bytes"`
	Type          string  `json:"type"`
	Licence       string  `json:"licence"`
	LicenceUrl    string  `json:"licence_url"`
}

type Resumable struct {
	FileName     string `schema:"resumableFilename"`
	Type         string `schema:"resumableType"`
	CurrentChunk int64  `schema:"resumableChunkNumber"`
	TotalChunks  int    `schema:"resumableTotalChunks"`
}

type uploadComplete struct {
	State string `json:"state"`
	ETag  string `json:"etag"`
}

type jsonError struct {
	Code        string `json:"code"`
	Description string `json:"description"`
}

type jsonErrors struct {
	Error []jsonError `json:"errors"`
}
