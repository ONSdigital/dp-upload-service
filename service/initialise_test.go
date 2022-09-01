package service_test

import (
	"context"
	"net/http"
	"testing"
	"time"

	"github.com/ONSdigital/dp-upload-service/aws"
	mock_aws "github.com/ONSdigital/dp-upload-service/aws/mock"
	"github.com/ONSdigital/dp-upload-service/config"
	"github.com/ONSdigital/dp-upload-service/encryption"
	mock_encryption "github.com/ONSdigital/dp-upload-service/encryption/mock"
	"github.com/ONSdigital/dp-upload-service/service"
	mock_service "github.com/ONSdigital/dp-upload-service/service/mock"

	"github.com/gorilla/mux"
	"github.com/pkg/errors"

	. "github.com/smartystreets/goconvey/convey"
)

var errFunc = func() error {
	return errors.New("Server error")
}

var cfg, _ = config.Get()

func TestGetHTTPServer(t *testing.T) {
	Convey("Given a service list that includes a mocked server", t, func() {
		serverMock := &mock_service.HTTPServerMock{}
		newServiceMock := &mock_service.InitialiserMock{
			DoGetHTTPServerFunc: func(bindAddr string, router http.Handler) service.HTTPServer {
				return serverMock
			},
		}
		r := mux.NewRouter()
		svcList := service.NewServiceList(newServiceMock)
		Convey("When GetHTTPServer is called", func() {
			server := svcList.GetHTTPServer(cfg.BindAddr, r)
			Convey("Then the mock server is returned and has been initialised with the correct bind address", func() {
				So(len(newServiceMock.DoGetHTTPServerCalls()), ShouldEqual, 1)
				So(newServiceMock.DoGetHTTPServerCalls()[0].BindAddr, ShouldEqual, cfg.BindAddr)
				So(server, ShouldEqual, serverMock)
			})
		})
	})

	Convey("Given a service list returns a mocked server that errors on ListenAndServe", t, func() {
		serverMock := &mock_service.HTTPServerMock{
			ListenAndServeFunc: errFunc,
		}
		newServiceMock := &mock_service.InitialiserMock{
			DoGetHTTPServerFunc: func(bindAddr string, router http.Handler) service.HTTPServer {
				return serverMock
			},
		}
		svcErrors := make(chan error, 1)
		r := mux.NewRouter()
		var err error
		svcList := service.NewServiceList(newServiceMock)
		Convey("When the server is retrieved and started", func() {
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
			case <-ctx.Done():
				t.Fatal("ListenAndServe returned no error")
				server.Shutdown(context.Background()) //nolint
			}
			Convey("Then the startup has failed and returns the expected error", func() {
				So(err.Error(), ShouldEqual, "Server error")
			})
		})
	})

	Convey("Given a service list that includes a mocked server", t, func() {
		serverMock := &mock_service.HTTPServerMock{
			ListenAndServeFunc: func() error {
				return nil
			},
		}
		newServiceMock := &mock_service.InitialiserMock{
			DoGetHTTPServerFunc: func(bindAddr string, router http.Handler) service.HTTPServer {
				return serverMock
			},
		}
		r := mux.NewRouter()
		svcList := service.NewServiceList(newServiceMock)
		svcErrors := make(chan error, 1)
		Convey("When GetHTTPServer is called", func() {
			server := svcList.GetHTTPServer(cfg.BindAddr, r)
			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			go func() {
				if err := server.ListenAndServe(); err != nil {
					svcErrors <- err
				} else {
					cancel()
				}
			}()

			var err error
			select {
			case err = <-svcErrors:
				cancel()
			case errDone := <-ctx.Done():
				So(errDone, ShouldBeZeroValue)
			}
			Convey("Then the server returns nil", func() {
				So(len(newServiceMock.DoGetHTTPServerCalls()), ShouldEqual, 1)
				So(len(serverMock.ListenAndServeCalls()), ShouldEqual, 1)
				So(err, ShouldBeNil)
			})
		})
	})

}

func TestGetVault(t *testing.T) {
	Convey("Given a service list that includes a mocked vault", t, func() {
		vaultMock := &mock_encryption.VaultClienterMock{}
		newServiceMock := &mock_service.InitialiserMock{
			DoGetVaultFunc: func(ctx context.Context, cfg *config.Config) (encryption.VaultClienter, error) {
				return vaultMock, nil
			},
		}
		svcList := service.NewServiceList(newServiceMock)
		Convey("When GetVault is called", func() {
			vault, _ := svcList.GetVault(ctx, cfg)
			Convey("Then the vault is returned and vault flag is set to true", func() {
				So(svcList.Vault, ShouldBeTrue)
				So(vault, ShouldEqual, vaultMock)
				So(len(newServiceMock.DoGetVaultCalls()), ShouldEqual, 1)
			})
		})
	})

	Convey("Given a service list that returns nil for vault client", t, func() {
		newServiceMock := &mock_service.InitialiserMock{
			DoGetVaultFunc: func(ctx context.Context, cfg *config.Config) (encryption.VaultClienter, error) {
				return nil, errVault
			},
		}
		svcList := service.NewServiceList(newServiceMock)
		Convey("When GetVault is called", func() {
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

	Convey("Given a service list that includes a mocked s3Client", t, func() {

		s3UploadedMock := &mock_aws.S3ClienterMock{}
		newServiceMock := &mock_service.InitialiserMock{
			DoGetS3UploadedFunc: func(ctx context.Context, cfg *config.Config) (aws.S3Clienter, error) {
				return s3UploadedMock, nil
			},
		}
		svcList := service.NewServiceList(newServiceMock)
		Convey("When GetS3Uploaded is called", func() {
			s3Client, err := svcList.GetS3Uploaded(ctx, cfg)
			Convey("Then the S3Uploaded flag is set to true s3Client is returned", func() {
				So(svcList.S3Uploaded, ShouldBeTrue)
				So(s3Client, ShouldEqual, s3UploadedMock)
				So(len(newServiceMock.DoGetS3UploadedCalls()), ShouldEqual, 1)
				So(err, ShouldBeNil)
			})
		})
	})

	Convey("Given a service list returns nil for mocked S3 client", t, func() {
		newServiceMock := &mock_service.InitialiserMock{
			DoGetS3UploadedFunc: func(ctx context.Context, cfg *config.Config) (aws.S3Clienter, error) {
				return nil, errS3Uploaded
			},
		}
		svcList := service.NewServiceList(newServiceMock)
		Convey("When GetS3Uploaded is called", func() {
			s3Client, err := svcList.GetS3Uploaded(ctx, cfg)
			Convey("Then the S3Uploaded flag is set to false and s3Client returns nil ", func() {
				So(s3Client, ShouldBeNil)
				So(err, ShouldResemble, errS3Uploaded)
				So(svcList.S3Uploaded, ShouldBeFalse)
			})
		})
	})
}

func TestGetHealthCheck(t *testing.T) {

	Convey("Given a service list that returns a mocked healthchecker", t, func() {

		hcMock := &mock_service.HealthCheckerMock{}

		newServiceMock := &mock_service.InitialiserMock{
			DoGetHealthCheckFunc: func(cfg *config.Config, buildTime string, gitCommit string, version string) (service.HealthChecker, error) {
				return hcMock, nil

			},
		}
		svcList := service.NewServiceList(newServiceMock)
		Convey("When GetHealthCheck is called", func() {
			hc, err := svcList.GetHealthCheck(cfg, testBuildTime, testGitCommit, testVersion)
			Convey("Then the HealthCheck flag is set to true and HealthCheck is returned", func() {
				So(svcList.HealthCheck, ShouldBeTrue)
				So(hc, ShouldEqual, hcMock)
				So(len(newServiceMock.DoGetHealthCheckCalls()), ShouldEqual, 1)
				So(err, ShouldBeNil)
			})
		})
	})

	Convey("Given a service list that returns nil for healthcheck", t, func() {
		newServiceMock := &mock_service.InitialiserMock{
			DoGetHealthCheckFunc: func(cfg *config.Config, buildTime string, gitCommit string, version string) (service.HealthChecker, error) {
				return nil, errHealthcheck
			},
		}
		svcList := service.NewServiceList(newServiceMock)
		Convey("When GetHealthCheck is called", func() {
			hc, err := svcList.GetHealthCheck(cfg, testBuildTime, testGitCommit, testVersion)
			Convey("Then the HealthCheck flag is set to false and HealthCheck is nil", func() {
				So(hc, ShouldBeNil)
				So(err, ShouldResemble, errHealthcheck)
				So(svcList.HealthCheck, ShouldBeFalse)
			})
		})
	})
}

func TestInit_DoGetVault(t *testing.T) {
	Convey("Given a an empty initialiser struct", t, func() {
		init := service.Init{}
		cfg, err := config.Get()

		Convey("When DoGetVault is called with encryption disabled", func() {
			cfg.EncryptionDisabled = true
			So(err, ShouldBeNil)
			vault, err := init.DoGetVault(ctx, cfg)
			So(err, ShouldBeNil)
			Convey("Then the returned vault client should be nil", func() {
				So(vault, ShouldBeNil)
			})
		})

		Convey("When DoGetVault is called with encryption enabled", func() {
			cfg.EncryptionDisabled = false
			So(err, ShouldBeNil)
			vault, err := init.DoGetVault(ctx, cfg)
			So(err, ShouldBeNil)
			Convey("Then the returned vault client should not be nil", func() {
				So(vault, ShouldNotBeNil)
			})
		})
	})
}
