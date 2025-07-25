package steps

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"os"
	"strconv"
	"strings"

	s3client "github.com/ONSdigital/dp-s3/v3"

	"github.com/ONSdigital/dp-net/v3/request"

	"github.com/ONSdigital/dp-upload-service/config"
	"github.com/cucumber/godog"
	"github.com/pkg/errors"
	"github.com/rdumont/assistdog"
	"github.com/stretchr/testify/assert"
)

var requests map[string]string

const filesURI = "/files"

func (c *UploadComponent) RegisterSteps(ctx *godog.ScenarioContext) {
	// Givens
	ctx.Step(`^dp-files-api does not have a file "([^"]*)" registered$`, c.dpfilesapiDoesNotHaveAFileRegistered)
	ctx.Step(`^dp-files-api has a file with path "([^"]*)" and filename "([^"]*)" registered with meta-data:$`, c.dpfilesapiHasAFileWithPathAndFilenameRegisteredWithMetadata)
	ctx.Step(`^the data file "([^"]*)" with content:$`, c.theDataFile)
	ctx.Step(`^the file meta-data is:$`, c.theFileMetadataIs)
	ctx.Step(`^the 1st part of the file "([^"]*)" has been uploaded with resumable parameters:$`, c.the1StPartOfTheFileHasBeenUploaded)

	// Whens
	ctx.Step(`^I upload the file "([^"]*)" with the following form resumable parameters:$`, c.iUploadTheFileWithTheFollowingFormResumableParameters)
	ctx.Step(`^I upload the file "([^"]*)" with the following form resumable parameters and auth header "([^"]*)"$`, c.iUploadTheFileWithTheFollowingFormResumableParametersAndAuthHeader)

	// Thens
	ctx.Step(`^the file upload should be marked as started using payload:$`, c.theFileUploadOfShouldBeMarkedAsStartedUsingPayload)
	ctx.Step(`^the file "([^"]*)" should be marked as uploaded using payload:$`, c.theFileUploadOfShouldBeMarkedAsUploadedUsingPayload)
	ctx.Step(`^the files api POST request should contain a default authorization header$`, c.theFilesApiPOSTRequestShouldContainADefaultAuthorizationHeader)
	ctx.Step(`^the files api PATCH request with path \("([^"]*)"\) should contain a default authorization header$`, c.theFilesApiPATCHRequestWithPathShouldContainADefaultAuthorizationHeader)
	// Buts
	ctx.Step(`^the file should not be marked as uploaded$`, c.theFileShouldNotBeMarkedAsUploaded)
	ctx.Step(`^the file upload should not have been registered again$`, c.theFileUploadShouldNotHaveBeenRegisteredAgain)

}

// ------
// Givens
// ------

func (c *UploadComponent) theDataFile(filename string, fileContent *godog.DocString) error {
	file, err := os.Create(fmt.Sprintf("%s/%s", testFilePath, filename))
	if err != nil {
		return errors.New(fmt.Sprintf("Cannot create file: %s", err.Error()))
	}

	defer file.Close()

	_, err = file.Write([]byte(fileContent.Content))
	if err != nil {
		return errors.New("Cannot write to file")
	}

	return nil
}

func (c *UploadComponent) dpfilesapiDoesNotHaveAFileRegistered(filename string) error {
	requests = make(map[string]string)
	s := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		requests[fmt.Sprintf("%s|%s", r.URL.Path, r.Method)] = string(body)
		requests[fmt.Sprintf("%s|%s|auth", r.URL.Path, r.Method)] = r.Header.Get(request.AuthHeaderKey)

		if r.Method == http.MethodPost {
			w.WriteHeader(http.StatusCreated)
		} else {
			w.WriteHeader(http.StatusOK)
		}
	}))

	os.Setenv("FILES_API_URL", s.URL)

	return nil
}

func (c *UploadComponent) dpfilesapiHasAFileWithPathAndFilenameRegisteredWithMetadata(path, filename string, jsonResponse *godog.DocString) error {
	s := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPost {
			body, _ := io.ReadAll(r.Body)
			expectedPath := fmt.Sprintf("%s/%s", path, filename)
			if strings.Contains(string(body), expectedPath) {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusBadRequest)
				w.Write([]byte(`{"errors":[{"errorCode":"DuplicateFileError","description":"file already registered"}]}`))
				return
			}
			w.WriteHeader(http.StatusCreated)
			return
		}

		if strings.HasSuffix(r.URL.Path, "/valid") {
			w.WriteHeader(http.StatusOK)
			w.Header().Set("Content-Type", "application/json")
			w.Write([]byte(jsonResponse.Content))
		} else {
			w.WriteHeader(http.StatusNotFound)
		}
	}))

	pathAndFilename := path + "/" + filename
	cfg, _ := config.Get()

	//setup s3
	s3Client, _ := c.svcList.GetS3StaticFileUploader(context.Background(), cfg)
	_, err := s3Client.UploadPart(context.Background(), &s3client.UploadPartRequest{
		UploadKey:   pathAndFilename,
		Type:        "type",
		ChunkNumber: 1,
		TotalChunks: 1,
		FileName:    filename,
	}, []byte("content"))

	os.Setenv("FILES_API_URL", s.URL)

	return err
}

func (c *UploadComponent) theFileMetadataIs(table *godog.Table) error {
	assist := assistdog.NewDefault()
	c.fileMetadata, _ = assist.ParseMap(table)
	return nil
}

// -----
// Whens
// -----

func (c *UploadComponent) iUploadTheFileWithTheFollowingFormResumableParameters(filename string, table *godog.Table) error {
	b := &bytes.Buffer{}
	formWriter := multipart.NewWriter(b)

	for key, value := range c.fileMetadata {
		formWriter.WriteField(key, value)
	}

	part, _ := formWriter.CreateFormFile("file", filename)

	testPayload, err := os.ReadFile(filename)
	if err != nil {
		return err
	}

	assist := assistdog.NewDefault()
	queryParams, _ := assist.ParseMap(table)

	total, _ := strconv.ParseInt(queryParams["resumableTotalChunks"], 10, 32)
	current, _ := strconv.ParseInt(queryParams["resumableChunkNumber"], 10, 32)

	if total > 1 {
		if current == 1 {
			b := testPayload[:(5 * 1024 * 1024)]
			part.Write(b)
		} else if total > 1 {
			b := testPayload[(5 * 1024 * 1024):]
			part.Write(b)
		}
	} else {
		part.Write(testPayload)
	}

	formWriter.Close()

	handler, err := c.ApiFeature.Initialiser()
	if err != nil {
		return err
	}
	req := httptest.NewRequest(http.MethodPost, "http://foo/upload-new", b)
	req.Header.Set("Content-Type", formWriter.FormDataContentType())

	q := req.URL.Query()
	for key, value := range queryParams {
		q.Add(key, value)
	}
	req.URL.RawQuery = q.Encode()

	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	c.ApiFeature.HTTPResponse = w.Result()
	return nil
}

func (c *UploadComponent) iUploadTheFileWithTheFollowingFormResumableParametersAndAuthHeader(filename, authHeader string, table *godog.Table) error {
	b := &bytes.Buffer{}
	formWriter := multipart.NewWriter(b)

	for key, value := range c.fileMetadata {
		formWriter.WriteField(key, value)
	}

	part, _ := formWriter.CreateFormFile("file", filename)

	testPayload, err := os.ReadFile(filename)
	if err != nil {
		return err
	}

	assist := assistdog.NewDefault()
	queryParams, _ := assist.ParseMap(table)

	total, _ := strconv.ParseInt(queryParams["resumableTotalChunks"], 10, 32)
	current, _ := strconv.ParseInt(queryParams["resumableChunkNumber"], 10, 32)

	if total > 1 {
		if current == 1 {
			b := testPayload[:(5 * 1024 * 1024)]
			part.Write(b)
		} else if total > 1 {
			b := testPayload[(5 * 1024 * 1024):]
			part.Write(b)
		}
	} else {
		part.Write(testPayload)
	}

	formWriter.Close()

	handler, err := c.ApiFeature.Initialiser()
	if err != nil {
		return err
	}
	req := httptest.NewRequest(http.MethodPost, "http://foo/upload-new", b)
	req.Header.Set("Content-Type", formWriter.FormDataContentType())
	req.Header.Set("Authorization", authHeader)

	q := req.URL.Query()
	for key, value := range queryParams {
		q.Add(key, value)
	}
	req.URL.RawQuery = q.Encode()

	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	c.ApiFeature.HTTPResponse = w.Result()

	return nil
}

// deprecated
func (c *UploadComponent) iUploadTheFileWithMetaData(filename string, table *godog.Table) error {
	b := &bytes.Buffer{}
	formWriter := multipart.NewWriter(b)

	assist := assistdog.NewDefault()
	data, err := assist.ParseMap(table)
	for key, value := range data {
		formWriter.WriteField(key, value)
	}

	part, err := formWriter.CreateFormFile("file", filename)
	if err != nil {
		return err
	}

	testPayload, err := os.ReadFile(fmt.Sprintf("%s/%s", testFilePath, filename))
	if err != nil {
		return err
	}

	if _, err = part.Write(testPayload); err != nil {
		return err
	}
	err = formWriter.Close()
	if err != nil {
		return err
	}

	handler, err := c.ApiFeature.Initialiser()
	if err != nil {
		return err
	}
	req := httptest.NewRequest(http.MethodPost, "http://foo/upload-new", b)
	req.Header.Set("Content-Type", formWriter.FormDataContentType())

	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	c.ApiFeature.HTTPResponse = w.Result()

	return nil
}

// -----
// Thens
// -----

func (c *UploadComponent) theFileUploadOfShouldBeMarkedAsStartedUsingPayload(expectedFilesPayload *godog.DocString) error {
	assert.JSONEq(c.ApiFeature, expectedFilesPayload.Content, requests[fmt.Sprintf("%s|%s", filesURI, http.MethodPost)])

	return c.ApiFeature.StepError()
}

func (c *UploadComponent) theFileUploadOfShouldBeMarkedAsUploadedUsingPayload(filepath string, expectedFilesPayload *godog.DocString) error {
	assert.JSONEq(c.ApiFeature, expectedFilesPayload.Content, requests[fmt.Sprintf("%s/%s|%s", filesURI, filepath, http.MethodPatch)])
	return c.ApiFeature.StepError()
}

func (c *UploadComponent) theFileShouldNotBeMarkedAsUploaded() error {
	assert.NotContains(c.ApiFeature, requests, fmt.Sprintf("%s|%s", filesURI, http.MethodPatch))
	return c.ApiFeature.StepError()
}
func (c *UploadComponent) the1StPartOfTheFileHasBeenUploaded(filename string, table *godog.Table) error {
	err := c.iUploadTheFileWithTheFollowingFormResumableParameters(filename, table)

	requests = make(map[string]string)

	return err
}

func (c *UploadComponent) theFileUploadShouldNotHaveBeenRegisteredAgain() error {
	assert.NotContains(c.ApiFeature, requests, "/files")
	return c.ApiFeature.StepError()
}

func (c *UploadComponent) theFilesApiPOSTRequestShouldContainADefaultAuthorizationHeader() error {
	cfg, _ := config.Get()
	assert.Equal(c.ApiFeature, "Bearer "+cfg.ServiceAuthToken, requests[fmt.Sprintf("%s|%s|auth", filesURI, http.MethodPost)])
	return c.ApiFeature.StepError()
}

func (c *UploadComponent) theFilesApiPATCHRequestWithPathShouldContainADefaultAuthorizationHeader(filepath string) error {
	cfg, _ := config.Get()
	assert.Equal(c.ApiFeature, "Bearer "+cfg.ServiceAuthToken, requests[fmt.Sprintf("%s/%s|%s|auth", filesURI, filepath, http.MethodPatch)])
	return c.ApiFeature.StepError()
}
