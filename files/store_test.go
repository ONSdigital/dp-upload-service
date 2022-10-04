package files_test

import (
	"context"
	"errors"
	"testing"

	"github.com/ONSdigital/dp-upload-service/aws"
	"github.com/aws/aws-sdk-go/service/s3"

	"github.com/stretchr/testify/suite"

	filesAPI "github.com/ONSdigital/dp-api-clients-go/v2/files"
	s3client "github.com/ONSdigital/dp-s3/v2"
	mock_aws "github.com/ONSdigital/dp-upload-service/aws/mock"
	"github.com/ONSdigital/dp-upload-service/encryption"
	mock_encryption "github.com/ONSdigital/dp-upload-service/encryption/mock"
	"github.com/ONSdigital/dp-upload-service/files"
	mock_files "github.com/ONSdigital/dp-upload-service/files/mock"
)

var (
	firstResumable = files.Resumable{CurrentChunk: 1}
	lastResumable  = files.Resumable{CurrentChunk: 2}
	vaultPath      = "secret/path/psk"
	content        = []byte("CONTENT")
)

type StoreSuite struct {
	suite.Suite

	mockS3          *mock_aws.S3ClienterMock
	mockFiles       *mock_files.FilesClienterMock
	mockVaultClient *mock_encryption.VaultClienterMock
	vault           *encryption.Vault
	bucket          *aws.Bucket
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
			return &s3.HeadObjectOutput{ContentLength: &size}, nil
		},
		UploadPartWithPskFunc: func(ctx context.Context, req *s3client.UploadPartRequest, payload []byte, psk []byte) (s3client.MultipartUploadResponse, error) {
			return s3client.MultipartUploadResponse{Etag: "123456789", AllPartsUploaded: true}, nil
		},
	}
	s.bucket = aws.NewBucket("region", "name", s.mockS3)

	s.mockVaultClient = &mock_encryption.VaultClienterMock{
		ReadKeyFunc: func(path string, key string) (string, error) {
			return "123456789123456789", nil
		},
		WriteKeyFunc: func(path string, key string, value string) error {
			return nil
		},
	}
	fakeKeyGenerator := func() ([]byte, error) { return []byte("testing"), nil }
	s.vault = encryption.NewVault(fakeKeyGenerator, s.mockVaultClient, vaultPath)
}

// Upload
func (s *StoreSuite) TestFileUploadIsRegisteredWithFilesApi() {
	store := files.NewStore(s.mockFiles, s.bucket, s.vault)

	_, err := store.UploadFile(context.Background(), filesAPI.FileMetaData{}, firstResumable, content)
	s.NoError(err)
	s.Len(s.mockFiles.RegisterFileCalls(), 1)
}

func (s *StoreSuite) TestFileRegistrationFailsWithFilesApi() {
	expectedError := errors.New("registration error")
	s.mockFiles.RegisterFileFunc = func(ctx context.Context, metadata filesAPI.FileMetaData) error {
		return expectedError
	}

	store := files.NewStore(s.mockFiles, s.bucket, s.vault)

	_, err := store.UploadFile(context.Background(), filesAPI.FileMetaData{}, firstResumable, content)
	s.Equal(expectedError, err)
}

func (s *StoreSuite) TestErrorGeneratingEncryptionKey() {
	badKeyGenerator := func() ([]byte, error) { return nil, errors.New("no key available") }
	s.vault = encryption.NewVault(badKeyGenerator, s.mockVaultClient, vaultPath)

	store := files.NewStore(s.mockFiles, s.bucket, s.vault)

	_, err := store.UploadFile(context.Background(), filesAPI.FileMetaData{}, firstResumable, []byte("CONTENT"))

	s.Equal(encryption.ErrKeyGeneration, err)
}

func (s *StoreSuite) TestErrorStoringEncryptionKeyInVault() {
	s.mockVaultClient.WriteKeyFunc = func(path string, key string, value string) error {
		return errors.New("failed writing to vault")
	}

	store := files.NewStore(s.mockFiles, s.bucket, s.vault)

	_, err := store.UploadFile(context.Background(), filesAPI.FileMetaData{}, firstResumable, content)

	s.Equal(encryption.ErrVaultWrite, err)
}

func (s *StoreSuite) TestErrorReadingEncryptionKeyFromValue() {
	s.mockVaultClient.ReadKeyFunc = func(path string, key string) (string, error) {
		return "", errors.New("failed writing to vault")
	}

	store := files.NewStore(s.mockFiles, s.bucket, s.vault)

	_, err := store.UploadFile(context.Background(), filesAPI.FileMetaData{}, lastResumable, content)

	s.Equal(encryption.ErrVaultRead, err)
}

func (s *StoreSuite) TestEncryptionKeyContainsNonHexCharacters() {
	s.mockVaultClient.ReadKeyFunc = func(path string, key string) (string, error) {
		return "NON HEX CHARACTERS", nil
	}

	store := files.NewStore(s.mockFiles, s.bucket, s.vault)

	_, err := store.UploadFile(context.Background(), filesAPI.FileMetaData{}, lastResumable, content)

	s.Equal(encryption.ErrInvalidEncryptionKey, err)
}

func (s StoreSuite) TestUploadPartReturnsAnError() {
	s.mockS3.UploadPartWithPskFunc = func(ctx context.Context, req *s3client.UploadPartRequest, payload []byte, psk []byte) (s3client.MultipartUploadResponse, error) {
		return s3client.MultipartUploadResponse{}, s3client.NewError(errors.New("broken"), nil)
	}

	store := files.NewStore(s.mockFiles, s.bucket, s.vault)

	_, err := store.UploadFile(context.Background(), filesAPI.FileMetaData{}, lastResumable, content)
	s.Equal(files.ErrS3Upload, err)
}

func (s StoreSuite) TestUploadChunkTooSmallReturnsErrChuckTooSmall() {
	s.mockS3.UploadPartWithPskFunc = func(ctx context.Context, req *s3client.UploadPartRequest, payload []byte, psk []byte) (s3client.MultipartUploadResponse, error) {
		return s3client.MultipartUploadResponse{}, s3client.NewChunkTooSmallError(errors.New("chunk size below minimum 5MB"), nil)
	}

	store := files.NewStore(s.mockFiles, s.bucket, s.vault)

	_, err := store.UploadFile(context.Background(), filesAPI.FileMetaData{}, lastResumable, content)
	s.Equal(files.ErrChunkTooSmall, err)
}

func (s StoreSuite) TestErrorMarkingAsUploaded() {
	expectedError := errors.New("marking error")
	s.mockFiles.MarkFileUploadedFunc = func(ctx context.Context, path string, etag string) error {
		return expectedError
	}
	store := files.NewStore(s.mockFiles, s.bucket, s.vault)

	_, err := store.UploadFile(context.Background(), filesAPI.FileMetaData{}, lastResumable, content)
	s.Equal(expectedError, err)
}

//Status
func (s *StoreSuite) TestStatusHappyPath() {
	store := files.NewStore(s.mockFiles, s.bucket, s.vault)
	response, err := store.Status(context.Background(), "valid")
	s.NoError(err)
	s.True(response.EncryptionKey.Value)
	s.True(response.FileContent.Value)
	s.Len(s.mockFiles.GetFileCalls(), 1)
	s.Len(s.mockVaultClient.ReadKeyCalls(), 1)
	s.Len(s.mockS3.HeadCalls(), 1)
}

func (s *StoreSuite) TestStatusWhenErrorOnFilesAPIGetCall() {
	s.mockFiles.GetFileFunc = func(ctx context.Context, path string, authToken string) (filesAPI.FileMetaData, error) {
		return filesAPI.FileMetaData{}, errors.New("downstream error")
	}
	store := files.NewStore(s.mockFiles, s.bucket, s.vault)
	_, err := store.Status(context.Background(), "invalid-path")
	s.Equal(files.ErrFilesAPINotFound, err)
	s.Len(s.mockFiles.GetFileCalls(), 1)
	s.Len(s.mockVaultClient.ReadKeyCalls(), 0)
	s.Len(s.mockS3.HeadCalls(), 0)
}

func (s *StoreSuite) TestStatusStillReturnedIfVaultAndBucketReadFails() {
	s.mockS3.HeadFunc = func(key string) (*s3.HeadObjectOutput, error) {
		return nil, errors.New("downstream error")
	}
	s.mockVaultClient.ReadKeyFunc = func(path string, key string) (string, error) {
		return "", errors.New("downstream error")
	}
	store := files.NewStore(s.mockFiles, s.bucket, s.vault)
	response, err := store.Status(context.Background(), "valid")
	s.NoError(err)
	s.False(response.EncryptionKey.Value)
	s.NotEmpty(response.EncryptionKey.Err)
	s.False(response.FileContent.Value)
	s.NotEmpty(response.FileContent.Err)
	s.Len(s.mockFiles.GetFileCalls(), 1)
	s.Len(s.mockVaultClient.ReadKeyCalls(), 1)
	s.Len(s.mockS3.HeadCalls(), 1)
}
