package steps

import (
	"context"
	"net/http"
	"time"

	dpaws "github.com/ONSdigital/dp-upload-service/aws"

	"github.com/ONSdigital/dp-healthcheck/healthcheck"
	dphttp "github.com/ONSdigital/dp-net/v3/http"
	s3client "github.com/ONSdigital/dp-s3/v3"
	"github.com/ONSdigital/dp-upload-service/config"
	"github.com/ONSdigital/dp-upload-service/service"
	"github.com/aws/aws-sdk-go-v2/aws"
	awsConfig "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
)

type external struct {
	Server *dphttp.Server
}

func (e external) DoGetHTTPServer(bindAddr string, router http.Handler) service.HTTPServer {
	e.Server.Server.Addr = bindAddr
	e.Server.Server.Handler = router

	return e.Server
}

func (e external) DoGetHealthCheck(cfg *config.Config, buildTime, gitCommit, version string) (service.HealthChecker, error) {
	info, _ := healthcheck.NewVersionInfo("1", "1", "1")
	check := healthcheck.New(info, 5*time.Second, 500*time.Millisecond)
	return &check, nil
}

func (e external) DoGetS3Uploaded(ctx context.Context, cfg *config.Config) (dpaws.S3Clienter, error) {
	return generateS3Client(ctx, cfg, cfg.UploadBucketName)
}

func (e external) DoGetStaticFileS3Uploader(ctx context.Context, cfg *config.Config) (dpaws.S3Clienter, error) {
	return generateS3Client(ctx, cfg, cfg.StaticFilesEncryptedBucketName)
}

func generateS3Client(ctx context.Context, cfg *config.Config, bucketName string) (dpaws.S3Clienter, error) {

	var AWSConfig aws.Config
	var err error
	var client *s3client.Client

	AWSConfig, err = awsConfig.LoadDefaultConfig(
		ctx,
		awsConfig.WithRegion(cfg.AwsRegion),
		awsConfig.WithCredentialsProvider(credentials.NewStaticCredentialsProvider("test", "test", "")),
	)

	if err != nil {
		return nil, err
	}

	client = s3client.NewClientWithConfig(bucketName, AWSConfig, func(options *s3.Options) {
		options.BaseEndpoint = aws.String(localStackHost)
		options.UsePathStyle = true
	})

	return client, nil
}
