package service

import (
	"context"
	"net/http"

	"github.com/ONSdigital/dp-healthcheck/healthcheck"
	"github.com/ONSdigital/dp-upload-service/aws"
	"github.com/ONSdigital/dp-upload-service/config"
	"github.com/ONSdigital/dp-upload-service/encryption"
)

//go:generate moq -out mock/initialiser.go -pkg mock_service . Initialiser
//go:generate moq -out mock/server.go -pkg mock_service . HTTPServer
//go:generate moq -out mock/healthCheck.go -pkg mock_service . HealthChecker

// Initialiser defines the methods to initialise external services
type Initialiser interface {
	DoGetHTTPServer(bindAddr string, router http.Handler) HTTPServer
	DoGetVault(ctx context.Context, cfg *config.Config) (encryption.VaultClienter, error)
	DoGetHealthCheck(cfg *config.Config, buildTime, gitCommit, version string) (HealthChecker, error)
	DoGetS3Uploaded(ctx context.Context, cfg *config.Config) (aws.S3Clienter, error)
	DoGetStaticFileS3Uploader(ctx context.Context, cfg *config.Config) (aws.S3Clienter, error)
	DoGetEncryptionKeyGenerator() encryption.GenerateKey
}

// HTTPServer defines the required methods from the HTTP server
type HTTPServer interface {
	ListenAndServe() error
	Shutdown(ctx context.Context) error
}

// HealthChecker defines the required methods from Healthcheck
type HealthChecker interface {
	Handler(w http.ResponseWriter, req *http.Request)
	Start(ctx context.Context)
	Stop()
	AddCheck(name string, checker healthcheck.Checker) (err error)
}
