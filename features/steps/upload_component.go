package steps

import (
	"context"
	componenttest "github.com/ONSdigital/dp-component-test"
	dphttp "github.com/ONSdigital/dp-net/http"
	"github.com/ONSdigital/dp-upload-service/service"
	"net/http"
	"time"
)

type UploadComponent struct {
	server     *dphttp.Server
	svc        *service.Service
	svcList    *service.ExternalServiceList
	ApiFeature *componenttest.APIFeature
	errChan    chan error
}

func NewUploadComponent() *UploadComponent {
	s := dphttp.NewServer("", http.NewServeMux())
	s.HandleOSSignals = false

	return &UploadComponent{
		server:  s,
		errChan: make(chan error),
		svcList: service.NewServiceList(external{s}),
	}
}

func (c *UploadComponent) Initialiser() (http.Handler, error) {
	var err error
	c.svc, err = service.Run(context.Background(), c.svcList, "1", "1", "1", c.errChan)
	time.Sleep(1 * time.Second) // Wait for healthchecks to run before executing tests. TODO consider moving to a Given step for healthchecks
	return c.server.Handler, err
}

func (c *UploadComponent) Reset() {
}

func (c *UploadComponent) Close() error {
	ctx, _ := context.WithTimeout(context.Background(), 10*time.Second) //nolint
	return c.svc.Close(ctx)
}
