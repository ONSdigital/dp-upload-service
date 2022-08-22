package files_test

import (
	"context"
	"errors"
	"testing"

	filesAPI "github.com/ONSdigital/dp-api-clients-go/v2/files"
	s3client "github.com/ONSdigital/dp-s3/v2"
	"github.com/ONSdigital/dp-upload-service/encryption"
	"github.com/ONSdigital/dp-upload-service/encryption/mock"
	"github.com/ONSdigital/dp-upload-service/files"
	"github.com/ONSdigital/dp-upload-service/upload/mock"
	"github.com/stretchr/testify/suite"
)

var (
	firstResumable = files.Resumable{CurrentChunk: 1}
	lastResumable  = files.Resumable{CurrentChunk: 2}
	vaultPath      = "secret/path/psk"
)

type StoreSuite struct {
	suite.Suite

	mockS3           *upload_mock.S3ClienterMock
	mockFiles        *upload_mock.FilesClienterMock
	fakeKeyGenerator encryption.GenerateKey
	mockVaultClient  *encryption_mock.VaultClienterMock
	mockVault        *encryption.Vault
}

func TestStore(t *testing.T) {
	suite.Run(t, new(StoreSuite))
}

// beforeEach
func (s *StoreSuite) SetupTest() {
	s.mockFiles = &upload_mock.FilesClienterMock{
		RegisterFileFunc: func(ctx context.Context, metadata filesAPI.FileMetaData) error {
			return nil
		},
		MarkFileUploadedFunc: func(ctx context.Context, path string, etag string) error {
			return nil
		},
	}
	s.mockS3 = &upload_mock.S3ClienterMock{
		UploadPartWithPskFunc: func(ctx context.Context, req *s3client.UploadPartRequest, payload []byte, psk []byte) (s3client.MultipartUploadResponse, error) {
			return s3client.MultipartUploadResponse{Etag: "123456789", AllPartsUploaded: true}, nil
		},
	}
	s.mockVaultClient = &encryption_mock.VaultClienterMock{
		ReadKeyFunc: func(path string, key string) (string, error) {
			return "123456789123456789", nil
		},
		WriteKeyFunc: func(path string, key string, value string) error {
			return nil
		},
	}
	fakeKeyGenerator := func() []byte { return []byte("testing") }
	s.mockVault = encryption.NewVault(fakeKeyGenerator, s.mockVaultClient, vaultPath)
}

func (s *StoreSuite) TestFileUploadIsRegisteredWithFilesApi() {
	store := files.NewStore(s.mockFiles, s.mockS3, s.mockVault)

	_, err := store.UploadFile(context.Background(), filesAPI.FileMetaData{}, firstResumable, []byte("CONTENT"))
	s.NoError(err)
	s.Len(s.mockFiles.RegisterFileCalls(), 1)
}

func (s *StoreSuite) TestFileRegistrationFailsWithFilesApi() {

	expectedError := errors.New("registration error")
	s.mockFiles.RegisterFileFunc = func(ctx context.Context, metadata filesAPI.FileMetaData) error {
		return expectedError
	}

	store := files.NewStore(s.mockFiles, s.mockS3, s.mockVault)

	_, err := store.UploadFile(context.Background(), filesAPI.FileMetaData{}, firstResumable, []byte("CONTENT"))
	s.Equal(expectedError, err)
}

func (s *StoreSuite) TestErrorStoringEncryptionKeyInVault() {
	s.mockVaultClient.WriteKeyFunc = func(path string, key string, value string) error {
		return errors.New("failed writing to vault")
	}

	store := files.NewStore(s.mockFiles, s.mockS3, s.mockVault)

	_, err := store.UploadFile(context.Background(), filesAPI.FileMetaData{}, firstResumable, []byte("CONTENT"))

	s.Equal(encryption.ErrVaultWrite, err)
}

func (s *StoreSuite) TestErrorReadingEncryptionKeyFromValue() {
	s.mockVaultClient.ReadKeyFunc = func(path string, key string) (string, error) {
		return "", errors.New("failed writing to vault")
	}

	store := files.NewStore(s.mockFiles, s.mockS3, s.mockVault)

	_, err := store.UploadFile(context.Background(), filesAPI.FileMetaData{}, lastResumable, []byte("CONTENT"))

	s.Equal(encryption.ErrVaultRead, err)
}

func (s *StoreSuite) TestEncryptionKeyContainsNonHexCharacters() {
	s.mockVaultClient.ReadKeyFunc = func(path string, key string) (string, error) {
		return "NON HEX CHARACTERS", nil
	}

	store := files.NewStore(s.mockFiles, s.mockS3, s.mockVault)

	_, err := store.UploadFile(context.Background(), filesAPI.FileMetaData{}, lastResumable, []byte("CONTENT"))

	s.Equal(encryption.ErrInvalidEncryptionKey, err)
}

func (s StoreSuite) TestUploadPartReturnsAnError() {
	s.mockS3.UploadPartWithPskFunc = func(ctx context.Context, req *s3client.UploadPartRequest, payload []byte, psk []byte) (s3client.MultipartUploadResponse, error) {
		return s3client.MultipartUploadResponse{}, s3client.NewError(errors.New("broken"), nil)
	}

	store := files.NewStore(s.mockFiles, s.mockS3, s.mockVault)

	_, err := store.UploadFile(context.Background(), filesAPI.FileMetaData{}, lastResumable, []byte("CONTENT"))
	s.Equal(files.ErrS3Upload, err)
}

func (s StoreSuite) TestUploadChunkTooSmallReturnsErrChuckTooSmall() {
	s.mockS3.UploadPartWithPskFunc = func(ctx context.Context, req *s3client.UploadPartRequest, payload []byte, psk []byte) (s3client.MultipartUploadResponse, error) {
		return s3client.MultipartUploadResponse{}, s3client.NewChunkTooSmallError(errors.New("chunk size below minimum 5MB"), nil)
	}

	store := files.NewStore(s.mockFiles, s.mockS3, s.mockVault)

	_, err := store.UploadFile(context.Background(), filesAPI.FileMetaData{}, lastResumable, []byte("CONTENT"))
	s.Equal(files.ErrChunkTooSmall, err)
}

func (s StoreSuite) TestErrorMarkingAsUploaded() {
	expectedError := errors.New("marking error")
	s.mockFiles.MarkFileUploadedFunc = func(ctx context.Context, path string, etag string) error {
		return expectedError
	}
	store := files.NewStore(s.mockFiles, s.mockS3, s.mockVault)

	_, err := store.UploadFile(context.Background(), filesAPI.FileMetaData{}, lastResumable, []byte("CONTENT"))
	s.Equal(expectedError, err)
}
