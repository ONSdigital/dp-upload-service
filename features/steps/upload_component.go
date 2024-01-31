package steps

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/ONSdigital/dp-upload-service/config"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"

	componenttest "github.com/ONSdigital/dp-component-test"
	dphttp "github.com/ONSdigital/dp-net/v2/http"
	"github.com/ONSdigital/dp-upload-service/service"
)

const (
	localStackHost = "http://localstack:4566"
	testFilePath   = "test-data"
)

type UploadComponent struct {
	server       *dphttp.Server
	svc          *service.Service
	svcList      *service.ExternalServiceList
	ApiFeature   *componenttest.APIFeature
	fileMetadata map[string]string

	errChan chan error
}

func NewUploadComponent() *UploadComponent {
	s := dphttp.NewServer("", http.NewServeMux())
	s.HandleOSSignals = false

	return &UploadComponent{
		server:  s,
		errChan: make(chan error),
		svcList: service.NewServiceList(external{Server: s}),
	}
}

func (c *UploadComponent) Initialiser() (http.Handler, error) {
	var err error
	c.svcList = service.NewServiceList(external{Server: c.server})
	c.svc, err = service.Run(context.Background(), c.svcList, "1", "1", "1", c.errChan)
	time.Sleep(1 * time.Second) // Wait for healthchecks to run before executing tests. TODO consider moving to a Given step for healthchecks
	return c.server.Handler, err
}

func (c *UploadComponent) Reset() {
	// clear out test bucket
	cfg, _ := config.Get()
	s, _ := session.NewSession(&aws.Config{
		Endpoint:         aws.String(localStackHost),
		Region:           aws.String(cfg.AwsRegion),
		S3ForcePathStyle: aws.Bool(true),
		Credentials:      credentials.NewStaticCredentials("test", "test", ""),
	})

	s3client := s3.New(s)

	err := s3manager.NewBatchDeleteWithClient(s3client).Delete(
		aws.BackgroundContext(), s3manager.NewDeleteListIterator(s3client, &s3.ListObjectsInput{
			Bucket: aws.String(cfg.UploadBucketName),
		}))

	if err != nil {
		panic(fmt.Sprintf("Failed to empty localstack s3: %s", err.Error()))
	}

	// removet
	err = os.RemoveAll(testFilePath)
	if err != nil {
		panic("Failed to remove test-data")
	}
	os.Mkdir(testFilePath, 0755)
}

func (c *UploadComponent) Close() error {
	ctx, _ := context.WithTimeout(context.Background(), 10*time.Second) //nolint
	if c.svc != nil {
		return c.svc.Close(ctx)
	}
	return nil
}
