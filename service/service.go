package service

import (
	"context"
	"net/http"

	"github.com/ONSdigital/dp-upload-service/config"
	"github.com/ONSdigital/dp-upload-service/upload"
	"github.com/ONSdigital/log.go/log"

	"github.com/gorilla/mux"
	"github.com/pkg/errors"
)

// Service contains all the configs, server and clients to run the dp-upload-service API
type Service struct {
	config      *config.Config
	server      HTTPServer
	router      *mux.Router
	serviceList *ExternalServiceList
	healthCheck HealthChecker
	vault       upload.VaultClienter
	uploader    *upload.Uploader
}

// Run the service
func Run(ctx context.Context, serviceList *ExternalServiceList, buildTime, gitCommit, version string, svcErrors chan error) (*Service, error) {

	log.Event(ctx, "running service", log.INFO)

	//Read config
	cfg, err := config.Get()
	if err != nil {
		log.Event(ctx, "unable to retrieve service configuration", log.FATAL, log.Error(err))
		return nil, err
	}

	log.Event(ctx, "got service configuration", log.Data{"config": cfg}, log.INFO)

	// Get HTTP Server with collectionID checkHeader middleware
	r := mux.NewRouter()

	s := serviceList.GetHTTPServer(cfg.BindAddr, r)

	// Get S3Uploaded client
	s3Uploaded, err := serviceList.GetS3Uploaded(ctx, cfg)
	if err != nil {
		log.Event(ctx, "failed to initialise S3 client for uploaded bucket", log.FATAL, log.Error(err))
		return nil, err
	}

	var vault upload.VaultClienter

	// Get Vault client
	vault, err = serviceList.GetVault(ctx, cfg)
	if err != nil {
		log.Event(ctx, "failed to initialise Vault client", log.FATAL, log.Error(err))
		return nil, err
	}

	// Create Uploader with S3 client and Vault
	uploader := upload.New(s3Uploaded, vault, cfg.VaultPath, cfg.AwsRegion, cfg.UploadBucketName)

	hc, err := serviceList.GetHealthCheck(cfg, buildTime, gitCommit, version)

	if err != nil {
		log.Event(ctx, "could not instantiate healthcheck", log.FATAL, log.Error(err))
		return nil, err
	}

	if err := registerCheckers(ctx, hc, vault, s3Uploaded); err != nil {
		log.Event(ctx, "unable to register checkers", log.FATAL, log.Error(err))
		return nil, err
	}
	r.StrictSlash(true).Path("/health").Methods(http.MethodGet).HandlerFunc(hc.Handler)
	r.Path("/upload").Methods(http.MethodGet).HandlerFunc(uploader.CheckUploaded)
	r.Path("/upload").Methods(http.MethodPost).HandlerFunc(uploader.Upload)
	r.Path("/upload/{id}").Methods(http.MethodGet).HandlerFunc(uploader.GetS3URL)

	hc.Start(ctx)

	// Run the http server in a new go-routine
	go func() {
		if err := s.ListenAndServe(); err != nil {
			svcErrors <- errors.Wrap(err, "failure in http listen and serve")
		}
	}()

	return &Service{
		config:      cfg,
		router:      r,
		healthCheck: hc,
		serviceList: serviceList,
		server:      s,
		vault:       vault,
		uploader:    uploader,
	}, nil
}

// Close gracefully shuts the service down in the required order, with timeout
func (svc *Service) Close(ctx context.Context) error {
	timeout := svc.config.GracefulShutdownTimeout
	log.Event(ctx, "commencing graceful shutdown", log.Data{"graceful_shutdown_timeout": timeout}, log.INFO)
	ctx, cancel := context.WithTimeout(ctx, timeout)

	// track shutown gracefully closes up
	var hasShutdownError bool

	go func() {
		defer cancel()

		// stop healthcheck, as it depends on everything else
		if svc.serviceList.HealthCheck {
			svc.healthCheck.Stop()
		}

		// stop any incoming requests before closing any outbound connections
		if err := svc.server.Shutdown(ctx); err != nil {
			log.Event(ctx, "failed to shutdown http server", log.Error(err), log.ERROR)
			hasShutdownError = true
		}

	}()

	// wait for shutdown success (via cancel) or failure (timeout)
	<-ctx.Done()

	if ctx.Err() == context.DeadlineExceeded {
		log.Event(ctx, "shutdown timed out", log.ERROR, log.Error(ctx.Err()))
		return ctx.Err()
	}
	if hasShutdownError {
		err := errors.New("failed to shutdown gracefully")
		log.Event(ctx, "failed to shutdown gracefully ", log.ERROR, log.Error(err))
		return err
	}

	log.Event(ctx, "graceful shutdown was successful", log.INFO)
	return nil

}

func registerCheckers(ctx context.Context,
	hc HealthChecker,
	vault upload.VaultClienter,
	s3Uploaded upload.S3Clienter) (err error) {

	hasErrors := false

	if vault != nil {
		if err = hc.AddCheck("Vault client", vault.Checker); err != nil {
			hasErrors = true
			log.Event(ctx, "error adding check for vault", log.ERROR, log.Error(err))
		}
	}

	if err := hc.AddCheck("S3 uploaded bucket", s3Uploaded.Checker); err != nil {
		hasErrors = true
		log.Event(ctx, "error adding check for s3Private uploaded bucket", log.ERROR, log.Error(err))
	}

	if hasErrors {
		return errors.New("Error(s) registering checkers for healthcheck")
	}

	return nil
}
