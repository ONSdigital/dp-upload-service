package api

import (
	"context"
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/ONSdigital/dp-upload-service/files"

	"github.com/ONSdigital/log.go/v2/log"
	"github.com/gabriel-vasile/mimetype"
	"github.com/go-playground/validator"
	"github.com/gorilla/schema"
)

func mimeValidator(fl validator.FieldLevel) bool {
	mt := fl.Field().String()
	mtype := mimetype.Lookup(mt)

	return mtype != nil
}

type StoreFile func(ctx context.Context, uf files.Metadata, content []byte) error

func CreateV1UploadHandler(storeFile StoreFile) http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {
		if err := req.ParseMultipartForm(4); err != nil {
			log.Error(req.Context(), "error parsing form", err)
			writeError(w, buildErrors(err, "error parsing form"), http.StatusBadRequest)
			return
		}

		metadata := files.Metadata{}

		if err := schema.NewDecoder().Decode(&metadata, req.Form); err != nil {
			log.Error(req.Context(), "error decoding form", err)
			writeError(w, buildErrors(err, "error decoding form"), http.StatusBadRequest)
			return
		}

		v := validator.New()
		v.RegisterValidation("mime-type", mimeValidator) // nolint // Only fails due to coding error
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

		err = storeFile(req.Context(), metadata, payload)
		if err != nil {
			switch err {
			case files.ErrFilesAPIDuplicateFile:
				writeError(w, buildErrors(err, "DuplicateFile"), http.StatusBadRequest)
			case files.ErrFileAPICreateInvalidData:
				writeError(w, buildErrors(err, "RemoteValidationError"), http.StatusInternalServerError)
			default:
				writeError(w, buildErrors(err, "InternalError"), http.StatusInternalServerError)
			}
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
