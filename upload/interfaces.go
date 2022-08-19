package upload

import (
	"context"

	"github.com/ONSdigital/dp-healthcheck/healthcheck"
	s3client "github.com/ONSdigital/dp-s3/v2"
)

//go:generate moq -out mock/s3.go -pkg upload_mock . S3Clienter

// S3Clienter defines the required method
type S3Clienter interface {
	UploadPart(ctx context.Context, req *s3client.UploadPartRequest, payload []byte) (s3client.MultipartUploadResponse, error)
	UploadPartWithPsk(ctx context.Context, req *s3client.UploadPartRequest, payload []byte, psk []byte) (s3client.MultipartUploadResponse, error)
	CheckPartUploaded(ctx context.Context, req *s3client.UploadPartRequest) (bool, error)
	Checker(ctx context.Context, state *healthcheck.CheckState) error
}
