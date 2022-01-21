package service

import (
	"context"
	"net/http"

	"github.com/ONSdigital/dp-upload-service/encryption"

	"github.com/ONSdigital/dp-upload-service/config"
	"github.com/ONSdigital/dp-upload-service/upload"

	"github.com/ONSdigital/dp-healthcheck/healthcheck"
	dphttp "github.com/ONSdigital/dp-net/http"
	dps3 "github.com/ONSdigital/dp-s3/v2"
	dpvault "github.com/ONSdigital/dp-vault"
)

// ExternalServiceList holds the initialiser and initialisation state of external services.
type ExternalServiceList struct {
	Vault       bool
	S3Uploaded  bool
	HealthCheck bool
	Init        Initialiser
}

// NewServiceList creates a new service list with the provided initialiser
func NewServiceList(initialiser Initialiser) *ExternalServiceList {
	return &ExternalServiceList{
		Vault:       false,
		S3Uploaded:  false,
		HealthCheck: false,
		Init:        initialiser,
	}
}

// Init implements the Initialiser interface to initialise dependencies
type Init struct{}

// GetHTTPServer creates an http server
func (e *ExternalServiceList) GetHTTPServer(bindAddr string, router http.Handler) HTTPServer {
	s := e.Init.DoGetHTTPServer(bindAddr, router)
	return s
}

// GetVault creates a Vault client and sets the Vault flag to true
func (e *ExternalServiceList) GetVault(ctx context.Context, cfg *config.Config) (upload.VaultClienter, error) {
	vault, err := e.Init.DoGetVault(ctx, cfg)
	if err != nil {
		return nil, err
	}
	e.Vault = true
	return vault, nil
}

// GetS3Uploaded creates a S3 client and sets the S3Uploaded flag to true
func (e *ExternalServiceList) GetS3Uploaded(ctx context.Context, cfg *config.Config) (upload.S3Clienter, error) {
	s3, err := e.Init.DoGetS3Uploaded(ctx, cfg)
	if err != nil {
		return nil, err
	}
	e.S3Uploaded = true
	return s3, nil
}

// GetHealthCheck creates a healthcheck with versionInfo and sets teh HealthCheck flag to true
func (e *ExternalServiceList) GetHealthCheck(cfg *config.Config, buildTime, gitCommit, version string) (HealthChecker, error) {
	hc, err := e.Init.DoGetHealthCheck(cfg, buildTime, gitCommit, version)
	if err != nil {
		return nil, err
	}
	e.HealthCheck = true
	return hc, nil
}

func (e *ExternalServiceList) GetEncryptionKeyGenerator() encryption.GenerateKey {
	return e.Init.DoGetEncryptionKeyGenerator()
}

// DoGetHTTPServer creates an HTTP Server with the provided bind address and router
func (e *Init) DoGetHTTPServer(bindAddr string, router http.Handler) HTTPServer {
	s := dphttp.NewServer(bindAddr, router)
	s.HandleOSSignals = false
	return s
}

// DoGetS3Uploaded returns a S3Client
func (e *Init) DoGetS3Uploaded(ctx context.Context, cfg *config.Config) (upload.S3Clienter, error) {
	s3Client, err := dps3.NewClient(cfg.AwsRegion, cfg.UploadBucketName)
	if err != nil {
		return nil, err
	}
	return s3Client, nil
}

// DoGetVault returns a VaultClient unless encryption is disabled
//
// If cfg.EncryptionDisabled is true then the function returns nil
func (e *Init) DoGetVault(ctx context.Context, cfg *config.Config) (upload.VaultClienter, error) {
	if cfg.EncryptionDisabled {
		return nil, nil
	}
	vault, err := dpvault.CreateClient(cfg.VaultToken, cfg.VaultAddress, 3)
	if err != nil {
		return nil, err
	}
	return vault, nil
}

// DoGetHealthCheck creates a healthcheck with versionInfo
func (e *Init) DoGetHealthCheck(cfg *config.Config, buildTime, gitCommit, version string) (HealthChecker, error) {
	versionInfo, err := healthcheck.NewVersionInfo(buildTime, gitCommit, version)
	if err != nil {
		return nil, err
	}
	hc := healthcheck.New(versionInfo, cfg.HealthCheckCriticalTimeout, cfg.HealthCheckInterval)
	return &hc, nil
}

func (e *Init) DoGetEncryptionKeyGenerator() encryption.GenerateKey {
	return encryption.CreateKey
}
