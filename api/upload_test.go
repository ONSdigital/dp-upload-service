package api_test

import (
	"bytes"
	"context"
	"errors"
	"github.com/stretchr/testify/suite"
	"io/ioutil"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/ONSdigital/dp-upload-service/files"

	"github.com/ONSdigital/dp-upload-service/api"
)

var stubStoreFunction = func(ctx context.Context, uf files.StoreMetadata, r files.Resumable, c []byte) (bool, error) {
	return false, nil
}

const UploadURI = "/upload-new"

var rec *httptest.ResponseRecorder

type UploadTestSuite struct {
	suite.Suite
}

func TestUploadTestSuite(t *testing.T) {
	suite.Run(t, new(UploadTestSuite))
}

func (suite *UploadTestSuite) SetupTest() {
	rec = httptest.NewRecorder()
}

func (s UploadTestSuite) TestJsonProvidedRatherThanMultiPartForm() {
	req, _ := http.NewRequest(http.MethodPost, UploadURI, bytes.NewBufferString(`{"key": "value"}`))

	h := api.CreateV1UploadHandler(stubStoreFunction)

	h.ServeHTTP(rec, req)
	s.Equal(http.StatusBadRequest, rec.Code)
	response, _ := ioutil.ReadAll(rec.Body)
	s.Contains(string(response), "ParsingForm")
}

func (s UploadTestSuite) TestFailureToWriteErrorToResponse() {
	rec := &ErrorWriter{}
	req, _ := http.NewRequest(http.MethodPost, UploadURI, bytes.NewBufferString(`{"key": "value"}`))

	h := api.CreateV1UploadHandler(stubStoreFunction)

	h.ServeHTTP(rec, req)

	s.Equal(http.StatusInternalServerError, rec.status)
}

func (s UploadTestSuite) TestRequiredFields() {
	b := &bytes.Buffer{}
	formWriter := multipart.NewWriter(b)
	formWriter.Close()

	h := api.CreateV1UploadHandler(stubStoreFunction)
	h.ServeHTTP(rec, generateRequest(b, formWriter))

	s.Equal(http.StatusBadRequest, rec.Code)
	response, _ := ioutil.ReadAll(rec.Body)

	s.Contains(string(response), "Path required")
	s.Contains(string(response), "IsPublishable required")
	s.Contains(string(response), "SizeInBytes required")
	s.Contains(string(response), "SizeInBytes required")
	s.Contains(string(response), "Type required")
	s.Contains(string(response), "Licence required")
	s.Contains(string(response), "LicenceUrl required")
}

func (s UploadTestSuite) TestPathValid() {
	b, formWriter := generateFormWriter("\\x")
	formWriter.Close()

	h := api.CreateV1UploadHandler(stubStoreFunction)

	h.ServeHTTP(rec, generateRequest(b, formWriter))

	s.Equal(http.StatusBadRequest, rec.Code)
	response, _ := ioutil.ReadAll(rec.Body)
	s.Contains(string(response), "Path aws-upload-key")
}

func (s UploadTestSuite) TestIsPublishableSetToFalseInNotARequireFailure() {
	b, formWriter := generateFormWriter("valid")
	formWriter.Close()

	h := api.CreateV1UploadHandler(stubStoreFunction)
	h.ServeHTTP(rec, generateRequest(b, formWriter))

	s.Equal(http.StatusBadRequest, rec.Code)
	response, _ := ioutil.ReadAll(rec.Body)
	s.NotContains(string(response), "IsPublishable required")
}

func (s UploadTestSuite) TestFileWasSupplied() {
	b, formWriter := generateFormWriter("valid")
	formWriter.Close()

	h := api.CreateV1UploadHandler(stubStoreFunction)

	h.ServeHTTP(rec, generateRequest(b, formWriter))

	s.Equal(http.StatusBadRequest, rec.Code)
	response, _ := ioutil.ReadAll(rec.Body)
	s.Contains(string(response), "FileForm")
}

func (s UploadTestSuite) TestSuccessfulStorageOfCompleteFileReturns201() {
	payload := "TEST DATA"
	funcCalled := false
	st := func(ctx context.Context, uf files.StoreMetadata, r files.Resumable, fileContent []byte) (bool, error) {
		funcCalled = true
		s.Equal(payload, string(fileContent))
		return true, nil
	}

	b, formWriter := generateFormWriter("valid")
	part, _ := formWriter.CreateFormFile("file", "testing.csv")
	part.Write([]byte(payload))
	formWriter.Close()

	h := api.CreateV1UploadHandler(st)

	h.ServeHTTP(rec, generateRequest(b, formWriter))

	s.Equal(http.StatusCreated, rec.Code)
	s.True(funcCalled)
}

func (s UploadTestSuite) TestChunkTooSmallReturns400() {
	payload := "TEST DATA"
	st := func(ctx context.Context, uf files.StoreMetadata, r files.Resumable, fileContent []byte) (bool, error) {
		return true, files.ErrChunkTooSmall
	}

	b, formWriter := generateFormWriter("valid")
	part, _ := formWriter.CreateFormFile("file", "testing.csv")
	part.Write([]byte(payload))
	formWriter.Close()

	h := api.CreateV1UploadHandler(st)

	req, _ := http.NewRequest(http.MethodPost, UploadURI, b)
	req.Header.Set("Content-Type", formWriter.FormDataContentType())

	h.ServeHTTP(rec, req)

	s.Equal(http.StatusBadRequest, rec.Code)
}

func (s UploadTestSuite) TestFilePathExistsInFilesAPIReturns409() {
	st := func(ctx context.Context, uf files.StoreMetadata, r files.Resumable, fileContent []byte) (bool, error) {
		return false, files.ErrFilesAPIDuplicateFile
	}

	b, formWriter := generateFormWriter("valid")
	part, _ := formWriter.CreateFormFile("file", "testing.csv")
	part.Write([]byte("TEST DATA"))
	formWriter.Close()

	h := api.CreateV1UploadHandler(st)
	h.ServeHTTP(rec, generateRequest(b, formWriter))

	s.Equal(http.StatusBadRequest, rec.Code)
	response, _ := ioutil.ReadAll(rec.Body)
	s.Contains(string(response), "DuplicateFile")
}

func (s UploadTestSuite) TestInvalidContentReturns500() {
	st := func(ctx context.Context, uf files.StoreMetadata, r files.Resumable, fileContent []byte) (bool, error) {
		return false, files.ErrFileAPICreateInvalidData
	}

	b, formWriter := generateFormWriter("valid")
	part, _ := formWriter.CreateFormFile("file", "testing.csv")
	part.Write([]byte("TEST DATA"))
	formWriter.Close()

	h := api.CreateV1UploadHandler(st)
	h.ServeHTTP(rec, generateRequest(b, formWriter))

	s.Equal(http.StatusInternalServerError, rec.Code)
	response, _ := ioutil.ReadAll(rec.Body)
	s.Contains(string(response), "RemoteValidationError")
}

func (s UploadTestSuite) TestUnexpectedErrorReturns500() {
	st := func(ctx context.Context, uf files.StoreMetadata, r files.Resumable, fileContent []byte) (bool, error) {
		return false, errors.New("its broken")
	}

	b, formWriter := generateFormWriter("valid")
	part, _ := formWriter.CreateFormFile("file", "testing.csv")
	part.Write([]byte("TEST DATA"))
	formWriter.Close()

	h := api.CreateV1UploadHandler(st)
	h.ServeHTTP(rec, generateRequest(b, formWriter))

	s.Equal(http.StatusInternalServerError, rec.Code)
	response, _ := ioutil.ReadAll(rec.Body)
	s.Contains(string(response), "InternalError")
}

func generateFormWriter(path string) (*bytes.Buffer, *multipart.Writer) {
	b := &bytes.Buffer{}
	formWriter := multipart.NewWriter(b)
	formWriter.WriteField("resumableFilename", "file.csv")
	formWriter.WriteField("path", path)
	formWriter.WriteField("isPublishable", "false")
	formWriter.WriteField("collectionId", "1234567890")
	formWriter.WriteField("title", "A New File")
	formWriter.WriteField("resumableTotalSize", "1478")
	formWriter.WriteField("resumableType", "text/csv")
	formWriter.WriteField("licence", "OGL v3")
	formWriter.WriteField("licenceUrl", "http://www.nationalarchives.gov.uk/doc/open-government-licence/version/3/")
	return b, formWriter
}

func generateRequest(b *bytes.Buffer, formWriter *multipart.Writer) *http.Request {
	req, _ := http.NewRequest(http.MethodPost, UploadURI, b)
	req.Header.Set("Content-Type", formWriter.FormDataContentType())
	return req
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
