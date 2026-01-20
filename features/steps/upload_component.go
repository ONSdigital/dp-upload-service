package steps

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	awsConfig "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"

	componenttest "github.com/ONSdigital/dp-component-test"
	dphttp "github.com/ONSdigital/dp-net/v3/http"
	"github.com/ONSdigital/dp-upload-service/config"
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
	time.Sleep(5 * time.Second) // Wait for healthchecks to run before executing tests. TODO consider moving to a Given step for healthchecks
	return c.server.Handler, err
}

func (c *UploadComponent) Reset() {
	// clear out test bucket
	cfg, _ := config.Get()
	var err error

	AWSConfig, err := awsConfig.LoadDefaultConfig(
		context.Background(),
		awsConfig.WithRegion(cfg.AwsRegion),
		awsConfig.WithCredentialsProvider(credentials.NewStaticCredentialsProvider("test", "test", "")),
	)

	if err != nil {
		panic(fmt.Sprintf("Failed to load default config s3: %s", err.Error()))
	}

	s3client := s3.NewFromConfig(AWSConfig, func(options *s3.Options) {
		options.BaseEndpoint = aws.String(localStackHost)
		options.UsePathStyle = true
	})
	deleteObjectsInBucket(cfg.UploadBucketName, s3client)

	// removet
	err = os.RemoveAll(testFilePath)
	if err != nil {
		panic("Failed to remove test-data")
	}
	if err := os.Mkdir(testFilePath, 0755); err != nil {
		panic(fmt.Sprintf("Failed to create test-data directory: %s", err.Error()))
	}
}

func (c *UploadComponent) Close() error {
	ctx, _ := context.WithTimeout(context.Background(), 10*time.Second) //nolint
	if c.svc != nil {
		return c.svc.Close(ctx)
	}
	return nil
}

func deleteObjectsInBucket(bucketName string, client *s3.Client) {
	if client != nil {
		listObjectInput := &s3.ListObjectsInput{
			Bucket: aws.String(bucketName),
		}
		listObjectOutput, err := client.ListObjects(context.Background(), listObjectInput)

		if err != nil {
			fmt.Println("Error is :", err.Error())
			return
		}
		for _, object := range listObjectOutput.Contents {
			deleteObjectInput := &s3.DeleteObjectInput{
				Bucket: aws.String(bucketName),
				Key:    object.Key,
			}
			_, _ = client.DeleteObject(context.Background(), deleteObjectInput)
		}
	}
}
