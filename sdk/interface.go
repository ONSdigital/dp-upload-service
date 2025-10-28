package sdk

import (
	"context"
	"io"

	"github.com/ONSdigital/dp-api-clients-go/v2/health"
	"github.com/ONSdigital/dp-healthcheck/healthcheck"
	"github.com/ONSdigital/dp-upload-service/api"
)

//go:generate moq -out ./mocks/client.go -pkg mocks . Clienter

type Clienter interface {
	Checker(ctx context.Context, check *healthcheck.CheckState) error
	Health() *health.Client
	URL() string

	Upload(ctx context.Context, fileContent io.ReadCloser, metadata api.Metadata, headers Headers) (*UploadResponse, error)
}
