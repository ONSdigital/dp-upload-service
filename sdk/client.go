package sdk

import (
	"context"

	"github.com/ONSdigital/dp-api-clients-go/v2/health"
	"github.com/ONSdigital/dp-healthcheck/healthcheck"
)

const (
	serviceName = "dp-upload-service"
)

type Client struct {
	hcCli *health.Client
}

// New creates a new instance of Client with the provided upload service URL
func New(uploadServiceURL string) *Client {
	return &Client{
		hcCli: health.NewClient(serviceName, uploadServiceURL),
	}
}

// NewWithHealthClient creates a new instance of Client using an existing health.Client
func NewWithHealthClient(hcCli *health.Client) *Client {
	return &Client{
		hcCli: health.NewClientWithClienter(serviceName, hcCli.URL, hcCli.Client),
	}
}

// URL returns the URL used by the Client
func (cli *Client) URL() string {
	return cli.hcCli.URL
}

// Health returns the health.Client used by the Client
func (cli *Client) Health() *health.Client {
	return cli.hcCli
}

// Checker calls the health.Client's Checker method
func (cli *Client) Checker(ctx context.Context, check *healthcheck.CheckState) error {
	return cli.hcCli.Checker(ctx, check)
}
