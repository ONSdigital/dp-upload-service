package upload

import (
	"context"
	"encoding/json"
	"io"
	"net/http"

	s3client "github.com/ONSdigital/dp-s3/v2"
	"github.com/ONSdigital/dp-upload-service/aws"
	"github.com/ONSdigital/log.go/v2/log"
	"github.com/gorilla/mux"
	"github.com/gorilla/schema"
)

var decoder = schema.NewDecoder()

// Resumable represents resumable js upload query pararmeters
type Resumable struct {
	ChunkNumber      int    `schema:"resumableChunkNumber" validate:"required"`
	TotalChunks      int    `schema:"resumableTotalChunks" validate:"required"`
	ChunkSize        int    `schema:"resumableChunkSize" validate:"required"`
	CurrentChunkSize int    `schema:"resumableCurrentChunkSize" validate:"required"`
	TotalSize        int    `schema:"resumableTotalSize" validate:"required"`
	Type             string `schema:"resumableType" validate:"required"`
	Identifier       string `schema:"resumableIdentifier" validate:"required"`
	FileName         string `schema:"resumableFilename" validate:"required"`
	RelativePath     string `schema:"resumableRelativePath" validate:"required"`
	AliasName        string `schema:"aliasName" validate:"required"`
}

// createS3Request creates a S3 UploadRequest struct from a Resumable struct
func (resum *Resumable) createS3Request() *s3client.UploadPartRequest {
	log.Info(context.Background(), "calling function s3 request", log.Data{"resumeidentifier": resum.Identifier})
	return &s3client.UploadPartRequest{
		UploadKey:   resum.Identifier,
		Type:        resum.Type,
		ChunkNumber: int64(resum.ChunkNumber),
		TotalChunks: resum.TotalChunks,
		FileName:    resum.FileName,
	}
}

// Uploader represents the necessary configuration for uploading a file
type Uploader struct {
	bucket *aws.Bucket
}

// New returns a new Uploader from the provided clients
func New(bucket *aws.Bucket) *Uploader {
	return &Uploader{
		bucket: bucket,
	}
}

// CheckUploaded godoc
// @Description  checks to see if a chunk has been uploaded
// @Tags         upload
// @Accept       json
// @Produce      json
// @Param        request formData Resumable false "Resumable"
// @Success      200
// @Failure      400
// @Failure      404
// @Failure      500
// @Router       /upload [get]
func (u *Uploader) CheckUploaded(w http.ResponseWriter, req *http.Request) {

	if err := req.ParseForm(); err != nil {
		log.Error(req.Context(), "error parsing form", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	resum := new(Resumable)

	if err := decoder.Decode(resum, req.Form); err != nil {
		log.Error(req.Context(), "error decoding form", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	_, err := u.bucket.CheckPartUploaded(req.Context(), resum.createS3Request())
	if err != nil {
		log.Error(req.Context(), "error returned from check part uploaded", err)
		w.WriteHeader(statusCodeFromS3Error(err))
		return
	}

	log.Info(req.Context(), "uploaded file successfully", log.Data{"file-name": resum.FileName, "uid": resum.Identifier, "size": resum.TotalSize})
	w.WriteHeader(http.StatusOK)

}

// Upload godoc
// @Description  handles the uploading of a file to AWS s3
// @Tags         upload
// @Accept       json
// @Produce      json
// @Param        request formData Resumable false "Resumable"
// @Success      200
// @Failure      400
// @Failure      404
// @Failure      500
// @Router       /upload [post]
func (u *Uploader) Upload(w http.ResponseWriter, req *http.Request) {
	if err := req.ParseForm(); err != nil {
		log.Error(req.Context(), "error parsing form", err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	resum := new(Resumable)

	if err := decoder.Decode(resum, req.Form); err != nil {
		log.Warn(req.Context(), "error decoding form", log.FormatErrors([]error{err}))
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	content, _, err := req.FormFile("file")
	if err != nil {
		log.Error(req.Context(), "error getting file from form", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	defer content.Close()

	payload, err := io.ReadAll(content)
	if err != nil {
		log.Error(req.Context(), "error reading file", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	// Perform upload
	if _, err := u.bucket.UploadPart(req.Context(), resum.createS3Request(), payload); err != nil {
		log.Error(req.Context(), "error returned from upload", err)
		w.WriteHeader(statusCodeFromS3Error(err))
		return
	}
	w.WriteHeader(http.StatusOK)
}

// GetS3URL godoc
// @Description  returns an S3 URL for a requested path, and the client's region and bucket name
// @Tags         upload
// @Accept       json
// @Produce      json
// @Param        id   path      string  true  "S3 object key"
// @Success      200
// @Failure      400
// @Failure      404
// @Failure      500
// @Router       /upload/{id} [get]
func (u *Uploader) GetS3URL(w http.ResponseWriter, req *http.Request) {
	param := req.URL.Query().Get(":id")
	path := mux.Vars(req)["id"]

	// Florence historically sent the query param, but this is being removed. Where
	// it is provided, it should be used by preference for now.
	if len(param) > 0 {
		path = param
	}

	url, err := u.bucket.GetS3URL(path)
	if err != nil {
		log.Error(req.Context(), "error getting S3 URL", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	body := struct {
		URL string `json:"url"`
	}{
		URL: url,
	}

	b, err := json.Marshal(body)
	if err != nil {
		log.Error(req.Context(), "error marshalling json", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write(b) //nolint
}

// handleError decides the HTTP status according to the provided error
func statusCodeFromS3Error(err error) int {
	//nolint
	switch err.(type) {
	case *s3client.ErrNotUploaded:
		return http.StatusNotFound
	case *s3client.ErrListParts:
		return http.StatusNotFound
	case *s3client.ErrChunkNumberNotFound:
		return http.StatusNotFound
	default:
		return http.StatusInternalServerError
	}
	// TODO I would suggest considering S3 client errors to be '502 BAD gateway'
}
