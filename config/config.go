package config

import (
	"time"

	"github.com/kelseyhightower/envconfig"
)

// Config represents service configuration for dp-upload-service
type Config struct {
	BindAddr                   string        `envconfig:"BIND_ADDR"`
	AwsRegion                  string        `envconfig:"AWS_REGION"`
	UploadBucketName           string        `envconfig:"UPLOAD_BUCKET_NAME"`
	EncryptionDisabled         bool          `envconfig:"ENCRYPTION_DISABLED"`
	GracefulShutdownTimeout    time.Duration `envconfig:"GRACEFUL_SHUTDOWN_TIMEOUT"`
	HealthCheckInterval        time.Duration `envconfig:"HEALTHCHECK_INTERVAL"`
	HealthCheckCriticalTimeout time.Duration `envconfig:"HEALTHCHECK_CRITICAL_TIMEOUT"`
	VaultToken                 string        `envconfig:"VAULT_TOKEN"`
	VaultAddress               string        `envconfig:"VAULT_ADDR"`
	VaultPath                  string        `envconfig:"VAULT_PATH"`
	FilesAPIURL                string        `envconfig:"FILE_API_URL"`
}

var cfg *Config

// Get returns the default config with any modifications through environment
// variables
func Get() (*Config, error) {
	if cfg != nil {
		return cfg, nil
	}

	cfg = &Config{
		BindAddr:                   "localhost:25100",
		AwsRegion:                  "eu-west-1",
		UploadBucketName:           "dp-frontend-florence-file-uploads",
		EncryptionDisabled:         false,
		GracefulShutdownTimeout:    5 * time.Second,
		HealthCheckInterval:        30 * time.Second,
		HealthCheckCriticalTimeout: 90 * time.Second,
		VaultToken:                 "",
		VaultAddress:               "http://localhost:8200",
		VaultPath:                  "secret/shared/psk",
		FilesAPIURL: 				"http://localhost:26900",
	}

	return cfg, envconfig.Process("", cfg)
}
