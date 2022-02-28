package api

import (
	"context"
	"fmt"
	"github.com/ONSdigital/dp-upload-service/files"
	"io/ioutil"
	"net/http"
	"regexp"

	"github.com/ONSdigital/log.go/v2/log"
	"github.com/gabriel-vasile/mimetype"
	"github.com/go-playground/validator"
	"github.com/gorilla/schema"
)

type Metadata struct {
	Path    string `schema:"path" validate:"required,aws-upload-key"`
	IsPublishable bool   `schema:"isPublishable" validate:"required"`
	CollectionId  string `schema:"collectionId" validate:"required"`
	Title         string `schema:"title"`
	SizeInBytes   int    `schema:"resumableTotalSize" validate:"required"`
	Type          string `schema:"resumableType" validate:"required,mime-type"`
	Licence       string `schema:"licence" validate:"required"`
	LicenceUrl    string `schema:"licenceUrl" validate:"required"`
}

type StoreFile func(ctx context.Context, uf files.StoreMetadata, r files.Resumable, content []byte) (bool, error)

func mimeValidator(fl validator.FieldLevel) bool {
	mt := fl.Field().String()
	mtype := mimetype.Lookup(mt)

	return mtype != nil
}

func awsUploadKeyValidator(fl validator.FieldLevel) bool {
	path := fl.Field().String()
	matched, _ := regexp.MatchString("^[a-z\\.\\-0-9]{3,63}$", path)

	return matched
}

func CreateV1UploadHandler(storeFile StoreFile) http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {
		if err := req.ParseMultipartForm(4); err != nil {
			log.Error(req.Context(), "error parsing form", err)
			writeError(w, buildErrors(err, "error parsing form"), http.StatusBadRequest)
			return
		}

		metadata := Metadata{}

		d := schema.NewDecoder()
		d.IgnoreUnknownKeys(true)
		if err := d.Decode(&metadata, req.Form); err != nil {
			log.Error(req.Context(), "error decoding metadata form", err)
			writeError(w, buildErrors(err, "error decoding metadata form"), http.StatusBadRequest)
			return
		}

		resumable := files.Resumable{}
		if err := d.Decode(&resumable, req.Form); err != nil {
			log.Error(req.Context(), "error decoding resumable form", err)
			writeError(w, buildErrors(err, "error decoding resumable form"), http.StatusBadRequest)
			return
		}

		v := validator.New()
		v.RegisterValidation("mime-type", mimeValidator)              // nolint // Only fails due to coding error
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
			writeError(w, buildErrors(err, "error getting file from form"), http.StatusBadRequest)
			return
		}
		defer content.Close()

		payload, err := ioutil.ReadAll(content)
		if err != nil {
			log.Error(req.Context(), "error getting file from form", err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		storeMetadata := files.StoreMetadata{
			Path:          fmt.Sprintf("%s/%s", resumable.Path, resumable.FileName),
			IsPublishable: metadata.IsPublishable,
			CollectionId:  metadata.CollectionId,
			Title:         metadata.Title,
			SizeInBytes:   metadata.SizeInBytes,
			Type:          metadata.Type,
			Licence:       metadata.Licence,
			LicenceUrl:    metadata.LicenceUrl,
		}

		allPartsUploaded, err := storeFile(req.Context(), storeMetadata, resumable, payload)
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
		}

		if !allPartsUploaded {
			w.WriteHeader(http.StatusContinue)
		}
	}
}

func buildValidationErrors(validationErrs validator.ValidationErrors) jsonErrors {
	jsonErrs := jsonErrors{Error: []jsonError{}}

	for _, validationErr := range validationErrs {
		desc := fmt.Sprintf("%s %s", validationErr.Field(), validationErr.Tag())
		jsonErrs.Error = append(jsonErrs.Error, jsonError{Code: "ValidationError", Description: desc})
	}
	return jsonErrs
}
