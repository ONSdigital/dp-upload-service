package api

import (
	"context"

	"github.com/ONSdigital/dp-healthcheck/healthcheck"
	s3client "github.com/ONSdigital/dp-s3"
)

//go:generate moq -out mock_s3.go  . S3Clienter
//go:generate moq -out mock_vault.go . VaultClienter

//VaultClienter defines the required method
type VaultClienter interface {
	ReadKey(path, key string) (string, error)
	WriteKey(path, key, value string) error
	Checker(ctx context.Context, state *healthcheck.CheckState) error
}

//S3Clienter defines the required method
type S3Clienter interface {
	UploadPart(ctx context.Context, req *s3client.UploadPartRequest, payload []byte) error
	UploadPartWithPsk(ctx context.Context, req *s3client.UploadPartRequest, payload []byte, psk []byte) error
	CheckPartUploaded(ctx context.Context, req *s3client.UploadPartRequest) (bool, error)
	Checker(ctx context.Context, state *healthcheck.CheckState) error
}
