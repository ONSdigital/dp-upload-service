package files_test

import (
	"context"
	"errors"
	"net/http"
	"testing"

	s3client "github.com/ONSdigital/dp-s3/v2"

	"github.com/ONSdigital/dp-upload-service/files"
	"github.com/ONSdigital/dp-upload-service/upload/mock"

	"github.com/maxcnunes/httpfake"
	"github.com/stretchr/testify/suite"
)

type StoreSuite struct {
	suite.Suite

	fakeFilesApi *httpfake.HTTPFake
	mockS3       *mock.S3ClienterMock
}

func TestStore(t *testing.T) {
	suite.Run(t, new(StoreSuite))
}

// beforeEach
func (s *StoreSuite) SetupTest() {
	s.fakeFilesApi = httpfake.New(httpfake.WithTesting(s.T()))
	s.mockS3 = &mock.S3ClienterMock{
		UploadPartFunc: func(ctx context.Context, req *s3client.UploadPartRequest, payload []byte) (s3client.MultipartUploadResponse, error) {
			return s3client.MultipartUploadResponse{Etag: "123456789", AllPartsUploaded: true}, nil
		},
	}
}

var (
	firstResumable = files.Resumable{CurrentChunk: 1}
	lastResumable  = files.Resumable{CurrentChunk: 2}
)

// afterEach
func (s *StoreSuite) TearDownTest() {
	s.fakeFilesApi.Close()
}

func (s *StoreSuite) TestFileUploadIsRegisteredWithFilesApi() {
	s.fakeFilesApi.NewHandler().Post("/v1/files/register").Reply(http.StatusCreated)
	s.fakeFilesApi.NewHandler().Post("/v1/files/upload-complete").Reply(http.StatusCreated)

	store := files.NewStore(s.fakeFilesApi.ResolveURL(""), s.mockS3)

	_, err := store.UploadFile(context.Background(), files.Metadata{}, firstResumable, []byte("CONTENT"))
	s.NoError(err)
}

func (s *StoreSuite) TestFileAlreadyRegisteredWithFilesApi() {
	s.fakeFilesApi.NewHandler().
		Post("/v1/files/register").
		Reply(http.StatusBadRequest).
		Body([]byte(`{"errors": [{"code": "DuplicateFileError", "description": "file already exists"}]}`))

	store := files.NewStore(s.fakeFilesApi.ResolveURL(""), s.mockS3)

	_, err := store.UploadFile(context.Background(), files.Metadata{}, firstResumable, []byte("CONTENT"))
	s.Equal(files.ErrFilesAPIDuplicateFile, err)
}

func (s *StoreSuite) TestFileRegisteredWithInvalidContent() {
	s.fakeFilesApi.NewHandler().
		Post("/v1/files/register").
		Reply(http.StatusBadRequest).
		Body([]byte(`{"errors": [{"code": "ValidationError", "description": "fields were invalid"}]}`))

	store := files.NewStore(s.fakeFilesApi.ResolveURL(""), s.mockS3)

	_, err := store.UploadFile(context.Background(), files.Metadata{}, firstResumable, []byte("CONTENT"))
	s.Equal(files.ErrFileAPICreateInvalidData, err)
}

func (s *StoreSuite) TestFileRegisterReturnsUnknownError() {
	s.fakeFilesApi.NewHandler().
		Post("/v1/files/register").
		Reply(http.StatusBadRequest).
		Body([]byte(`{"errors": [{"code": "SpecialError", "description": "fields were invalid"}]}`))

	store := files.NewStore(s.fakeFilesApi.ResolveURL(""), s.mockS3)

	_, err := store.UploadFile(context.Background(), files.Metadata{}, firstResumable, []byte("CONTENT"))
	s.Equal(files.ErrUnknownError, err)
}

func (s *StoreSuite) TestFileRegisterReturnsMalformedJSON() {
	s.fakeFilesApi.NewHandler().
		Post("/v1/files/register").
		Reply(http.StatusBadRequest).
		Body([]byte(`<json>Error occurred</json>`))

	store := files.NewStore(s.fakeFilesApi.ResolveURL(""), s.mockS3)

	_, err := store.UploadFile(context.Background(), files.Metadata{}, firstResumable, []byte("CONTENT"))

	s.Error(err)
}

func (s *StoreSuite) TestErrorConnectingToRegisterFiles() {
	store := files.NewStore("does.not.work", s.mockS3)

	_, err := store.UploadFile(context.Background(), files.Metadata{}, firstResumable, []byte("CONTENT"))
	s.Equal(files.ErrConnectingToFilesApi, err)
}

func (s StoreSuite) TestUploadPartReturnsAnError() {
	s.mockS3.UploadPartFunc = func(ctx context.Context, req *s3client.UploadPartRequest, payload []byte) (s3client.MultipartUploadResponse, error) {
		return s3client.MultipartUploadResponse{}, s3client.NewError(errors.New("broken"), nil)
	}

	store := files.NewStore(s.fakeFilesApi.ResolveURL(""), s.mockS3)

	_, err := store.UploadFile(context.Background(), files.Metadata{}, lastResumable, []byte("CONTENT"))
	s.Equal(files.ErrS3Upload, err)
}

func (s StoreSuite) TestUploadChunkTooSmallReturnsErrChuckTooSmall() {
	s.mockS3.UploadPartFunc = func(ctx context.Context, req *s3client.UploadPartRequest, payload []byte) (s3client.MultipartUploadResponse, error) {
		return s3client.MultipartUploadResponse{}, s3client.NewChunkTooSmallError(errors.New("chunk size below minimum 5MB"), nil)
	}

	store := files.NewStore(s.fakeFilesApi.ResolveURL(""), s.mockS3)

	_, err := store.UploadFile(context.Background(), files.Metadata{}, lastResumable, []byte("CONTENT"))
	s.Equal(files.ErrChunkTooSmall, err)
}

func (s StoreSuite) TestFileNotFoundWhenMarkedAsUploaded() {
	s.fakeFilesApi.NewHandler().Post("/v1/files/upload-complete").Reply(http.StatusNotFound)

	store := files.NewStore(s.fakeFilesApi.ResolveURL(""), s.mockS3)

	_, err := store.UploadFile(context.Background(), files.Metadata{}, lastResumable, []byte("CONTENT"))
	s.Equal(files.ErrFileNotFound, err)
}

func (s StoreSuite) TestReturnsConflictWhenFileInUnexpectedState() {
	s.fakeFilesApi.NewHandler().Post("/v1/files/upload-complete").Reply(http.StatusConflict)

	store := files.NewStore(s.fakeFilesApi.ResolveURL(""), s.mockS3)

	_, err := store.UploadFile(context.Background(), files.Metadata{}, lastResumable, []byte("CONTENT"))
	s.Equal(files.ErrFileStateConflict, err)
}

func (s StoreSuite) TestUploadCompleteUnknownError() {
	s.fakeFilesApi.NewHandler().Post("/v1/files/upload-complete").Reply(http.StatusTeapot)

	store := files.NewStore(s.fakeFilesApi.ResolveURL(""), s.mockS3)

	_, err := store.UploadFile(context.Background(), files.Metadata{}, lastResumable, []byte("CONTENT"))
	s.Equal(files.ErrUnknownError, err)
}

func (s *StoreSuite) TestErrorConnectingToUploadComplete() {
	store := files.NewStore("does.not.work", s.mockS3)

	_, err := store.UploadFile(context.Background(), files.Metadata{}, lastResumable, []byte("CONTENT"))
	s.Equal(files.ErrConnectingToFilesApi, err)
}
