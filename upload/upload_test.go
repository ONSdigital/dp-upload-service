package upload_test

import (
	"bytes"
	"context"
	"errors"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"testing"

	s3client "github.com/ONSdigital/dp-s3/v2"
	"github.com/ONSdigital/dp-upload-service/aws"
	mock_aws "github.com/ONSdigital/dp-upload-service/aws/mock"
	"github.com/ONSdigital/dp-upload-service/upload"

	. "github.com/smartystreets/goconvey/convey"
)

var (
	s3Region = "eu-west-2"
	s3Bucket = "test-bucket"

	expectedPayload  = []byte(`some test file bytes to be uploaded`)
	fakeKeyGenerator = func() ([]byte, error) { return []byte("testing"), nil }
)

func TestGetUpload(t *testing.T) {

	Convey("given a GET /upload request", t, func() {

		w := httptest.NewRecorder()
		req, err := http.NewRequest("GET", "/upload", nil)
		So(err, ShouldBeNil)

		Convey("A 404 http status code is returned if chunk has not yet been uploaded", func() {

			addQueryParams(req, "1", "1")

			// S3 client returns ErrNotUploaded if uploadID cannot be found
			s3 := &mock_aws.S3ClienterMock{
				CheckPartUploadedFunc: func(ctx context.Context, req *s3client.UploadPartRequest) (bool, error) {
					return false, s3client.NewErrNotUploaded(errors.New("Not working"), nil)
				},
			}
			bucket := aws.NewBucket(s3Region, s3Bucket, s3)
			up := upload.New(bucket)
			up.CheckUploaded(w, req)

			// Validations
			So(len(s3.CheckPartUploadedCalls()), ShouldEqual, 1)
			So(s3.CheckPartUploadedCalls()[0].Req, ShouldResemble, &s3client.UploadPartRequest{
				UploadKey:   "12345",
				Type:        "text/plain",
				ChunkNumber: 1,
				TotalChunks: 1,
				FileName:    "helloworld",
			})
			So(w.Code, ShouldEqual, 404)
		})

		Convey("A 200 http status code is returned if chunk has been uploaded", func() {

			addQueryParams(req, "1", "2")

			// S3 client returns true if upload could be found and chunk was already uploaded
			s3 := &mock_aws.S3ClienterMock{
				CheckPartUploadedFunc: func(ctx context.Context, req *s3client.UploadPartRequest) (bool, error) {
					return true, nil
				},
			}
			bucket := aws.NewBucket(s3Region, s3Bucket, s3)
			up := upload.New(bucket)
			up.CheckUploaded(w, req)

			// Validations
			So(len(s3.CheckPartUploadedCalls()), ShouldEqual, 1)
			So(s3.CheckPartUploadedCalls()[0].Req, ShouldResemble, &s3client.UploadPartRequest{
				UploadKey:   "12345",
				Type:        "text/plain",
				ChunkNumber: 1,
				TotalChunks: 2,
				FileName:    "helloworld",
			})
			So(w.Code, ShouldEqual, 200)
		})

		Convey("A 500 http status code is returned if multi part upload fails", func() {

			addQueryParams(req, "1", "1")

			// S3 client returns generic error if ListMultipartUploads fails
			s3 := &mock_aws.S3ClienterMock{
				CheckPartUploadedFunc: func(ctx context.Context, req *s3client.UploadPartRequest) (bool, error) {
					return false, errors.New("could not list uploads")
				},
			}
			bucket := aws.NewBucket(s3Region, s3Bucket, s3)
			up := upload.New(bucket)
			up.CheckUploaded(w, req)

			// Validations
			So(len(s3.CheckPartUploadedCalls()), ShouldEqual, 1)
			So(s3.CheckPartUploadedCalls()[0].Req, ShouldResemble, &s3client.UploadPartRequest{
				UploadKey:   "12345",
				Type:        "text/plain",
				ChunkNumber: 1,
				TotalChunks: 1,
				FileName:    "helloworld",
			})
			So(w.Code, ShouldEqual, 500)
		})
	})
}

func TestPostUpload(t *testing.T) {

	Convey("given a POST /upload request", t, func() {

		w := httptest.NewRecorder()
		req, err := createTestFileUploadPart(expectedPayload)
		So(err, ShouldBeNil)

		Convey("test upload successfully uploads with only one chunk", func() {
			addQueryParams(req, "1", "1")

			// S3 client returns generic error if ListMultipartUploads fails
			s3 := &mock_aws.S3ClienterMock{
				UploadPartFunc: func(ctx context.Context, req *s3client.UploadPartRequest, payload []byte) (s3client.MultipartUploadResponse, error) {
					return s3client.MultipartUploadResponse{}, nil
				},
			}
			bucket := aws.NewBucket(s3Region, s3Bucket, s3)
			up := upload.New(bucket)
			up.Upload(w, req)

			// Validations
			So(len(s3.UploadPartCalls()), ShouldEqual, 1)
			So(s3.UploadPartCalls()[0].Req, ShouldResemble, &s3client.UploadPartRequest{
				UploadKey:   "12345",
				Type:        "text/plain",
				ChunkNumber: 1,
				TotalChunks: 1,
				FileName:    "helloworld",
			})
			So(s3.UploadPartCalls()[0].Payload, ShouldResemble, expectedPayload)
			So(w.Code, ShouldEqual, 200)

		})

		Convey("test 500 status returned if client throws an error", func() {
			addQueryParams(req, "1", "1")

			// S3 client returns generic error if ListMultipartUploads fails
			s3 := &mock_aws.S3ClienterMock{
				UploadPartFunc: func(ctx context.Context, req *s3client.UploadPartRequest, payload []byte) (s3client.MultipartUploadResponse, error) {
					return s3client.MultipartUploadResponse{}, errors.New("could not list uploads")
				},
			}
			bucket := aws.NewBucket(s3Region, s3Bucket, s3)
			up := upload.New(bucket)
			up.Upload(w, req)

			// Validations
			So(len(s3.UploadPartCalls()), ShouldEqual, 1)
			So(s3.UploadPartCalls()[0].Req, ShouldResemble, &s3client.UploadPartRequest{
				UploadKey:   "12345",
				Type:        "text/plain",
				ChunkNumber: 1,
				TotalChunks: 1,
				FileName:    "helloworld",
			})
			So(w.Code, ShouldEqual, 500)

		})

	})

}

func TestGetS3Url(t *testing.T) {

	Convey("Given a GET /upload request with a path parameter", t, func() {
		w := httptest.NewRecorder()
		req, err := http.NewRequest("GET", "/upload?:id=173849-helloworldtxt", nil)
		So(err, ShouldBeNil)

		Convey("A 200 OK status is returned, with the fully qualified s3 url for the region, bucket and s3 key", func() {
			bucket := aws.NewBucket(s3Region, s3Bucket, &mock_aws.S3ClienterMock{})
			up := upload.New(bucket)
			up.GetS3URL(w, req)

			// Validations
			So(w.Code, ShouldEqual, 200)
			So(w.Body.String(), ShouldEqual, `{"url":"https://s3-eu-west-2.amazonaws.com/test-bucket/173849-helloworldtxt"}`)
			So(w.Header().Get("Content-Type"), ShouldEqual, "application/json")
		})
	})
}

// createTestFileUploadPart creates an http Request with the expected body payload
func createTestFileUploadPart(testPayload []byte) (*http.Request, error) {
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)

	part, err := writer.CreateFormFile("file", "helloworld.txt")
	if err != nil {
		return nil, err
	}

	if _, err = part.Write(testPayload); err != nil {
		return nil, err
	}
	err = writer.Close()
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest("POST", "/upload", body)
	req.Header.Set("Content-Type", writer.FormDataContentType())

	return req, err
}

func addQueryParams(req *http.Request, chunkNumber, maxChunks string) {
	q := req.URL.Query()
	q.Add("resumableChunkNumber", chunkNumber)
	q.Add("resumableChunkSize", "5242880")
	q.Add("resumableCurrentChunkSize", "5242880")
	q.Add("resumableTotalSize", "5242880")
	q.Add("resumableType", "text/plain")
	q.Add("resumableIdentifier", "12345")
	q.Add("resumableFilename", "helloworld")
	q.Add("resumableRelativePath", ".")
	q.Add("resumableTotalChunks", maxChunks)
	req.URL.RawQuery = q.Encode()
}
