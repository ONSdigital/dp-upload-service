package steps

import (
	"bytes"
	"context"
	"encoding/hex"
	"fmt"
	s3client "github.com/ONSdigital/dp-s3/v2"
	"github.com/ONSdigital/dp-upload-service/encryption"
	"io"
	"io/ioutil"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"os"
	"strconv"
	"strings"

	"github.com/ONSdigital/dp-net/v2/request"

	"github.com/ONSdigital/dp-upload-service/config"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
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
	ctx.Step(`^encryption key will be "([^"]*)"$`, c.encryptionKeyWillBe)

	// Whens
	ctx.Step(`^I upload the file "([^"]*)" with the following form resumable parameters:$`, c.iUploadTheFileWithTheFollowingFormResumableParameters)
	ctx.Step(`^I upload the file "([^"]*)" with the following form resumable parameters and auth header "([^"]*)"$`, c.iUploadTheFileWithTheFollowingFormResumableParametersAndAuthHeader)

	// Thens
	ctx.Step(`^the path "([^"]*)" should be available in the S3 bucket matching content using encryption key "([^"]*)":`, c.theFileShouldBeAvailableInTheSBucketMatchingContent)
	ctx.Step(`^the file upload should be marked as started using payload:$`, c.theFileUploadOfShouldBeMarkedAsStartedUsingPayload)
	ctx.Step(`^the file "([^"]*)" should be marked as uploaded using payload:$`, c.theFileUploadOfShouldBeMarkedAsUploadedUsingPayload)
	ctx.Step(`^the stored file "([^"]*)" should match the sent file "([^"]*)" using encryption key "([^"]*)"$`, c.theStoredFileShouldMatchTheSentFile)
	ctx.Step(`^the encryption key "([^"]*)" should be stored against file "([^"]*)"$`, c.theEncryptionKeyShouldBeStored)
	ctx.Step(`^the files api POST request should contain a default authorization header$`, c.theFilesApiPOSTRequestShouldContainADefaultAuthorizationHeader)
	ctx.Step(`^the files api PATCH request with path \("([^"]*)"\) should contain a default authorization header$`, c.theFilesApiPATCHRequestWithPathShouldContainADefaultAuthorizationHeader)
	// Buts
	ctx.Step(`^the file should not be marked as uploaded$`, c.theFileShouldNotBeMarkedAsUploaded)
	ctx.Step(`^the file upload should not have been registered again$`, c.theFileUploadShouldNotHaveBeenRegisteredAgain)

}

// ------
// Givens
// ------

func (c *UploadComponent) encryptionKeyWillBe(key string) error {
	bytes, err := hex.DecodeString(key)
	if err != nil {
		return err
	}

	c.EncryptionKey = bytes
	return nil
}

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
		body, _ := ioutil.ReadAll(r.Body)
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
		if strings.HasSuffix(r.URL.Path, "/valid") {
			w.WriteHeader(http.StatusOK)
			w.Header().Set("Content-Type", "application/json")
			w.Write([]byte(jsonResponse.Content))
		} else {
			w.WriteHeader(http.StatusNotFound)
		}
	}))

	pathAndFilename := path + "/" + filename
	encryptedKey := encryption.CreateKey()

	//setup vault
	cfg, _ := config.Get()
	vault, _ := c.svcList.GetVault(context.Background(), cfg)
	_ = vault.WriteKey(fmt.Sprintf("%s/%s", cfg.VaultPath, pathAndFilename), "key", hex.EncodeToString(encryptedKey))

	//setup s3
	s3Client, _ := c.svcList.GetS3StaticFileUploader(context.Background(), cfg)
	_, err := s3Client.UploadPartWithPsk(context.Background(), &s3client.UploadPartRequest{
		UploadKey:   pathAndFilename,
		Type:        "type",
		ChunkNumber: 1,
		TotalChunks: 1,
		FileName:    filename,
	}, []byte("content"), encryptedKey)

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

	c.ApiFeature.HttpResponse = w.Result()

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

	c.ApiFeature.HttpResponse = w.Result()

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

	c.ApiFeature.HttpResponse = w.Result()

	return nil
}

// -----
// Thens
// -----

func (c *UploadComponent) theStoredFileShouldMatchTheSentFile(s3Filename, localFilename, encryptionKey string) error {
	expectedPayload, err := os.ReadFile(localFilename)
	if err != nil {
		return err
	}

	return c.theFileShouldBeAvailableInTheSBucketMatchingContent(s3Filename, encryptionKey, &godog.DocString{Content: string(expectedPayload)})
}

func (c *UploadComponent) theFileShouldBeAvailableInTheSBucketMatchingContent(filename, encryptionKey string, expectedFileContent *godog.DocString) error {
	cfg, _ := config.Get()
	s, _ := session.NewSession(&aws.Config{
		Endpoint:         aws.String(localStackHost),
		Region:           aws.String(cfg.AwsRegion),
		S3ForcePathStyle: aws.Bool(true),
		Credentials:      credentials.NewStaticCredentials("test", "test", ""),
	})

	buf := aws.WriteAtBuffer{}
	s3client := s3.New(s)

	dl := s3manager.NewDownloaderWithClient(s3client)

	_, err := s3client.HeadObject(&s3.HeadObjectInput{
		Bucket: aws.String(cfg.StaticFilesEncryptedBucketName),
		Key:    aws.String(filename),
	})

	assert.NoError(c.ApiFeature, err)

	_, err = dl.Download(&buf, &s3.GetObjectInput{
		Bucket: aws.String(cfg.StaticFilesEncryptedBucketName),
		Key:    aws.String(filename),
	})

	assert.NoError(c.ApiFeature, err)

	d, _ := hex.DecodeString(encryptionKey)
	reader := &cryptoReader{
		reader:    ioutil.NopCloser(bytes.NewReader(buf.Bytes())),
		psk:       d,
		chunkSize: 5 * 1024 * 1024,
		currChunk: nil,
	}

	unencryptedBytes, _ := io.ReadAll(reader)

	assert.Equal(c.ApiFeature, expectedFileContent.Content, string(unencryptedBytes))

	return c.ApiFeature.StepError()
}

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

func (c *UploadComponent) theEncryptionKeyShouldBeStored(expectedEncryptionKey, filepath string) error {
	cfg, _ := config.Get()

	vault, _ := c.svcList.GetVault(context.Background(), cfg)
	actualEncryptionKey, err := vault.ReadKey(fmt.Sprintf("%s/%s", cfg.VaultPath, filepath), "key")

	assert.NoError(c.ApiFeature, err)
	assert.Equal(c.ApiFeature, expectedEncryptionKey, actualEncryptionKey)

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
