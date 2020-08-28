package api

import (
	"context"

	"github.com/ONSdigital/dp-healthcheck/healthcheck"
)

//go:generate moq -out mock/s3.go -pkg mock . S3Clienter

//S3Clienter defines the required method
type S3Clienter interface {
	Checker(ctx context.Context, state *healthcheck.CheckState) error
}
