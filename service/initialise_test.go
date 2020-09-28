package service

import (
	"context"
	"net/http"
	"testing"
	"time"

	"github.com/ONSdigital/dp-healthcheck/healthcheck"
	"github.com/ONSdigital/dp-upload-service/api"
	"github.com/ONSdigital/dp-upload-service/config"

	"github.com/gorilla/mux"
	"github.com/pkg/errors"
	. "github.com/smartystreets/goconvey/convey"
)

var errFunc = func() error {
	return errors.New("Server error")
}

var cfg, _ = config.Get()
var err error

func TestGetHTTPServer(t *testing.T) {
	Convey("Given a service list returns a server", t, func() {
		serverMock := &HTTPServerMock{}
		newServiceMock := &InitialiserMock{
			DoGetHTTPServerFunc: func(bindAddr string, router http.Handler) HTTPServer {
				return serverMock
			},
		}
		r := mux.NewRouter()
		svcList := NewServiceList(newServiceMock)
		Convey("When the server values are retrieved", func() {
			server := svcList.GetHTTPServer(cfg.BindAddr, r)
			Convey("Then http server is returned with the correct bind address ", func() {
				So(len(newServiceMock.DoGetHTTPServerCalls()), ShouldEqual, 1)
				So(newServiceMock.DoGetHTTPServerCalls()[0].BindAddr, ShouldEqual, cfg.BindAddr)
				So(server, ShouldEqual, serverMock)
			})
		})
	})

	Convey("Given a service list returns server error", t, func() {
		serverMock := &HTTPServerMock{
			ListenAndServeFunc: errFunc,
		}
		newServiceMock := &InitialiserMock{
			DoGetHTTPServerFunc: func(bindAddr string, router http.Handler) HTTPServer {
				return serverMock
			},
		}
		svcErrors := make(chan error, 1)
		r := mux.NewRouter()
		svcList := NewServiceList(newServiceMock)
		Convey("When the server values are retrieved", func() {
			server := svcList.GetHTTPServer(cfg.BindAddr, r)
			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			go func() {
				if err := server.ListenAndServe(); err != nil {
					svcErrors <- err
				}
			}()

			select {
			case err = <-svcErrors:
				cancel()
			case errDone := <-ctx.Done():
				So(errDone, ShouldBeNil)
				cancel()
				server.Shutdown(context.Background())
			}
			Convey("Then the GetHTTPServer function returns a server error", func() {
				So(err.Error(), ShouldEqual, "Server error")
			})
		})
	})

	Convey("Given a service list with Http server listenandserve function returning nil", t, func() {
		serverMock := &HTTPServerMock{
			ListenAndServeFunc: func() error {
				return nil
			},
		}
		newServiceMock := &InitialiserMock{
			DoGetHTTPServerFunc: func(bindAddr string, router http.Handler) HTTPServer {
				return serverMock
			},
		}
		r := mux.NewRouter()
		svcList := NewServiceList(newServiceMock)
		Convey("When the server values are retrieved", func() {
			svcList.GetHTTPServer(cfg.BindAddr, r)
			Convey("Then the server's ListenAndServeFunc returns a nil", func() {
				So(len(newServiceMock.DoGetHTTPServerCalls()), ShouldEqual, 1)
				So(serverMock.ListenAndServeFunc(), ShouldBeNil)
			})
		})
	})

}

func TestGetVault(t *testing.T) {
	Convey("Given a service list the func creates a Vault", t, func() {
		vaultMock := &api.VaultClienterMock{}
		newServiceMock := &InitialiserMock{
			DoGetVaultFunc: func(ctx context.Context, cfg *config.Config) (api.VaultClienter, error) {
				return vaultMock, nil
			},
		}
		svcList := NewServiceList(newServiceMock)
		Convey("When the vault information is retrieved from the service list", func() {
			vault, _ := svcList.GetVault(ctx, cfg)
			Convey("Then the vault flag is set to true and DoGetVault function is called", func() {
				So(svcList.Vault, ShouldBeTrue)
				So(vault, ShouldEqual, vaultMock)
				So(len(newServiceMock.DoGetVaultCalls()), ShouldEqual, 1)
			})
		})
	})

	Convey("Given a service list returns a error for vault client", t, func() {
		newServiceMock := &InitialiserMock{
			DoGetVaultFunc: func(ctx context.Context, cfg *config.Config) (api.VaultClienter, error) {
				return nil, errVault
			},
		}
		svcList := NewServiceList(newServiceMock)
		Convey("When the vault information is retrieved from the service list", func() {
			vault, err := svcList.GetVault(ctx, cfg)
			Convey("Then the vault flag is set to false and vault is nil", func() {
				So(vault, ShouldBeNil)
				So(err, ShouldResemble, errVault)
				So(svcList.Vault, ShouldBeFalse)
			})
		})
	})
}

func TestGetS3Uploaded(t *testing.T) {

	Convey("Given a service list the func creates a S3 client ", t, func() {

		s3UploadedMock := &api.S3ClienterMock{}
		newServiceMock := &InitialiserMock{
			DoGetS3UploadedFunc: func(ctx context.Context, cfg *config.Config) (api.S3Clienter, error) {
				return s3UploadedMock, nil
			},
		}
		svcList := NewServiceList(newServiceMock)
		Convey("When the s3client value is retrieved", func() {
			s3Client, _ := svcList.GetS3Uploaded(ctx, cfg)
			Convey("Then the S3Uploaded flag is set to true and DoGetS3Uploaded function is called", func() {
				So(svcList.S3Uploaded, ShouldBeTrue)
				So(s3Client, ShouldEqual, s3UploadedMock)
				So(len(newServiceMock.DoGetS3UploadedCalls()), ShouldEqual, 1)
			})
		})
	})

	Convey("Given a service list returns a error for S3 client", t, func() {
		newServiceMock := &InitialiserMock{
			DoGetS3UploadedFunc: func(ctx context.Context, cfg *config.Config) (api.S3Clienter, error) {
				return nil, errS3Uploaded
			},
		}
		svcList := NewServiceList(newServiceMock)
		Convey("When the value for s3client is retrieved from the service list", func() {
			s3Client, err := svcList.GetS3Uploaded(ctx, cfg)
			Convey("Then the S3Uploaded flag is set to false and s3Client is nil ", func() {
				So(s3Client, ShouldBeNil)
				So(err, ShouldResemble, errS3Uploaded)
				So(svcList.S3Uploaded, ShouldBeFalse)
			})
		})
	})
}

func TestGetHealthCheck(t *testing.T) {

	Convey("Given a service list the func creates a healthcheck, sets the healthcheck flag to true and calls the DoGetHealthCheck function", t, func() {

		hcMock := &HealthCheckerMock{
			AddCheckFunc: func(name string, checker healthcheck.Checker) error { return nil },
			StartFunc:    func(ctx context.Context) {},
		}

		newServiceMock := &InitialiserMock{
			DoGetHealthCheckFunc: func(cfg *config.Config, buildTime string, gitCommit string, version string) (HealthChecker, error) {
				return hcMock, nil

			},
		}
		svcList := NewServiceList(newServiceMock)
		Convey("When the health check information is retrieved from the service list", func() {
			hc, _ := svcList.GetHealthCheck(cfg, testBuildTime, testGitCommit, testVersion)
			Convey("Then the HealthCheck flag is set to true and DoGetHealthCheck function is called", func() {
				So(svcList.HealthCheck, ShouldBeTrue)
				So(hc, ShouldEqual, hcMock)
				So(len(newServiceMock.DoGetHealthCheckCalls()), ShouldEqual, 1)
			})
		})
	})

	Convey("Given a service list returns a error for healthcheck", t, func() {
		newServiceMock := &InitialiserMock{
			DoGetHealthCheckFunc: func(cfg *config.Config, buildTime string, gitCommit string, version string) (HealthChecker, error) {
				return nil, errHealthcheck
			},
		}
		svcList := NewServiceList(newServiceMock)
		Convey("When the health check information is retrieved from the service list", func() {
			hc, err := svcList.GetHealthCheck(cfg, testBuildTime, testGitCommit, testVersion)
			Convey("Then the HealthCheck flag is set to false and HealthCheck is nil", func() {
				So(hc, ShouldBeNil)
				So(err, ShouldResemble, errHealthcheck)
				So(svcList.HealthCheck, ShouldBeFalse)
			})
		})
	})
}
