package config

import (
	"time"

	"github.com/ONSdigital/dp-net/v2/request"

	"github.com/kelseyhightower/envconfig"
)

const (
	AuthContextKey ContextKey = request.AuthHeaderKey
)

type ContextKey string

// Config represents service configuration for dp-upload-service
type Config struct {
	BindAddr                   string        `envconfig:"BIND_ADDR"`
	AwsRegion                  string        `envconfig:"AWS_REGION"`
	LocalstackHost             string        `envconfig:"LOCALSTACK_HOST"`
	UploadBucketName           string        `envconfig:"UPLOAD_BUCKET_NAME"`
	StaticFilesEncryptedBucketName      string        `envconfig:"STATIC_FILES_ENCRYPTED_BUCKET_NAME"`
	GracefulShutdownTimeout    time.Duration `envconfig:"GRACEFUL_SHUTDOWN_TIMEOUT"`
	HealthCheckInterval        time.Duration `envconfig:"HEALTHCHECK_INTERVAL"`
	HealthCheckCriticalTimeout time.Duration `envconfig:"HEALTHCHECK_CRITICAL_TIMEOUT"`
	FilesAPIURL                string        `envconfig:"FILES_API_URL"`
	ServiceAuthToken           string        `envconfig:"SERVICE_AUTH_TOKEN"         json:"-"`
}

// Get returns the default config with any modifications through environment
// variables
func Get() (*Config, error) {
	cfg := &Config{
		BindAddr:                   ":25100",
		AwsRegion:                  "eu-west-2",
		UploadBucketName:           "deprecated",
		StaticFilesEncryptedBucketName:      "testing",
		GracefulShutdownTimeout:    5 * time.Second,
		HealthCheckInterval:        30 * time.Second,
		HealthCheckCriticalTimeout: 90 * time.Second,
		FilesAPIURL:                "http://localhost:26900", //401 via api-router [http://localhost:23200/v1]
		ServiceAuthToken:           "c60198e9-1864-4b68-ad0b-1e858e5b46a4",
	}

	return cfg, envconfig.Process("", cfg)
}
