package api_test

import (
	"bytes"
	"context"
	"errors"
	"io/ioutil"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/ONSdigital/dp-upload-service/files"

	"github.com/ONSdigital/dp-upload-service/api"
	"github.com/stretchr/testify/assert"
)

func TestJsonProvidedRatherThanMultiPartFrom(t *testing.T) {
	buf := bytes.NewBufferString(`{"key": "value"}`)

	rec := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodPost, "/v1/upload", buf)

	h := api.CreateV1UploadHandler(func(ctx context.Context, uf files.Metadata, fileContent []byte) error { return nil })

	h.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusBadRequest, rec.Code)
	response, _ := ioutil.ReadAll(rec.Body)
	assert.Contains(t, string(response), "error parsing form")
}

type ErrorWriter struct {
	status int
}

func (e *ErrorWriter) Header() http.Header {
	return http.Header{}
}

func (e *ErrorWriter) Write(i []byte) (int, error) {
	return 0, errors.New("broken")
}

func (e *ErrorWriter) WriteHeader(statusCode int) {
	e.status = statusCode
}

func TestFailureToWriteErrorToResponse(t *testing.T) {
	buf := bytes.NewBufferString(`{"key": "value"}`)

	rec := &ErrorWriter{}
	req, _ := http.NewRequest(http.MethodPost, "/v1/upload", buf)

	h := api.CreateV1UploadHandler(func(ctx context.Context, uf files.Metadata, fileContent []byte) error { return nil })

	h.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusInternalServerError, rec.status)
}

func TestBadMultiPartForm(t *testing.T) {
	b := &bytes.Buffer{}
	formWriter := multipart.NewWriter(b)
	formWriter.WriteField("key", "value")
	formWriter.Close()

	rec := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodPost, "/v1/upload", b)
	req.Header.Set("Content-Type", formWriter.FormDataContentType())

	h := api.CreateV1UploadHandler(func(ctx context.Context, uf files.Metadata, fileContent []byte) error { return nil })

	h.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusBadRequest, rec.Code)
	response, _ := ioutil.ReadAll(rec.Body)
	assert.Contains(t, string(response), "error decoding form")
}

func TestRequiredFields(t *testing.T) {
	b := &bytes.Buffer{}
	formWriter := multipart.NewWriter(b)

	formWriter.Close()

	rec := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodPost, "/v1/upload", b)
	req.Header.Set("Content-Type", formWriter.FormDataContentType())

	h := api.CreateV1UploadHandler(func(ctx context.Context, uf files.Metadata, fileContent []byte) error { return nil })

	h.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusBadRequest, rec.Code)
	response, _ := ioutil.ReadAll(rec.Body)

	assert.Contains(t, string(response), "Path required")
	assert.Contains(t, string(response), "IsPublishable required")
	assert.Contains(t, string(response), "CollectionId required")
	assert.Contains(t, string(response), "SizeInBytes required")
	assert.Contains(t, string(response), "Type required")
	assert.Contains(t, string(response), "Licence required")
	assert.Contains(t, string(response), "LicenceUrl required")
}

func TestPathValid(t *testing.T) {
	b := &bytes.Buffer{}
	formWriter := multipart.NewWriter(b)
	formWriter.WriteField("path", "bad_uri")
	formWriter.WriteField("isPublishable", "true")
	formWriter.WriteField("collectionId", "1234567890")
	formWriter.WriteField("title", "A New File")
	formWriter.WriteField("sizeInBytes", "1478")
	formWriter.WriteField("type", "text/csv")
	formWriter.WriteField("licence", "OGL v3")
	formWriter.WriteField("licenceUrl", "http://www.nationalarchives.gov.uk/doc/open-government-licence/version/3/")
	formWriter.Close()

	rec := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodPost, "/v1/upload", b)
	req.Header.Set("Content-Type", formWriter.FormDataContentType())

	h := api.CreateV1UploadHandler(func(ctx context.Context, uf files.Metadata, fileContent []byte) error { return nil })

	h.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusBadRequest, rec.Code)
	response, _ := ioutil.ReadAll(rec.Body)
	assert.Contains(t, string(response), "Path uri")
}

func TestTypeValid(t *testing.T) {
	b := &bytes.Buffer{}
	formWriter := multipart.NewWriter(b)
	formWriter.WriteField("path", "/valid/path.csv")
	formWriter.WriteField("isPublishable", "true")
	formWriter.WriteField("collectionId", "1234567890")
	formWriter.WriteField("title", "A New File")
	formWriter.WriteField("sizeInBytes", "1478")
	formWriter.WriteField("type", "INVALID")
	formWriter.WriteField("licence", "OGL v3")
	formWriter.WriteField("licenceUrl", "http://www.nationalarchives.gov.uk/doc/open-government-licence/version/3/")
	formWriter.Close()

	rec := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodPost, "/v1/upload", b)
	req.Header.Set("Content-Type", formWriter.FormDataContentType())

	h := api.CreateV1UploadHandler(func(ctx context.Context, uf files.Metadata, fileContent []byte) error { return nil })

	h.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusBadRequest, rec.Code)
	response, _ := ioutil.ReadAll(rec.Body)
	assert.Contains(t, string(response), "Type mime-type")
}

func TestFileWasSupplied(t *testing.T) {
	b := &bytes.Buffer{}
	formWriter := multipart.NewWriter(b)
	formWriter.WriteField("path", "/some/path.csv")
	formWriter.WriteField("isPublishable", "true")
	formWriter.WriteField("collectionId", "1234567890")
	formWriter.WriteField("title", "A New File")
	formWriter.WriteField("sizeInBytes", "1478")
	formWriter.WriteField("type", "text/csv")
	formWriter.WriteField("licence", "OGL v3")
	formWriter.WriteField("licenceUrl", "http://www.nationalarchives.gov.uk/doc/open-government-licence/version/3/")
	formWriter.Close()

	h := api.CreateV1UploadHandler(func(ctx context.Context, uf files.Metadata, fileContent []byte) error { return nil })

	rec := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodPost, "/v1/upload", b)
	req.Header.Set("Content-Type", formWriter.FormDataContentType())

	h.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusBadRequest, rec.Code)
	response, _ := ioutil.ReadAll(rec.Body)
	assert.Contains(t, string(response), "error getting file from form")
}

func TestSuccessfulStorageOfFileReturns200(t *testing.T) {
	payload := "TEST DATA"
	funcCalled := false
	st := func(ctx context.Context, uf files.Metadata, fileContent []byte) error {
		funcCalled = true
		assert.Equal(t, payload, string(fileContent))
		return nil
	}

	b := &bytes.Buffer{}
	formWriter := multipart.NewWriter(b)
	formWriter.WriteField("path", "/some/path.csv")
	formWriter.WriteField("isPublishable", "true")
	formWriter.WriteField("collectionId", "1234567890")
	formWriter.WriteField("title", "A New File")
	formWriter.WriteField("sizeInBytes", "1478")
	formWriter.WriteField("type", "text/csv")
	formWriter.WriteField("licence", "OGL v3")
	formWriter.WriteField("licenceUrl", "http://www.nationalarchives.gov.uk/doc/open-government-licence/version/3/")
	part, _ := formWriter.CreateFormFile("file", "testing.csv")
	part.Write([]byte(payload))
	formWriter.Close()

	h := api.CreateV1UploadHandler(st)

	rec := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodPost, "/v1/upload", b)
	req.Header.Set("Content-Type", formWriter.FormDataContentType())

	h.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)
	assert.True(t, funcCalled)
}

func TestFilePathExistsInFilesAPIReturns409(t *testing.T) {
	st := func(ctx context.Context, uf files.Metadata, fileContent []byte) error {
		return files.ErrFilesAPIDuplicateFile
	}

	b := &bytes.Buffer{}
	formWriter := multipart.NewWriter(b)
	formWriter.WriteField("path", "/some/path.csv")
	formWriter.WriteField("isPublishable", "true")
	formWriter.WriteField("collectionId", "1234567890")
	formWriter.WriteField("title", "A New File")
	formWriter.WriteField("sizeInBytes", "1478")
	formWriter.WriteField("type", "text/csv")
	formWriter.WriteField("licence", "OGL v3")
	formWriter.WriteField("licenceUrl", "http://www.nationalarchives.gov.uk/doc/open-government-licence/version/3/")
	part, _ := formWriter.CreateFormFile("file", "testing.csv")
	part.Write([]byte("TEST DATA"))
	formWriter.Close()

	h := api.CreateV1UploadHandler(st)

	rec := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodPost, "/v1/upload", b)
	req.Header.Set("Content-Type", formWriter.FormDataContentType())

	h.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusBadRequest, rec.Code)
	response, _ := ioutil.ReadAll(rec.Body)
	assert.Contains(t, string(response), "DuplicateFile")
}

func TestInvalidContentReturns500(t *testing.T) {
	st := func(ctx context.Context, uf files.Metadata, fileContent []byte) error {
		return files.ErrFileAPICreateInvalidData
	}

	b := &bytes.Buffer{}
	formWriter := multipart.NewWriter(b)
	formWriter.WriteField("path", "/some/path.csv")
	formWriter.WriteField("isPublishable", "true")
	formWriter.WriteField("collectionId", "1234567890")
	formWriter.WriteField("title", "A New File")
	formWriter.WriteField("sizeInBytes", "1478")
	formWriter.WriteField("type", "text/csv")
	formWriter.WriteField("licence", "OGL v3")
	formWriter.WriteField("licenceUrl", "http://www.nationalarchives.gov.uk/doc/open-government-licence/version/3/")
	part, _ := formWriter.CreateFormFile("file", "testing.csv")
	part.Write([]byte("TEST DATA"))
	formWriter.Close()

	h := api.CreateV1UploadHandler(st)

	rec := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodPost, "/v1/upload", b)
	req.Header.Set("Content-Type", formWriter.FormDataContentType())

	h.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusInternalServerError, rec.Code)
	response, _ := ioutil.ReadAll(rec.Body)
	assert.Contains(t, string(response), "RemoteValidationError")
}

func TestUnexpectedErrorReturns500(t *testing.T) {
	st := func(ctx context.Context, uf files.Metadata, fileContent []byte) error {
		return errors.New("its broken")
	}

	b := &bytes.Buffer{}
	formWriter := multipart.NewWriter(b)
	formWriter.WriteField("path", "/some/path.csv")
	formWriter.WriteField("isPublishable", "true")
	formWriter.WriteField("collectionId", "1234567890")
	formWriter.WriteField("title", "A New File")
	formWriter.WriteField("sizeInBytes", "1478")
	formWriter.WriteField("type", "text/csv")
	formWriter.WriteField("licence", "OGL v3")
	formWriter.WriteField("licenceUrl", "http://www.nationalarchives.gov.uk/doc/open-government-licence/version/3/")
	part, _ := formWriter.CreateFormFile("file", "testing.csv")
	part.Write([]byte("TEST DATA"))
	formWriter.Close()

	h := api.CreateV1UploadHandler(st)

	rec := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodPost, "/v1/upload", b)
	req.Header.Set("Content-Type", formWriter.FormDataContentType())

	h.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusInternalServerError, rec.Code)
	response, _ := ioutil.ReadAll(rec.Body)
	assert.Contains(t, string(response), "InternalError")
}
