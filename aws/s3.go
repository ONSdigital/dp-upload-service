package aws

import (
	"context"
	"github.com/ONSdigital/dp-healthcheck/healthcheck"
	s3client "github.com/ONSdigital/dp-s3/v2"
	"github.com/aws/aws-sdk-go/service/s3"
)

//go:generate moq -out mock/s3.go -pkg mock_aws . S3Clienter

const (
	PathStyle = s3client.PathStyle
)

type S3Clienter interface {
	UploadPart(ctx context.Context, req *s3client.UploadPartRequest, payload []byte) (s3client.MultipartUploadResponse, error)
	UploadPartWithPsk(ctx context.Context, req *s3client.UploadPartRequest, payload []byte, psk []byte) (s3client.MultipartUploadResponse, error)
	CheckPartUploaded(ctx context.Context, req *s3client.UploadPartRequest) (bool, error)
	Checker(ctx context.Context, state *healthcheck.CheckState) error
	Head(key string) (*s3.HeadObjectOutput, error)
}

type Bucket struct {
	region, name string
	S3Clienter
}

func NewBucket(region, name string, client S3Clienter) *Bucket {
	return &Bucket{
		region:     region,
		name:       name,
		S3Clienter: client,
	}
}

func (b *Bucket) GetS3URL(path string) (string, error) {
	// Generate URL from region, bucket and S3 key defined by query
	s3Url, err := s3client.NewURL(b.region, b.name, path)
	if err != nil {
		return "", err
	}
	return s3Url.String(PathStyle)
}
