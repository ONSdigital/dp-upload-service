package files_test

import (
	"context"
	"errors"
	encryption_mock "github.com/ONSdigital/dp-upload-service/encryption/mock"
	"net/http"
	"testing"

	"github.com/ONSdigital/dp-upload-service/encryption"

	s3client "github.com/ONSdigital/dp-s3/v2"

	"github.com/ONSdigital/dp-upload-service/files"
	"github.com/ONSdigital/dp-upload-service/upload/mock"

	"github.com/maxcnunes/httpfake"
	"github.com/stretchr/testify/suite"
)

var (
	firstResumable = files.Resumable{CurrentChunk: 1}
	lastResumable  = files.Resumable{CurrentChunk: 2}
	vaultPath      = "secret/path/psk"
)

type StoreSuite struct {
	suite.Suite

	fakeFilesApi    *httpfake.HTTPFake
	mockS3          *upload_mock.S3ClienterMock
	mockVaultClient *encryption_mock.VaultClienterMock
	mockVault       *encryption.Vault
}

func TestStore(t *testing.T) {
	suite.Run(t, new(StoreSuite))
}

// beforeEach
func (s *StoreSuite) SetupTest() {
	s.fakeFilesApi = httpfake.New(httpfake.WithTesting(s.T()))
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

// afterEach
func (s *StoreSuite) TearDownTest() {
	s.fakeFilesApi.Close()
}

func (s *StoreSuite) TestFileUploadIsRegisteredWithFilesApi() {
	s.fakeFilesApi.NewHandler().Post("/files").Reply(http.StatusCreated)
	s.fakeFilesApi.NewHandler().Patch("/files/").Reply(http.StatusOK)

	store := files.NewStore(s.fakeFilesApi.ResolveURL(""), s.mockS3, s.mockVault)

	_, err := store.UploadFile(context.Background(), files.StoreMetadata{}, firstResumable, []byte("CONTENT"))
	s.NoError(err)
}

func (s *StoreSuite) TestFileAlreadyRegisteredWithFilesApi() {
	s.fakeFilesApi.NewHandler().
		Post("/files").
		Reply(http.StatusBadRequest).
		Body([]byte(`{"errors": [{"code": "DuplicateFileError", "description": "file already exists"}]}`))

	store := files.NewStore(s.fakeFilesApi.ResolveURL(""), s.mockS3, s.mockVault)

	_, err := store.UploadFile(context.Background(), files.StoreMetadata{}, firstResumable, []byte("CONTENT"))
	s.Equal(files.ErrFilesAPIDuplicateFile, err)
}

func (s *StoreSuite) TestFileRegisteredWithInvalidContent() {
	s.fakeFilesApi.NewHandler().
		Post("/files").
		Reply(http.StatusBadRequest).
		Body([]byte(`{"errors": [{"code": "ValidationError", "description": "fields were invalid"}]}`))

	store := files.NewStore(s.fakeFilesApi.ResolveURL(""), s.mockS3, s.mockVault)

	_, err := store.UploadFile(context.Background(), files.StoreMetadata{}, firstResumable, []byte("CONTENT"))
	s.Equal(files.ErrFileAPICreateInvalidData, err)
}

func (s *StoreSuite) TestFileRegisterReturnsUnknownError() {
	s.fakeFilesApi.NewHandler().
		Post("/files").
		Reply(http.StatusBadRequest).
		Body([]byte(`{"errors": [{"code": "SpecialError", "description": "fields were invalid"}]}`))

	store := files.NewStore(s.fakeFilesApi.ResolveURL(""), s.mockS3, s.mockVault)

	_, err := store.UploadFile(context.Background(), files.StoreMetadata{}, firstResumable, []byte("CONTENT"))
	s.Equal(files.ErrUnknownError, err)
}

func (s *StoreSuite) TestFileRegisterReturnsInternalServerError() {
	s.fakeFilesApi.NewHandler().
		Post("/files").
		Reply(http.StatusInternalServerError).
		Body([]byte(`{"errors": [{"code": "BROKEN", "description": "nothing is working"}]}`))

	store := files.NewStore(s.fakeFilesApi.ResolveURL(""), s.mockS3, s.mockVault)

	_, err := store.UploadFile(context.Background(), files.StoreMetadata{}, firstResumable, []byte("CONTENT"))
	s.Equal(files.ErrFilesServer, err)
}

func (s *StoreSuite) TestFileRegisterReturnsUnexpectedError() {
	s.fakeFilesApi.NewHandler().
		Post("/files").
		Reply(http.StatusTeapot).
		Body([]byte("we're all mad down here"))

	store := files.NewStore(s.fakeFilesApi.ResolveURL(""), s.mockS3, s.mockVault)

	_, err := store.UploadFile(context.Background(), files.StoreMetadata{}, firstResumable, []byte("CONTENT"))
	s.Equal(files.ErrUnknownError, err)
}

func (s *StoreSuite) TestFileRegisterReturnsForbiddenError() {
	s.fakeFilesApi.NewHandler().
		Post("/files").
		Reply(http.StatusForbidden).
		Body([]byte(`{"errors": [{"code": "Forbidden", "description": "unauthorised"}]}`))

	store := files.NewStore(s.fakeFilesApi.ResolveURL(""), s.mockS3, s.mockVault)

	_, err := store.UploadFile(context.Background(), files.StoreMetadata{}, firstResumable, []byte("CONTENT"))
	s.Equal(files.ErrFilesUnauthorised, err)
}

func (s *StoreSuite) TestFileRegisterReturnsMalformedJSON() {
	s.fakeFilesApi.NewHandler().
		Post("/files").
		Reply(http.StatusBadRequest).
		Body([]byte(`<json>Error occurred</json>`))

	store := files.NewStore(s.fakeFilesApi.ResolveURL(""), s.mockS3, s.mockVault)

	_, err := store.UploadFile(context.Background(), files.StoreMetadata{}, firstResumable, []byte("CONTENT"))

	s.Error(err)
}

func (s *StoreSuite) TestErrorConnectingToRegisterFiles() {
	store := files.NewStore("does.not.work", s.mockS3, s.mockVault)

	_, err := store.UploadFile(context.Background(), files.StoreMetadata{}, firstResumable, []byte("CONTENT"))
	s.Equal(files.ErrConnectingToFilesApi, err)
}

func (s *StoreSuite) TestErrorStoringEncryptionKeyInVault() {
	s.fakeFilesApi.NewHandler().Post("/files").Reply(http.StatusCreated)
	s.mockVaultClient.WriteKeyFunc = func(path string, key string, value string) error {
		return errors.New("failed writing to vault")
	}

	store := files.NewStore(s.fakeFilesApi.ResolveURL(""), s.mockS3, s.mockVault)

	_, err := store.UploadFile(context.Background(), files.StoreMetadata{}, firstResumable, []byte("CONTENT"))

	s.Equal(encryption.ErrVaultWrite, err)
}

func (s *StoreSuite) TestErrorReadingEncryptionKeyFromValue() {
	s.mockVaultClient.ReadKeyFunc = func(path string, key string) (string, error) {
		return "", errors.New("failed writing to vault")
	}

	store := files.NewStore(s.fakeFilesApi.ResolveURL(""), s.mockS3, s.mockVault)

	_, err := store.UploadFile(context.Background(), files.StoreMetadata{}, lastResumable, []byte("CONTENT"))

	s.Equal(encryption.ErrVaultRead, err)
}

func (s *StoreSuite) TestEncryptionKeyContainsNonHexCharacters() {
	s.mockVaultClient.ReadKeyFunc = func(path string, key string) (string, error) {
		return "NON HEX CHARACTERS", nil
	}

	store := files.NewStore(s.fakeFilesApi.ResolveURL(""), s.mockS3, s.mockVault)

	_, err := store.UploadFile(context.Background(), files.StoreMetadata{}, lastResumable, []byte("CONTENT"))

	s.Equal(encryption.ErrInvalidEncryptionKey, err)
}

func (s StoreSuite) TestUploadPartReturnsAnError() {
	s.mockS3.UploadPartWithPskFunc = func(ctx context.Context, req *s3client.UploadPartRequest, payload []byte, psk []byte) (s3client.MultipartUploadResponse, error) {
		return s3client.MultipartUploadResponse{}, s3client.NewError(errors.New("broken"), nil)
	}

	store := files.NewStore(s.fakeFilesApi.ResolveURL(""), s.mockS3, s.mockVault)

	_, err := store.UploadFile(context.Background(), files.StoreMetadata{}, lastResumable, []byte("CONTENT"))
	s.Equal(files.ErrS3Upload, err)
}

func (s StoreSuite) TestUploadChunkTooSmallReturnsErrChuckTooSmall() {
	s.mockS3.UploadPartWithPskFunc = func(ctx context.Context, req *s3client.UploadPartRequest, payload []byte, psk []byte) (s3client.MultipartUploadResponse, error) {
		return s3client.MultipartUploadResponse{}, s3client.NewChunkTooSmallError(errors.New("chunk size below minimum 5MB"), nil)
	}

	store := files.NewStore(s.fakeFilesApi.ResolveURL(""), s.mockS3, s.mockVault)

	_, err := store.UploadFile(context.Background(), files.StoreMetadata{}, lastResumable, []byte("CONTENT"))
	s.Equal(files.ErrChunkTooSmall, err)
}

func (s StoreSuite) TestFileNotFoundWhenMarkedAsUploaded() {
	s.fakeFilesApi.NewHandler().Patch("/files/").Reply(http.StatusNotFound)

	store := files.NewStore(s.fakeFilesApi.ResolveURL(""), s.mockS3, s.mockVault)

	_, err := store.UploadFile(context.Background(), files.StoreMetadata{}, lastResumable, []byte("CONTENT"))
	s.Equal(files.ErrFileNotFound, err)
}

func (s StoreSuite) TestReturnsConflictWhenFileInUnexpectedState() {
	s.fakeFilesApi.NewHandler().Patch("/files/").Reply(http.StatusConflict)

	store := files.NewStore(s.fakeFilesApi.ResolveURL(""), s.mockS3, s.mockVault)

	_, err := store.UploadFile(context.Background(), files.StoreMetadata{}, lastResumable, []byte("CONTENT"))
	s.Equal(files.ErrFileStateConflict, err)
}

func (s StoreSuite) TestReturnsFileServerErrorWhenFileInteralServerError() {
	s.fakeFilesApi.NewHandler().Patch("/files/").Reply(http.StatusInternalServerError)

	store := files.NewStore(s.fakeFilesApi.ResolveURL(""), s.mockS3, s.mockVault)

	_, err := store.UploadFile(context.Background(), files.StoreMetadata{}, lastResumable, []byte("CONTENT"))
	s.Equal(files.ErrFilesServer, err)
}

func (s StoreSuite) TestReturnsAuthenticationErrorWhenFileForbidden() {
	s.fakeFilesApi.NewHandler().Patch("/files/").Reply(http.StatusForbidden)

	store := files.NewStore(s.fakeFilesApi.ResolveURL(""), s.mockS3, s.mockVault)

	_, err := store.UploadFile(context.Background(), files.StoreMetadata{}, lastResumable, []byte("CONTENT"))
	s.Equal(files.ErrFilesUnauthorised, err)
}

func (s StoreSuite) TestUploadCompleteUnknownError() {
	s.fakeFilesApi.NewHandler().Patch("/files/").Reply(http.StatusTeapot)

	store := files.NewStore(s.fakeFilesApi.ResolveURL(""), s.mockS3, s.mockVault)

	_, err := store.UploadFile(context.Background(), files.StoreMetadata{}, lastResumable, []byte("CONTENT"))
	s.Equal(files.ErrUnknownError, err)
}

func (s *StoreSuite) TestErrorConnectingToUploadComplete() {
	store := files.NewStore("does.not.work", s.mockS3, s.mockVault)

	_, err := store.UploadFile(context.Background(), files.StoreMetadata{}, lastResumable, []byte("CONTENT"))
	s.Equal(files.ErrConnectingToFilesApi, err)
}
