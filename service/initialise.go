package service

import (
	"context"
	"net/http"

	"github.com/ONSdigital/dp-healthcheck/healthcheck"
	dphttp "github.com/ONSdigital/dp-net/v2/http"
	dps3 "github.com/ONSdigital/dp-s3/v2"
	dpaws "github.com/ONSdigital/dp-upload-service/aws"
	"github.com/ONSdigital/dp-upload-service/config"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
)

// ExternalServiceList holds the initialiser and initialisation state of external services.
type ExternalServiceList struct {
	S3Uploaded  bool
	HealthCheck bool
	Init        Initialiser
}

// NewServiceList creates a new service list with the provided initialiser
func NewServiceList(initialiser Initialiser) *ExternalServiceList {
	return &ExternalServiceList{
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

// GetS3Uploaded creates a S3 client and sets the S3Uploaded flag to true
func (e *ExternalServiceList) GetS3Uploaded(ctx context.Context, cfg *config.Config) (dpaws.S3Clienter, error) {
	s3, err := e.Init.DoGetS3Uploaded(ctx, cfg)
	if err != nil {
		return nil, err
	}
	e.S3Uploaded = true
	return s3, nil
}

func (e *ExternalServiceList) GetS3StaticFileUploader(ctx context.Context, cfg *config.Config) (dpaws.S3Clienter, error) {
	return e.Init.DoGetStaticFileS3Uploader(ctx, cfg)
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

// DoGetHTTPServer creates an HTTP Server with the provided bind address and router
func (e *Init) DoGetHTTPServer(bindAddr string, router http.Handler) HTTPServer {
	s := dphttp.NewServer(bindAddr, router)
	s.HandleOSSignals = false
	return s
}

// DoGetS3Uploaded returns a S3Client
func (e *Init) DoGetS3Uploaded(ctx context.Context, cfg *config.Config) (dpaws.S3Clienter, error) {
	return generateS3Client(cfg, cfg.UploadBucketName)
}

// DoGetStaticFileS3Uploader returns a S3Client
func (e *Init) DoGetStaticFileS3Uploader(ctx context.Context, cfg *config.Config) (dpaws.S3Clienter, error) {
	return generateS3Client(cfg, cfg.StaticFilesBucketName)
}

func generateS3Client(cfg *config.Config, bucketName string) (dpaws.S3Clienter, error) {
	if cfg.LocalstackHost != "" {
		s, err := session.NewSession(&aws.Config{
			Endpoint:         aws.String(cfg.LocalstackHost),
			Region:           aws.String(cfg.AwsRegion),
			S3ForcePathStyle: aws.Bool(true),
			Credentials:      credentials.NewStaticCredentials("test", "test", ""),
		})

		if err != nil {
			return nil, err
		}

		return dps3.NewClientWithSession(bucketName, s), nil
	}

	s3Client, err := dps3.NewClient(cfg.AwsRegion, bucketName)
	if err != nil {
		return nil, err
	}
	return s3Client, nil
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
