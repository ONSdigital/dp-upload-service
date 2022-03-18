package api

import (
	"context"
	"fmt"
	"io/ioutil"
	"net/http"
	"regexp"

	"github.com/ONSdigital/dp-upload-service/files"

	"github.com/ONSdigital/log.go/v2/log"
	"github.com/go-playground/validator"
	"github.com/gorilla/schema"
)

const (
	maxChunkSize       = 5 * 1024 * 1024
	maxMultipartMemory = maxChunkSize + 1024
)

type Metadata struct {
	Path          string  `schema:"path" validate:"required,aws-upload-key"`
	IsPublishable *bool   `schema:"isPublishable,omitempty" validate:"required"`
	CollectionId  *string `schema:"collectionId,omitempty"`
	Title         string  `schema:"title"`
	SizeInBytes   int     `schema:"resumableTotalSize" validate:"required"`
	Type          string  `schema:"resumableType" validate:"required"`
	Licence       string  `schema:"licence" validate:"required"`
	LicenceUrl    string  `schema:"licenceUrl" validate:"required"`
}

type StoreFile func(ctx context.Context, uf files.StoreMetadata, r files.Resumable, content []byte) (bool, error)

func CreateV1UploadHandler(storeFile StoreFile) http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {
		if err := req.ParseMultipartForm(maxMultipartMemory); err != nil {
			log.Error(req.Context(), "error parsing form", err)
			writeError(w, buildErrors(err, "ParsingForm"), http.StatusBadRequest)
			return
		}

		d := schema.NewDecoder()
		d.IgnoreUnknownKeys(true)

		metadata := Metadata{}
		if err := d.Decode(&metadata, req.Form); err != nil {
			log.Error(req.Context(), "error decoding metadata form", err)
			writeError(w, buildErrors(err, "DecodingMetadata"), http.StatusBadRequest)
			return
		}

		resumable := files.Resumable{}
		if err := d.Decode(&resumable, req.Form); err != nil {
			log.Error(req.Context(), "error decoding resumable form", err)
			writeError(w, buildErrors(err, "DecodingResumable"), http.StatusBadRequest)
			return
		}

		v := validator.New()
		v.RegisterValidation("aws-upload-key", awsUploadKeyValidator) // nolint // Only fails due to coding error
		if err := v.Struct(metadata); err != nil {
			if validationErrs, ok := err.(validator.ValidationErrors); ok {
				writeError(w, buildValidationErrors(validationErrs), http.StatusBadRequest)
				return
			}
		}

		content, _, err := req.FormFile("file")
		if err != nil {
			log.Error(req.Context(), "error getting file from form", err)
			writeError(w, buildErrors(err, "FileForm"), http.StatusBadRequest)
			return
		}
		defer content.Close()

		payload, err := ioutil.ReadAll(content)
		if err != nil {
			log.Error(req.Context(), "error getting file from form", err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		allPartsUploaded, err := storeFile(req.Context(), getStoreMetadata(metadata, resumable), resumable, payload)
		if err != nil {
			switch err {
			case files.ErrFilesAPIDuplicateFile:
				writeError(w, buildErrors(err, "DuplicateFile"), http.StatusBadRequest)
			case files.ErrFileAPICreateInvalidData:
				writeError(w, buildErrors(err, "RemoteValidationError"), http.StatusInternalServerError)
			case files.ErrChunkTooSmall:
				writeError(w, buildErrors(err, "ChunkTooSmall"), http.StatusBadRequest)
			default:
				writeError(w, buildErrors(err, "InternalError"), http.StatusInternalServerError)
			}
			return
		}

		w.WriteHeader(getResponseStatus(allPartsUploaded))
	}
}

func awsUploadKeyValidator(fl validator.FieldLevel) bool {
	path := fl.Field().String()
	matched, _ := regexp.MatchString("^[a-z0-9A-Z\\/\\!\\*\\_\\'\\(\\)\\.\\-]*$", path)

	return matched
}

func getResponseStatus(allPartsUploaded bool) int {
	if allPartsUploaded {
		return http.StatusCreated
	}

	return http.StatusOK
}

func getStoreMetadata(metadata Metadata, resumable files.Resumable) files.StoreMetadata {
	return files.StoreMetadata{
		Path:          fmt.Sprintf("%s/%s", metadata.Path, resumable.FileName),
		IsPublishable: *metadata.IsPublishable,
		CollectionId:  metadata.CollectionId,
		Title:         metadata.Title,
		SizeInBytes:   metadata.SizeInBytes,
		Type:          metadata.Type,
		Licence:       metadata.Licence,
		LicenceUrl:    metadata.LicenceUrl,
	}
}
