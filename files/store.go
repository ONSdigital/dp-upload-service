package files

import (
	"bytes"
	"context"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"

	"github.com/ONSdigital/log.go/v2/log"

	"github.com/ONSdigital/dp-upload-service/encryption"

	s3client "github.com/ONSdigital/dp-s3/v2"

	"github.com/ONSdigital/dp-upload-service/upload"
)

var (
	ErrFilesAPIDuplicateFile    = errors.New("files API already contains a file with this path")
	ErrFileAPICreateInvalidData = errors.New("invalid data sent to Files API")
	ErrUnknownError             = errors.New("unknown error")
	ErrConnectingToFilesApi     = errors.New("could not connect to files API")
	ErrS3Upload                 = errors.New("uploading part failed")
	ErrFileNotFound             = errors.New("file not found")
	ErrFileStateConflict        = errors.New("file was not in the expected state")
	ErrChunkTooSmall            = errors.New("chunk size below minimum 5MB")
	ErrVaultWrite               = errors.New("failed to write to vault")
	ErrVaultRead                = errors.New("failed to read from vault")
)

const (
	vaultKey = "key"
)

type Store struct {
	hostname     string
	s3           upload.S3Clienter
	keyGenerator encryption.GenerateKey
	vault        upload.VaultClienter
	vaultPath    string
}

func NewStore(hostname string, s3 upload.S3Clienter, keyGenerator encryption.GenerateKey, vault upload.VaultClienter, vaultPath string) Store {
	return Store{hostname, s3, keyGenerator, vault, vaultPath}
}

type StoreMetadata struct {
	Path          string  `json:"path"`
	IsPublishable bool    `json:"is_publishable"`
	CollectionId  *string `json:"collection_id,omitempty"`
	Title         string  `json:"title"`
	SizeInBytes   int     `json:"size_in_bytes"`
	Type          string  `json:"type"`
	Licence       string  `json:"licence"`
	LicenceUrl    string  `json:"licence_url"`
}

type Resumable struct {
	FileName     string `schema:"resumableFilename"`
	Type         string `schema:"resumableType"`
	CurrentChunk int64  `schema:"resumableChunkNumber"`
	TotalChunks  int    `schema:"resumableTotalChunks"`
}

type uploadComplete struct {
	State string `json:"state"`
	ETag  string `json:"etag"`
}

type jsonError struct {
	Code        string `json:"code"`
	Description string `json:"description"`
}

type jsonErrors struct {
	Error []jsonError `json:"errors"`
}

func firstChunk(currentChunk int64) bool { return currentChunk == 1 }

func (s Store) UploadFile(ctx context.Context, metadata StoreMetadata, resumable Resumable, content []byte) (bool, error) {

	var encryptionkey []byte
	vaultPath := fmt.Sprintf("%s/%s", s.vaultPath, metadata.Path)
	if firstChunk(resumable.CurrentChunk) {
		encryptionkey = s.keyGenerator()

		if err := s.registerFileUpload(metadata); err != nil {
			log.Error(ctx, "failed to register file metadata with dp-files-api", err, log.Data{"metadata": metadata})
			return false, err
		}

		if err := s.vault.WriteKey(vaultPath, vaultKey, hex.EncodeToString(encryptionkey)); err != nil {
			log.Error(ctx, "failed to write encryption encryptionkey to vault", err, log.Data{"vault-path": vaultPath, "vault-encryptionkey": vaultKey})
			return false, ErrVaultWrite
		}
	} else {
		strKey, err := s.vault.ReadKey(vaultPath, vaultKey)
		if err != nil {
			log.Error(ctx, "failed to read encryption encryptionkey from vault", err, log.Data{"vault-path": vaultPath, "vault-encryptionkey": vaultKey})
			return false, ErrVaultRead
		}

		encryptionkey, _ = hex.DecodeString(strKey)
	}

	upr := s3client.UploadPartRequest{
		UploadKey:   metadata.Path,
		Type:        resumable.Type,
		ChunkNumber: resumable.CurrentChunk,
		TotalChunks: resumable.TotalChunks,
		FileName:    resumable.FileName,
	}

	response, err := s.s3.UploadPartWithPsk(ctx, &upr, content, encryptionkey)
	if err != nil {
		log.Error(ctx, "failed to write chunk to s3", err, log.Data{"s3-upload-part": upr})
		if _, ok := err.(*s3client.ErrChunkTooSmall); ok {
			return false, ErrChunkTooSmall
		}
		return false, ErrS3Upload
	}

	uc := uploadComplete{
		State: "UPLOADED",
		ETag:  strings.Trim(response.Etag, "\""),
	}

	if response.AllPartsUploaded {
		if err := s.markUploadComplete(metadata.Path, uc); err != nil {
			log.Error(ctx, "failed to mark upload complete with dp-files-api", err, log.Data{"upload-complete": uc})
			return true, err
		}
	}

	return response.AllPartsUploaded, nil
}

func (s Store) markUploadComplete(path string, uc uploadComplete) error {
	client := &http.Client{}
	req, err := http.NewRequest(http.MethodPatch, fmt.Sprintf("%s/files/%s", s.hostname, path), jsonEncode(uc))
	if err != nil {
		return err
	}

	req.Header.Set("Content-Type", "application/json")

	resp, err := client.Do(req)
	if err != nil {
		return ErrConnectingToFilesApi
	}

	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		return ErrFileNotFound
	} else if resp.StatusCode == http.StatusConflict {
		return ErrFileStateConflict
	} else if resp.StatusCode != http.StatusOK {
		return ErrUnknownError
	}

	return nil
}

func (s Store) registerFileUpload(metadata StoreMetadata) error {
	log.Info(context.Background(), "Register files API Call", log.Data{"hostname": s.hostname})
	resp, err := http.Post(fmt.Sprintf("%s/files", s.hostname), "application/json", jsonEncode(metadata))
	if err != nil {
		log.Error(context.Background(), "failed to connect to file API", err, log.Data{"hostname": s.hostname})
		return ErrConnectingToFilesApi
	}

	if resp.StatusCode == http.StatusBadRequest {
		var err error

		jsonErrors := jsonErrors{}
		err = json.NewDecoder(resp.Body).Decode(&jsonErrors)
		if err != nil {
			return err
		}

		switch jsonErrors.Error[0].Code {
		case "DuplicateFileError":
			err = ErrFilesAPIDuplicateFile
		case "ValidationError":
			err = ErrFileAPICreateInvalidData
		default:
			err = ErrUnknownError
		}

		return err
	}
	return nil
}

func jsonEncode(data interface{}) *bytes.Buffer {
	b := &bytes.Buffer{}
	json.NewEncoder(b).Encode(data) // nolint // Only fails due to coding error
	return b
}
