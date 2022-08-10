package upload

import (
	"context"

	"github.com/ONSdigital/dp-api-clients-go/v2/files"
	"github.com/ONSdigital/dp-healthcheck/healthcheck"
	s3client "github.com/ONSdigital/dp-s3/v2"
)

//go:generate moq -out mock/s3.go -pkg mock . S3Clienter
//go:generate moq -out mock/vault.go -pkg mock . VaultClienter
//go:generate moq -out mock/files.go -pkg mock . FilesClienter

// VaultClienter defines the required method
type VaultClienter interface {
	ReadKey(path, key string) (string, error)
	WriteKey(path, key, value string) error
	Checker(ctx context.Context, state *healthcheck.CheckState) error
}

// S3Clienter defines the required method
type S3Clienter interface {
	UploadPart(ctx context.Context, req *s3client.UploadPartRequest, payload []byte) (s3client.MultipartUploadResponse, error)
	UploadPartWithPsk(ctx context.Context, req *s3client.UploadPartRequest, payload []byte, psk []byte) (s3client.MultipartUploadResponse, error)
	CheckPartUploaded(ctx context.Context, req *s3client.UploadPartRequest) (bool, error)
	Checker(ctx context.Context, state *healthcheck.CheckState) error
}

type FilesClienter interface {
	RegisterFile(ctx context.Context, metadata files.FileMetaData) error
	MarkFileUploaded(ctx context.Context, path string, etag string) error
}
