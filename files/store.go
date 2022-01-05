package files

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"

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
)

type Store struct {
	hostname string
	s3       upload.S3Clienter
}

func NewStore(hostname string, s3 upload.S3Clienter) Store {
	return Store{hostname, s3}
}

type Metadata struct {
	Path          string `schema:"path" json:"path" validate:"required,uri"`
	IsPublishable bool   `schema:"isPublishable" json:"is_publishable" validate:"required"`
	CollectionId  string `schema:"collectionId" json:"collection_id" validate:"required"`
	Title         string `schema:"title" json:"title"`
	SizeInBytes   int    `schema:"sizeInBytes" json:"size_in_bytes" validate:"required"`
	Type          string `schema:"type" json:"type" validate:"required,mime-type"`
	Licence       string `schema:"licence" json:"licence" validate:"required"`
	LicenceUrl    string `schema:"licenceUrl" json:"licence_url" validate:"required"`
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

func (s Store) UploadFile(ctx context.Context, metadata Metadata, content []byte) error {
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

	upr := s3client.UploadPartRequest{
		UploadKey:   metadata.Path,
		Type:        metadata.Type,
		ChunkNumber: 1,
		TotalChunks: 1,
		FileName:    metadata.Path,
	}

	response, err := s.s3.UploadPart(ctx, &upr, content)
	if err != nil {
		return ErrS3Upload
	}

	uc := uploadComplete{
		Path: metadata.Path,
		ETag: strings.Trim(response.Etag, "\""),
	}

	resp, _ = http.Post(fmt.Sprintf("%s/v1/files/upload-complete", s.hostname), "application/json", jsonEncode(uc))
	if resp.StatusCode == http.StatusNotFound {
		return ErrFileNotFound
	} else if resp.StatusCode == http.StatusConflict {
		return ErrFileStateConflict
	} else if resp.StatusCode != http.StatusCreated {
		return ErrUnknownError
	}

	return nil
}

func jsonEncode(data interface{}) *bytes.Buffer {
	b := &bytes.Buffer{}
	json.NewEncoder(b).Encode(data) // nolint // Only fails due to coding error
	return b
}
