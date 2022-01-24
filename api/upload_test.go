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

var stubStoreFunction = func(ctx context.Context, uf files.Metadata, r files.Resumable, c []byte) (bool, error) {
	return false, nil
}

func TestJsonProvidedRatherThanMultiPartFrom(t *testing.T) {
	buf := bytes.NewBufferString(`{"key": "value"}`)

	rec := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodPost, "/v1/upload", buf)

	h := api.CreateV1UploadHandler(stubStoreFunction)

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

	h := api.CreateV1UploadHandler(stubStoreFunction)

	h.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusInternalServerError, rec.status)
}

func TestRequiredFields(t *testing.T) {
	b := &bytes.Buffer{}
	formWriter := multipart.NewWriter(b)

	formWriter.Close()

	rec := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodPost, "/v1/upload", b)
	req.Header.Set("Content-Type", formWriter.FormDataContentType())

	h := api.CreateV1UploadHandler(stubStoreFunction)

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
	formWriter.WriteField("resumableFilename", "/invalid/upload-key/file.csv")
	formWriter.WriteField("isPublishable", "true")
	formWriter.WriteField("collectionId", "1234567890")
	formWriter.WriteField("title", "A New File")
	formWriter.WriteField("resumableTotalSize", "1478")
	formWriter.WriteField("resumableType", "text/csv")
	formWriter.WriteField("licence", "OGL v3")
	formWriter.WriteField("licenceUrl", "http://www.nationalarchives.gov.uk/doc/open-government-licence/version/3/")
	formWriter.Close()

	rec := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodPost, "/v1/upload", b)
	req.Header.Set("Content-Type", formWriter.FormDataContentType())

	h := api.CreateV1UploadHandler(stubStoreFunction)

	h.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusBadRequest, rec.Code)
	response, _ := ioutil.ReadAll(rec.Body)
	assert.Contains(t, string(response), "Path aws-upload-key")
}

func TestTypeValid(t *testing.T) {
	b := &bytes.Buffer{}
	formWriter := multipart.NewWriter(b)
	formWriter.WriteField("resumableFilename", "valid/path.csv")
	formWriter.WriteField("isPublishable", "true")
	formWriter.WriteField("collectionId", "1234567890")
	formWriter.WriteField("title", "A New File")
	formWriter.WriteField("resumableTotalSize", "1478")
	formWriter.WriteField("resumableType", "INVALID")
	formWriter.WriteField("licence", "OGL v3")
	formWriter.WriteField("licenceUrl", "http://www.nationalarchives.gov.uk/doc/open-government-licence/version/3/")
	formWriter.Close()

	rec := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodPost, "/v1/upload", b)
	req.Header.Set("Content-Type", formWriter.FormDataContentType())

	h := api.CreateV1UploadHandler(stubStoreFunction)

	h.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusBadRequest, rec.Code)
	response, _ := ioutil.ReadAll(rec.Body)
	assert.Contains(t, string(response), "Type mime-type")
}

func TestFileWasSupplied(t *testing.T) {
	b := &bytes.Buffer{}
	formWriter := multipart.NewWriter(b)
	formWriter.WriteField("resumableFilename", "valid/path.csv")
	formWriter.WriteField("isPublishable", "true")
	formWriter.WriteField("collectionId", "1234567890")
	formWriter.WriteField("title", "A New File")
	formWriter.WriteField("resumableTotalSize", "1478")
	formWriter.WriteField("resumableType", "text/csv")
	formWriter.WriteField("licence", "OGL v3")
	formWriter.WriteField("licenceUrl", "http://www.nationalarchives.gov.uk/doc/open-government-licence/version/3/")
	formWriter.Close()

	h := api.CreateV1UploadHandler(stubStoreFunction)

	rec := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodPost, "/v1/upload", b)
	req.Header.Set("Content-Type", formWriter.FormDataContentType())

	h.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusBadRequest, rec.Code)
	response, _ := ioutil.ReadAll(rec.Body)
	assert.Contains(t, string(response), "error getting file from form")
}

func TestSuccessfulStorageOfCompleteFileReturns200(t *testing.T) {
	payload := "TEST DATA"
	funcCalled := false
	st := func(ctx context.Context, uf files.Metadata, r files.Resumable, fileContent []byte) (bool, error) {
		funcCalled = true
		assert.Equal(t, payload, string(fileContent))
		return true, nil
	}

	b := &bytes.Buffer{}
	formWriter := multipart.NewWriter(b)
	formWriter.WriteField("resumableFilename", "valid/path.csv")
	formWriter.WriteField("isPublishable", "true")
	formWriter.WriteField("collectionId", "1234567890")
	formWriter.WriteField("title", "A New File")
	formWriter.WriteField("resumableTotalSize", "1478")
	formWriter.WriteField("resumableType", "text/csv")
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

func TestChunkTooSmallReturns400(t *testing.T) {
	payload := "TEST DATA"
	st := func(ctx context.Context, uf files.Metadata, r files.Resumable, fileContent []byte) (bool, error) {
		return true, files.ErrChunkTooSmall
	}

	b := &bytes.Buffer{}
	formWriter := multipart.NewWriter(b)
	formWriter.WriteField("resumableFilename", "valid/path.csv")
	formWriter.WriteField("isPublishable", "true")
	formWriter.WriteField("collectionId", "1234567890")
	formWriter.WriteField("title", "A New File")
	formWriter.WriteField("resumableTotalSize", "1478")
	formWriter.WriteField("resumableType", "text/csv")
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

	assert.Equal(t, http.StatusBadRequest, rec.Code)
}

func TestFilePathExistsInFilesAPIReturns409(t *testing.T) {
	st := func(ctx context.Context, uf files.Metadata, r files.Resumable, fileContent []byte) (bool, error) {
		return false, files.ErrFilesAPIDuplicateFile
	}

	b := &bytes.Buffer{}
	formWriter := multipart.NewWriter(b)
	formWriter.WriteField("resumableFilename", "valid/path.csv")
	formWriter.WriteField("isPublishable", "true")
	formWriter.WriteField("collectionId", "1234567890")
	formWriter.WriteField("title", "A New File")
	formWriter.WriteField("resumableTotalSize", "1478")
	formWriter.WriteField("resumableType", "text/csv")
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
	st := func(ctx context.Context, uf files.Metadata, r files.Resumable, fileContent []byte) (bool, error) {
		return false, files.ErrFileAPICreateInvalidData
	}

	b := &bytes.Buffer{}
	formWriter := multipart.NewWriter(b)
	formWriter.WriteField("resumableFilename", "valid/path.csv")
	formWriter.WriteField("isPublishable", "true")
	formWriter.WriteField("collectionId", "1234567890")
	formWriter.WriteField("title", "A New File")
	formWriter.WriteField("resumableTotalSize", "1478")
	formWriter.WriteField("resumableType", "text/csv")
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
	st := func(ctx context.Context, uf files.Metadata, r files.Resumable, fileContent []byte) (bool, error) {
		return false, errors.New("its broken")
	}

	b := &bytes.Buffer{}
	formWriter := multipart.NewWriter(b)
	formWriter.WriteField("resumableFilename", "valid/path.csv")
	formWriter.WriteField("isPublishable", "true")
	formWriter.WriteField("collectionId", "1234567890")
	formWriter.WriteField("title", "A New File")
	formWriter.WriteField("resumableTotalSize", "1478")
	formWriter.WriteField("resumableType", "text/csv")
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
