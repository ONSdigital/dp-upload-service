package steps

import (
	"context"
	"fmt"
	dpaws "github.com/ONSdigital/dp-upload-service/aws"
	"net/http"
	"time"

	"github.com/ONSdigital/dp-healthcheck/healthcheck"
	dphttp "github.com/ONSdigital/dp-net/v2/http"
	s3client "github.com/ONSdigital/dp-s3/v2"
	"github.com/ONSdigital/dp-upload-service/config"
	"github.com/ONSdigital/dp-upload-service/service"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
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
	return generateS3Client(cfg, cfg.UploadBucketName)
}

func (e external) DoGetStaticFileS3Uploader(ctx context.Context, cfg *config.Config) (dpaws.S3Clienter, error) {
	return generateS3Client(cfg, cfg.StaticFilesEncryptedBucketName)
}

func generateS3Client(cfg *config.Config, bucketName string) (dpaws.S3Clienter, error) {
	s, err := session.NewSession(&aws.Config{
		Endpoint:         aws.String(localStackHost),
		Region:           aws.String(cfg.AwsRegion),
		S3ForcePathStyle: aws.Bool(true),
		Credentials:      credentials.NewStaticCredentials("test", "test", ""),
	})

	if err != nil {
		fmt.Println("S3 ERROR: " + err.Error())
	}

	return s3client.NewClientWithSession(bucketName, s), nil
}
