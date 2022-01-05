package steps

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"os"

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

func (c *UploadComponent) RegisterSteps(ctx *godog.ScenarioContext) {
	// Givens
	ctx.Step(`^dp-files-api does not have a file "([^"]*)" registered$`, c.dpfilesapiDoesNotHaveAFileRegistered)
	ctx.Step(`^the data file "([^"]*)" with content:$`, c.theDataFile)

	// Whens
	ctx.Step(`^I upload the file "([^"]*)" with the following form meta-data:$`, c.iUploadTheFileWithMetaData)

	// Thens
	ctx.Step(`^the path "([^"]*)" should be available in the S3 bucket matching content:`, c.theFileShouldBeAvailableInTheSBucketMatchingContent)
	ctx.Step(`^the file upload should be marked as started using payload:$`, c.theFileUploadOfShouldBeMarkedAsStartedUsingPayload)
	ctx.Step(`^the file should be marked as uploaded using payload:$`, c.theFileUploadOfShouldBeMarkedAsUploadedUsingPayload)
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
		body, _ := ioutil.ReadAll(r.Body)
		requests[r.URL.Path] = string(body)
		w.WriteHeader(http.StatusCreated)
	}))

	os.Setenv("FILES_API_URL", s.URL)

	return nil
}

// -----
// Whens
// -----

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
	req := httptest.NewRequest(http.MethodPost, "http://foo/v1/upload", b)
	req.Header.Set("Content-Type", formWriter.FormDataContentType())

	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	c.ApiFeature.HttpResponse = w.Result()

	return nil
}

// -----
// Thens
// -----

func (c *UploadComponent) theFileShouldBeAvailableInTheSBucketMatchingContent(filename string, expectedFileContent *godog.DocString) error {
	assert.Equal(c.ApiFeature, http.StatusOK, c.ApiFeature.HttpResponse.StatusCode)

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
		Bucket: aws.String(cfg.UploadBucketName),
		Key:    aws.String(filename),
	})

	assert.NoError(c.ApiFeature, err)

	_, err = dl.Download(&buf, &s3.GetObjectInput{
		Bucket: aws.String(cfg.UploadBucketName),
		Key:    aws.String(filename),
	})

	assert.NoError(c.ApiFeature, err)
	assert.Equal(c.ApiFeature, expectedFileContent.Content, string(buf.Bytes()))

	return c.ApiFeature.StepError()
}

func (c *UploadComponent) theFileUploadOfShouldBeMarkedAsStartedUsingPayload(expectedFilesPayload *godog.DocString) error {
	assert.JSONEq(c.ApiFeature, expectedFilesPayload.Content, requests["/v1/files/register"])

	return c.ApiFeature.StepError()
}

func (c *UploadComponent) theFileUploadOfShouldBeMarkedAsUploadedUsingPayload(expectedFilesPayload *godog.DocString) error {
	assert.JSONEq(c.ApiFeature, expectedFilesPayload.Content, requests["/v1/files/upload-complete"])
	return c.ApiFeature.StepError()
}
