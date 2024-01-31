package files_test

import (
	"context"
	"errors"
	"testing"

	"github.com/ONSdigital/dp-upload-service/aws"
	"github.com/ONSdigital/dp-upload-service/config"
	"github.com/aws/aws-sdk-go/service/s3"

	"github.com/stretchr/testify/suite"

	filesAPI "github.com/ONSdigital/dp-api-clients-go/v2/files"
	s3client "github.com/ONSdigital/dp-s3/v2"
	mock_aws "github.com/ONSdigital/dp-upload-service/aws/mock"
	"github.com/ONSdigital/dp-upload-service/files"
	mock_files "github.com/ONSdigital/dp-upload-service/files/mock"
)

var (
	firstResumable = files.Resumable{CurrentChunk: 1}
	lastResumable  = files.Resumable{CurrentChunk: 2}
	content        = []byte("CONTENT")
)

type StoreSuite struct {
	suite.Suite

	mockS3    *mock_aws.S3ClienterMock
	mockFiles *mock_files.FilesClienterMock
	bucket    *aws.Bucket
}

func TestStore(t *testing.T) {
	suite.Run(t, new(StoreSuite))
}

// beforeEach
func (s *StoreSuite) SetupTest() {
	s.mockFiles = &mock_files.FilesClienterMock{
		RegisterFileFunc: func(ctx context.Context, metadata filesAPI.FileMetaData) error {
			return nil
		},
		MarkFileUploadedFunc: func(ctx context.Context, path string, etag string) error {
			return nil
		},
		GetFileFunc: func(ctx context.Context, path string, authToken string) (filesAPI.FileMetaData, error) {
			return filesAPI.FileMetaData{Path: path}, nil
		},
	}

	s.mockS3 = &mock_aws.S3ClienterMock{
		HeadFunc: func(key string) (*s3.HeadObjectOutput, error) {
			size := int64(100)
			etag := "head-object-etag"
			return &s3.HeadObjectOutput{ContentLength: &size, ETag: &etag}, nil
		},
		UploadPartFunc: func(ctx context.Context, req *s3client.UploadPartRequest, payload []byte) (s3client.MultipartUploadResponse, error) {
			return s3client.MultipartUploadResponse{Etag: "uploaded-part-etag", AllPartsUploaded: true}, nil
		},
	}
	s.bucket = aws.NewBucket("region", "name", s.mockS3)
}

// Upload
func (s *StoreSuite) TestFileUploadIsRegisteredWithFilesApi() {
	store := files.NewStore(s.mockFiles, s.bucket, &config.Config{})

	_, err := store.UploadFile(context.Background(), filesAPI.FileMetaData{}, firstResumable, content)
	s.NoError(err)
	s.Len(s.mockFiles.RegisterFileCalls(), 1)
}

func (s *StoreSuite) TestFileRegistrationFailsWithFilesApi() {
	expectedError := errors.New("registration error")
	s.mockFiles.RegisterFileFunc = func(ctx context.Context, metadata filesAPI.FileMetaData) error {
		return expectedError
	}

	store := files.NewStore(s.mockFiles, s.bucket, &config.Config{})

	_, err := store.UploadFile(context.Background(), filesAPI.FileMetaData{}, firstResumable, content)
	s.Equal(expectedError, err)
}

func (s StoreSuite) TestUploadPartReturnsAnError() {
	s.mockS3.UploadPartFunc = func(ctx context.Context, req *s3client.UploadPartRequest, payload []byte) (s3client.MultipartUploadResponse, error) {
		return s3client.MultipartUploadResponse{}, s3client.NewError(errors.New("broken"), nil)
	}

	store := files.NewStore(s.mockFiles, s.bucket, &config.Config{})

	_, err := store.UploadFile(context.Background(), filesAPI.FileMetaData{}, lastResumable, content)
	s.Equal(files.ErrS3Upload, err)
}

func (s StoreSuite) TestUploadChunkTooSmallReturnsErrChuckTooSmall() {
	s.mockS3.UploadPartFunc = func(ctx context.Context, req *s3client.UploadPartRequest, payload []byte) (s3client.MultipartUploadResponse, error) {
		return s3client.MultipartUploadResponse{}, s3client.NewChunkTooSmallError(errors.New("chunk size below minimum 5MB"), nil)
	}

	store := files.NewStore(s.mockFiles, s.bucket, &config.Config{})

	_, err := store.UploadFile(context.Background(), filesAPI.FileMetaData{}, lastResumable, content)
	s.Equal(files.ErrChunkTooSmall, err)
}

func (s StoreSuite) TestHeadReturnsAnError() {
	s.mockS3.HeadFunc = func(key string) (*s3.HeadObjectOutput, error) {
		return nil, errors.New("head error")
	}

	store := files.NewStore(s.mockFiles, s.bucket, &config.Config{})

	_, err := store.UploadFile(context.Background(), filesAPI.FileMetaData{}, lastResumable, content)
	s.ErrorIs(err, files.ErrS3Head)
}

func (s StoreSuite) TestHeadReturnsNoEtag() {
	s.mockS3.HeadFunc = func(key string) (*s3.HeadObjectOutput, error) {
		size := int64(100)
		return &s3.HeadObjectOutput{ContentLength: &size}, nil
	}

	store := files.NewStore(s.mockFiles, s.bucket, &config.Config{})

	_, err := store.UploadFile(context.Background(), filesAPI.FileMetaData{}, lastResumable, content)
	s.ErrorIs(err, files.ErrS3Head)
}

func (s StoreSuite) TestErrorMarkingAsUploaded() {
	expectedError := errors.New("marking error")
	s.mockFiles.MarkFileUploadedFunc = func(ctx context.Context, path string, etag string) error {
		return expectedError
	}
	store := files.NewStore(s.mockFiles, s.bucket, &config.Config{})

	_, err := store.UploadFile(context.Background(), filesAPI.FileMetaData{}, lastResumable, content)
	s.Equal(expectedError, err)
}

func (s StoreSuite) TestMarkingAsUploadedAddsCorrectEtag() {
	s.mockFiles.MarkFileUploadedFunc = func(ctx context.Context, path string, etag string) error {
		s.Equal("head-object-etag", etag)
		return nil
	}
	store := files.NewStore(s.mockFiles, s.bucket, &config.Config{})

	_, err := store.UploadFile(context.Background(), filesAPI.FileMetaData{}, lastResumable, content)
	s.NoError(err)
}

// Status
func (s *StoreSuite) TestStatusHappyPath() {
	store := files.NewStore(s.mockFiles, s.bucket, &config.Config{})
	response, err := store.Status(context.Background(), "valid")
	s.NoError(err)
	s.True(response.FileContent.Value)
	s.Len(s.mockFiles.GetFileCalls(), 1)
	s.Len(s.mockS3.HeadCalls(), 1)
}

func (s *StoreSuite) TestStatusWhenErrorOnFilesAPIGetCall() {
	s.mockFiles.GetFileFunc = func(ctx context.Context, path string, authToken string) (filesAPI.FileMetaData, error) {
		return filesAPI.FileMetaData{}, errors.New("downstream error")
	}
	store := files.NewStore(s.mockFiles, s.bucket, &config.Config{})
	_, err := store.Status(context.Background(), "invalid-path")
	s.Equal(files.ErrFilesAPINotFound, err)
	s.Len(s.mockFiles.GetFileCalls(), 1)
	s.Len(s.mockS3.HeadCalls(), 0)
}

func (s *StoreSuite) TestStatusStillReturnedIfBucketReadFails() {
	s.mockS3.HeadFunc = func(key string) (*s3.HeadObjectOutput, error) {
		return nil, errors.New("downstream error")
	}
	store := files.NewStore(s.mockFiles, s.bucket, &config.Config{})
	response, err := store.Status(context.Background(), "valid")
	s.NoError(err)
	s.False(response.FileContent.Value)
	s.NotEmpty(response.FileContent.Err)
	s.Len(s.mockFiles.GetFileCalls(), 1)
	s.Len(s.mockS3.HeadCalls(), 1)
}
