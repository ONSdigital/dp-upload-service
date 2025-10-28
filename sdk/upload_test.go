package sdk

import (
	"bytes"
	"context"
	"errors"
	"io"
	"mime/multipart"
	"net/http"
	"path/filepath"
	"strconv"
	"testing"
	"testing/iotest"

	"github.com/ONSdigital/dp-upload-service/api"
	. "github.com/smartystreets/goconvey/convey"
)

var (
	expectedReadErr = errors.New("intentional read error")
	brokenReader    = io.NopCloser(iotest.ErrReader(expectedReadErr))

	trueBool     = true
	collectionId = "collection1"
	bundleId     = "bundle1"

	validMetadata = api.Metadata{
		Path:          "path/to/data.csv",
		IsPublishable: &trueBool,
		CollectionId:  &collectionId,
		Title:         "Test Title",
		SizeInBytes:   1,
		Type:          "text/csv",
		Licence:       "Test Licence",
		LicenceUrl:    "http://example.com/licence",
	}
)

func TestUpload_Success(t *testing.T) {
	t.Parallel()

	Convey("Given a client, fileContent, metadata and headers", t, func() {
		mockClienter := newMockClienter(&http.Response{StatusCode: http.StatusCreated}, nil)
		client := newMockUploadServiceClient(mockClienter)

		fileContent := io.NopCloser(bytes.NewReader([]byte("This is some test file content.")))
		metadata := validMetadata

		Convey("When Upload is called with a single chunk", func() {
			uploadResp, err := client.Upload(context.Background(), fileContent, metadata, Headers{})

			Convey("Then no error is returned", func() {
				So(err, ShouldBeNil)
			})

			Convey("And the expected UploadResponse is returned", func() {
				So(uploadResp, ShouldNotBeNil)
				So(uploadResp.StatusCode, ShouldEqual, http.StatusCreated)
				So(uploadResp.Errors, ShouldBeNil)
			})

			Convey("And the mock clienter's Do method is called once with the correct request details", func() {
				So(mockClienter.DoCalls(), ShouldHaveLength, 1)
				actualCall := mockClienter.DoCalls()[0]
				So(actualCall.Req.Method, ShouldEqual, http.MethodPost)
				So(actualCall.Req.URL.String(), ShouldEqual, uploadServiceURL+"/upload-new")
				So(actualCall.Req.Header.Get("Content-Type"), ShouldStartWith, "multipart/form-data; boundary=")
				So(actualCall.Req.Body, ShouldNotBeNil)
			})
		})

		Convey("When Upload is called with multiple chunks", func() {
			metadata.SizeInBytes = chunkSize*2 + 1
			uploadResp, err := client.Upload(context.Background(), fileContent, metadata, Headers{})

			Convey("Then no error is returned", func() {
				So(err, ShouldBeNil)
			})

			Convey("And the expected UploadResponse is returned", func() {
				So(uploadResp, ShouldNotBeNil)
				So(uploadResp.StatusCode, ShouldEqual, http.StatusCreated)
				So(uploadResp.Errors, ShouldBeNil)
			})

			Convey("And the mock clienter's Do method is called three times", func() {
				So(mockClienter.DoCalls(), ShouldHaveLength, 3)
			})
		})
	})
}

func TestUpload_Failure(t *testing.T) {
	t.Parallel()

	Convey("When the metadata SizeInBytes is more than the maxFileSize", t, func() {
		mockClienter := newMockClienter(&http.Response{StatusCode: http.StatusOK}, nil)
		client := newMockUploadServiceClient(mockClienter)

		fileContent := io.NopCloser(bytes.NewReader([]byte("This is some test file content.")))
		metadata := validMetadata
		metadata.SizeInBytes = maxFileSize + 1

		Convey("And Upload is called", func() {
			uploadResp, err := client.Upload(context.Background(), fileContent, metadata, Headers{})

			Convey("Then the expected error is returned", func() {
				So(err, ShouldNotBeNil)
				So(err, ShouldEqual, ErrFileTooLarge)
			})

			Convey("And no UploadResponse is returned", func() {
				So(uploadResp, ShouldBeNil)
			})

			Convey("And the mock clienter's Do method is not called", func() {
				So(mockClienter.DoCalls(), ShouldHaveLength, 0)
			})
		})
	})

	Convey("When the fileContent returns an error on read", t, func() {
		mockClienter := newMockClienter(&http.Response{StatusCode: http.StatusOK}, nil)
		client := newMockUploadServiceClient(mockClienter)

		metadata := validMetadata

		Convey("And Upload is called", func() {
			uploadResp, err := client.Upload(context.Background(), brokenReader, metadata, Headers{})

			Convey("Then the expected error is returned", func() {
				So(err, ShouldNotBeNil)
				So(err, ShouldEqual, expectedReadErr)
			})

			Convey("And no UploadResponse is returned", func() {
				So(uploadResp, ShouldBeNil)
			})

			Convey("And the mock clienter's Do method is not called", func() {
				So(mockClienter.DoCalls(), ShouldHaveLength, 0)
			})
		})
	})

	Convey("When the Client's Do() function fails", t, func() {
		expectedDoErr := errors.New("intentional Do error")
		mockClienter := newMockClienter(nil, expectedDoErr)
		client := newMockUploadServiceClient(mockClienter)

		fileContent := io.NopCloser(bytes.NewReader([]byte("This is some test file content.")))
		metadata := validMetadata

		Convey("And Upload is called", func() {
			uploadResp, err := client.Upload(context.Background(), fileContent, metadata, Headers{})

			Convey("Then the expected error is returned", func() {
				So(err, ShouldNotBeNil)
				So(err, ShouldEqual, expectedDoErr)
			})

			Convey("And no UploadResponse is returned", func() {
				So(uploadResp, ShouldBeNil)
			})

			Convey("And the mock clienter's Do method is called once", func() {
				So(mockClienter.DoCalls(), ShouldHaveLength, 1)
			})
		})
	})

	Convey("When the upload service returns an unexpected status code", t, func() {
		body := `{"errors":[{"code":"InternalError","description":"internal server error occurred"}]}`
		mockClienter := newMockClienter(
			&http.Response{
				StatusCode: http.StatusInternalServerError,
				Body:       io.NopCloser(bytes.NewReader([]byte(body))),
			}, nil)
		client := newMockUploadServiceClient(mockClienter)

		fileContent := io.NopCloser(bytes.NewReader([]byte("This is some test file content.")))
		metadata := validMetadata

		Convey("And Upload is called", func() {
			uploadResp, err := client.Upload(context.Background(), fileContent, metadata, Headers{})

			Convey("Then no error is returned", func() {
				So(err, ShouldBeNil)
			})

			Convey("And the expected UploadResponse is returned", func() {
				So(uploadResp, ShouldNotBeNil)
				So(uploadResp.StatusCode, ShouldEqual, http.StatusInternalServerError)
				So(uploadResp.Errors, ShouldNotBeNil)
				So(uploadResp.Errors.Error, ShouldHaveLength, 1)
				So(uploadResp.Errors.Error[0].Code, ShouldEqual, "InternalError")
				So(uploadResp.Errors.Error[0].Description, ShouldEqual, "internal server error occurred")
			})

			Convey("And the mock clienter's Do method is called once", func() {
				So(mockClienter.DoCalls(), ShouldHaveLength, 1)
			})
		})
	})
}

func TestCreateUploadRequestBody(t *testing.T) {
	t.Parallel()

	Convey("Given valid chunkInfo, fileContent and metadata", t, func() {
		fileContent := io.NopCloser(bytes.NewReader([]byte("This is some test file content.")))
		chunkInfo := ChunkInfo{
			Current: 1,
			Total:   2,
		}
		metadata := validMetadata

		Convey("When createUploadRequestBody is called", func() {
			reqBuff, contentType, err := createUploadRequestBody(chunkInfo, fileContent, metadata)

			Convey("Then no error is returned", func() {
				So(err, ShouldBeNil)
			})

			Convey("And the request body contains the correct multipart form data", func() {
				body := reqBuff.String()
				So(body, ShouldContainSubstring, *metadata.CollectionId)

				So(body, ShouldContainSubstring, metadata.Path)
				So(body, ShouldContainSubstring, strconv.FormatBool(*metadata.IsPublishable))
				So(body, ShouldContainSubstring, metadata.Title)
				So(body, ShouldContainSubstring, metadata.Licence)
				So(body, ShouldContainSubstring, metadata.LicenceUrl)
				So(body, ShouldContainSubstring, metadata.Type)
				So(body, ShouldContainSubstring, strconv.Itoa(metadata.SizeInBytes))
				So(body, ShouldContainSubstring, filepath.Base(metadata.Path))

				So(body, ShouldContainSubstring, strconv.Itoa(chunkInfo.Current))
				So(body, ShouldContainSubstring, strconv.Itoa(chunkInfo.Total))
			})

			Convey("And the content type is correctly set", func() {
				So(contentType, ShouldStartWith, "multipart/form-data; boundary=")
			})
		})

		Convey("When createUploadRequestBody is called with a reader that returns an error", func() {
			reqBuff, contentType, err := createUploadRequestBody(chunkInfo, brokenReader, metadata)

			Convey("Then the expected error is returned", func() {
				So(err, ShouldNotBeNil)
				So(err, ShouldEqual, expectedReadErr)
			})

			Convey("And no request body is returned", func() {
				So(reqBuff, ShouldBeNil)
			})

			Convey("And no content type is returned", func() {
				So(contentType, ShouldEqual, "")
			})
		})
	})
}

func TestChunkReader(t *testing.T) {
	t.Parallel()

	Convey("Given fileContent smaller than chunkSize", t, func() {
		data := []byte("small file")
		reader := io.NopCloser(bytes.NewReader(data))

		Convey("When chunkReader is called", func() {
			chunk, length, err := chunkReader(reader)

			Convey("Then no error is returned", func() {
				So(err, ShouldBeNil)
			})

			Convey("And it returns the correct chunk and length", func() {
				So(length, ShouldEqual, len(data))

				chunkData := make([]byte, length)
				_, readErr := chunk.Read(chunkData)
				So(readErr, ShouldBeNil)
				So(chunkData, ShouldEqual, data)
			})
		})
	})

	Convey("Given fileContent larger than chunkSize", t, func() {
		data := bytes.Repeat([]byte("a"), chunkSize+10)
		reader := io.NopCloser(bytes.NewReader(data))

		Convey("When chunkReader is called", func() {
			chunk, length, err := chunkReader(reader)

			Convey("Then no error is returned", func() {
				So(err, ShouldBeNil)
			})

			Convey("And it returns the correct chunk and length", func() {
				So(length, ShouldEqual, chunkSize)

				chunkData := make([]byte, length)
				_, readErr := chunk.Read(chunkData)
				So(readErr, ShouldBeNil)
				So(chunkData, ShouldEqual, bytes.Repeat([]byte("a"), chunkSize))
			})
		})
	})

	Convey("Given fileContent is exactly chunkSize", t, func() {
		data := bytes.Repeat([]byte("b"), chunkSize)
		reader := io.NopCloser(bytes.NewReader(data))

		Convey("When chunkReader is called", func() {
			chunk, length, err := chunkReader(reader)

			Convey("Then no error is returned", func() {
				So(err, ShouldBeNil)
			})

			Convey("And it returns the correct chunk and length", func() {
				So(length, ShouldEqual, chunkSize)

				chunkData := make([]byte, length)
				_, readErr := chunk.Read(chunkData)
				So(readErr, ShouldBeNil)
				So(chunkData, ShouldEqual, data)
			})
		})
	})

	Convey("Given fileContent is empty", t, func() {
		data := []byte{}
		reader := io.NopCloser(bytes.NewReader(data))

		Convey("When chunkReader is called", func() {
			chunk, length, err := chunkReader(reader)

			Convey("Then no error is returned", func() {
				So(err, ShouldBeNil)
			})

			Convey("And it returns zero length", func() {
				So(length, ShouldEqual, 0)

				chunkData := make([]byte, length)
				_, readErr := chunk.Read(chunkData)
				// EOF is expected here since there's no data to read
				So(readErr, ShouldEqual, io.EOF)
				So(chunkData, ShouldEqual, data)
			})
		})
	})

	Convey("Given fileContent that returns an error on read", t, func() {
		Convey("When chunkReader is called", func() {
			chunk, length, err := chunkReader(brokenReader)

			Convey("Then the expected error is returned", func() {
				So(err, ShouldNotBeNil)
				So(err, ShouldEqual, expectedReadErr)
				So(length, ShouldEqual, 0)
				So(chunk, ShouldBeNil)
			})
		})
	})
}

func TestWriteMetadataFormFields(t *testing.T) {
	t.Parallel()

	Convey("Given a multipart writer, fully populated metadata and chunkInfo", t, func() {
		var buf bytes.Buffer
		writer := multipart.NewWriter(&buf)

		// In a real scenario only one of CollectionId or BundleId would be set
		// We test both here for test coverage
		metadata := validMetadata
		metadata.BundleId = &bundleId

		chunkInfo := ChunkInfo{
			Current: 1,
			Total:   3,
		}

		Convey("When writeMetadataFormFields is called", func() {
			err := writeMetadataFormFields(writer, metadata, chunkInfo)
			closeErr := writer.Close()
			So(closeErr, ShouldBeNil)

			Convey("Then no error is returned", func() {
				So(err, ShouldBeNil)
			})

			Convey("And the multipart body contains the correct form fields", func() {
				body := buf.String()
				So(body, ShouldContainSubstring, *metadata.CollectionId)
				So(body, ShouldContainSubstring, *metadata.BundleId)

				So(body, ShouldContainSubstring, metadata.Path)
				So(body, ShouldContainSubstring, strconv.FormatBool(*metadata.IsPublishable))
				So(body, ShouldContainSubstring, metadata.Title)
				So(body, ShouldContainSubstring, metadata.Licence)
				So(body, ShouldContainSubstring, metadata.LicenceUrl)
				So(body, ShouldContainSubstring, metadata.Type)
				So(body, ShouldContainSubstring, strconv.Itoa(metadata.SizeInBytes))
				So(body, ShouldContainSubstring, filepath.Base(metadata.Path))

				So(body, ShouldContainSubstring, strconv.Itoa(chunkInfo.Current))
				So(body, ShouldContainSubstring, strconv.Itoa(chunkInfo.Total))
			})
		})
	})
}

func TestWriteFileFormField(t *testing.T) {
	t.Parallel()

	Convey("Given a multipart writer and file content", t, func() {
		var buf bytes.Buffer
		writer := multipart.NewWriter(&buf)

		fileContent := []byte("test file content")
		filename := "test.txt"
		fileReader := bytes.NewReader(fileContent)

		Convey("When writeFileFormField is called", func() {
			n, err := writeFileFormField(writer, fileReader, len(fileContent), filename)
			closeErr := writer.Close()
			So(closeErr, ShouldBeNil)

			Convey("Then no error is returned and correct number of bytes are written", func() {
				So(err, ShouldBeNil)
				So(n, ShouldEqual, len(fileContent))
			})

			Convey("And the multipart body contains the correct file content", func() {
				body := buf.String()
				So(body, ShouldContainSubstring, filename)
				So(body, ShouldContainSubstring, string(fileContent))
			})
		})

		Convey("When writeFileFormField is called with a reader that returns an error", func() {
			n, err := writeFileFormField(writer, brokenReader, len(fileContent), filename)
			closeErr := writer.Close()
			So(closeErr, ShouldBeNil)

			Convey("Then the expected error is returned", func() {
				So(err, ShouldNotBeNil)
				So(err, ShouldEqual, expectedReadErr)
				So(n, ShouldEqual, 0)
			})
		})
	})
}
