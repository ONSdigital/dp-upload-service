package service

import (
	"context"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"

	"github.com/ONSdigital/dp-upload-service/encryption"

	"github.com/ONSdigital/dp-healthcheck/healthcheck"
	"github.com/ONSdigital/dp-upload-service/config"
	"github.com/ONSdigital/dp-upload-service/upload"
	"github.com/ONSdigital/dp-upload-service/upload/mock"

	"github.com/gorilla/mux"
	"github.com/pkg/errors"
	. "github.com/smartystreets/goconvey/convey"
)

var (
	ctx           = context.Background()
	testBuildTime = "BuildTime"
	testGitCommit = "GitCommit"
	testVersion   = "Version"
)

var (
	errVault       = errors.New("vault error")
	errS3Uploaded  = errors.New("S3 uploaded error")
	errHealthcheck = errors.New("healthCheck error")
)

var funcDoGetVaultErr = func(ctx context.Context, cfg *config.Config) (upload.VaultClienter, error) {
	return nil, errVault
}

var funcDoS3UploadedErr = func(ctx context.Context, cfg *config.Config) (upload.S3Clienter, error) {
	return nil, errS3Uploaded
}

var funcDoGetHealthcheckErr = func(cfg *config.Config, buildTime string, gitCommit string, version string) (HealthChecker, error) {
	return nil, errHealthcheck
}

var funcDoGetHTTPServerNil = func(bindAddr string, router http.Handler) HTTPServer {
	return nil
}

var funcDoGetEncryptionKeyGenerator = func() encryption.GenerateKey {
	return func() []byte {
		return []byte("")
	}
}

func TestRun(t *testing.T) {

	Convey("Given a set of mocked dependencies", t, func() {

		funcHasRoute := func(r *mux.Router, method, path string) bool {
			req := httptest.NewRequest(method, path, nil)
			match := &mux.RouteMatch{}
			return r.Match(req, match)
		}

		vaultMock := &mock.VaultClienterMock{
			CheckerFunc: func(ctx context.Context, state *healthcheck.CheckState) error { return nil },
		}

		s3UploadedMock := &mock.S3ClienterMock{
			CheckerFunc: func(ctx context.Context, state *healthcheck.CheckState) error { return nil },
		}

		hcMock := &HealthCheckerMock{
			AddCheckFunc: func(name string, checker healthcheck.Checker) error { return nil },
			StartFunc:    func(ctx context.Context) {},
		}

		serverWg := &sync.WaitGroup{}
		serverMock := &HTTPServerMock{
			ListenAndServeFunc: func() error {
				serverWg.Done()
				return nil
			},
		}

		funcDoGetVaultOk := func(ctx context.Context, cfg *config.Config) (upload.VaultClienter, error) {
			return vaultMock, nil
		}

		funcDoGetS3UploadedOk := func(ctx context.Context, cfg *config.Config) (upload.S3Clienter, error) {
			return s3UploadedMock, nil
		}

		funcDoGetHealthcheckOk := func(cfg *config.Config, buildTime string, gitCommit string, version string) (HealthChecker, error) {
			return hcMock, nil
		}

		funcDoGetHTTPServer := func(bindAddr string, router http.Handler) HTTPServer {
			return serverMock
		}

		Convey("When initialising s3 uploaded bucket that returns an error", func() {
			initMock := &InitialiserMock{
				DoGetHTTPServerFunc:             funcDoGetHTTPServerNil,
				DoGetS3UploadedFunc:             funcDoS3UploadedErr,
				DoGetStaticFileS3UploaderFunc:   funcDoGetS3UploadedOk,
				DoGetHealthCheckFunc:            funcDoGetHealthcheckOk,
				DoGetVaultFunc:                  funcDoGetVaultOk,
				DoGetEncryptionKeyGeneratorFunc: funcDoGetEncryptionKeyGenerator,
			}
			svcErrors := make(chan error, 1)
			svcList := NewServiceList(initMock)
			_, err := Run(ctx, svcList, testBuildTime, testGitCommit, testVersion, svcErrors)

			Convey("Then service Run fails with the same error and the flag is not set", func() {
				So(err, ShouldResemble, errS3Uploaded)
				So(svcList.S3Uploaded, ShouldBeFalse)
				So(svcList.HealthCheck, ShouldBeFalse)
				So(svcList.Vault, ShouldBeFalse)
			})
		})

		Convey("When initialising vault returns an error", func() {
			initMock := &InitialiserMock{
				DoGetHTTPServerFunc:             funcDoGetHTTPServerNil,
				DoGetVaultFunc:                  funcDoGetVaultErr,
				DoGetS3UploadedFunc:             funcDoGetS3UploadedOk,
				DoGetStaticFileS3UploaderFunc:   funcDoGetS3UploadedOk,
				DoGetHealthCheckFunc:            funcDoGetHealthcheckOk,
				DoGetEncryptionKeyGeneratorFunc: funcDoGetEncryptionKeyGenerator,
			}
			svcErrors := make(chan error, 1)
			svcList := NewServiceList(initMock)
			_, err := Run(ctx, svcList, testBuildTime, testGitCommit, testVersion, svcErrors)

			Convey("Then service Run fails with the same error and the flag is not set", func() {
				So(err, ShouldResemble, errVault)
				So(svcList.Vault, ShouldBeFalse)
				So(svcList.S3Uploaded, ShouldBeTrue)
				So(svcList.HealthCheck, ShouldBeFalse)
			})
		})

		Convey("When initialising healthcheck that returns an error", func() {
			initMock := &InitialiserMock{
				DoGetHTTPServerFunc:             funcDoGetHTTPServerNil,
				DoGetHealthCheckFunc:            funcDoGetHealthcheckErr,
				DoGetS3UploadedFunc:             funcDoGetS3UploadedOk,
				DoGetStaticFileS3UploaderFunc:   funcDoGetS3UploadedOk,
				DoGetVaultFunc:                  funcDoGetVaultOk,
				DoGetEncryptionKeyGeneratorFunc: funcDoGetEncryptionKeyGenerator,
			}
			svcErrors := make(chan error, 1)
			svcList := NewServiceList(initMock)
			_, err := Run(ctx, svcList, testBuildTime, testGitCommit, testVersion, svcErrors)

			Convey("Then service Run fails with healthcheck error and HealthCheck flag is not set", func() {
				So(err, ShouldResemble, errHealthcheck)
				So(svcList.S3Uploaded, ShouldBeTrue)
				So(svcList.HealthCheck, ShouldBeFalse)
				So(svcList.Vault, ShouldBeTrue)
			})
		})

		Convey("When all dependencies are successfully initialised", func() {

			initMock := &InitialiserMock{
				DoGetHTTPServerFunc:             funcDoGetHTTPServer,
				DoGetHealthCheckFunc:            funcDoGetHealthcheckOk,
				DoGetS3UploadedFunc:             funcDoGetS3UploadedOk,
				DoGetStaticFileS3UploaderFunc:   funcDoGetS3UploadedOk,
				DoGetVaultFunc:                  funcDoGetVaultOk,
				DoGetEncryptionKeyGeneratorFunc: funcDoGetEncryptionKeyGenerator,
			}

			svcErrors := make(chan error, 1)
			svcList := NewServiceList(initMock)
			serverWg.Add(1)

			s, err := Run(ctx, svcList, testBuildTime, testGitCommit, testVersion, svcErrors)

			Convey("Then service Run succeeds with no eror and all the flags are set", func() {
				So(err, ShouldBeNil)
				So(svcList.S3Uploaded, ShouldBeTrue)
				So(svcList.HealthCheck, ShouldBeTrue)
				So(svcList.Vault, ShouldBeTrue)

			})

			Convey("And the following routes should have been added", func() {
				So(funcHasRoute(s.router, "GET", "/health"), ShouldBeTrue)
				So(funcHasRoute(s.router, "GET", "/upload"), ShouldBeTrue)
				So(funcHasRoute(s.router, "POST", "/upload"), ShouldBeTrue)
				So(funcHasRoute(s.router, "GET", "/upload/{id}"), ShouldBeTrue)

			})

			Convey("The checkers are registered and the healthcheck and http server started", func() {
				So(len(hcMock.AddCheckCalls()), ShouldEqual, 2)
				So(hcMock.AddCheckCalls()[0].Name, ShouldResemble, "Vault client")
				So(hcMock.AddCheckCalls()[1].Name, ShouldResemble, "S3 uploaded bucket")
				So(len(initMock.DoGetHTTPServerCalls()), ShouldEqual, 1)
				So(initMock.DoGetHTTPServerCalls()[0].BindAddr, ShouldEqual, "localhost:25100")
				So(len(hcMock.StartCalls()), ShouldEqual, 1)
				serverWg.Wait() // Wait for HTTP server go-routine to finish
				So(len(serverMock.ListenAndServeCalls()), ShouldEqual, 1)
			})
		})

		Convey("When the Checkers cannot be registered", func() {

			errAddCheckFail := errors.New("Error(s) registering checkers for healthcheck")
			hcMockAddFail := &HealthCheckerMock{
				AddCheckFunc: func(name string, checker healthcheck.Checker) error { return errAddCheckFail },
				StartFunc:    func(ctx context.Context) {},
			}

			initMock := &InitialiserMock{
				DoGetHTTPServerFunc: funcDoGetHTTPServerNil,
				DoGetVaultFunc:      funcDoGetVaultOk,
				DoGetHealthCheckFunc: func(cfg *config.Config, buildTime string, gitCommit string, version string) (HealthChecker, error) {
					return hcMockAddFail, nil
				},
				DoGetS3UploadedFunc:             funcDoGetS3UploadedOk,
				DoGetStaticFileS3UploaderFunc:   funcDoGetS3UploadedOk,
				DoGetEncryptionKeyGeneratorFunc: funcDoGetEncryptionKeyGenerator,
			}
			svcErrors := make(chan error, 1)
			svcList := NewServiceList(initMock)
			_, err := Run(ctx, svcList, testBuildTime, testGitCommit, testVersion, svcErrors)

			Convey("Then service Run fails, but all checks try to register", func() {
				So(err, ShouldNotBeNil)
				So(err.Error(), ShouldResemble, errAddCheckFail.Error())
				So(svcList.HealthCheck, ShouldBeTrue)
				So(svcList.S3Uploaded, ShouldBeTrue)
				So(svcList.Vault, ShouldBeTrue)
				So(len(hcMockAddFail.AddCheckCalls()), ShouldEqual, 2)
				So(hcMockAddFail.AddCheckCalls()[0].Name, ShouldResemble, "Vault client")
				So(hcMockAddFail.AddCheckCalls()[1].Name, ShouldResemble, "S3 uploaded bucket")
			})
		})
	})
}

func TestClose(t *testing.T) {

	Convey("Given the service is correctly initialised", t, func() {

		hcStopped := false

		vaultMock := &mock.VaultClienterMock{
			CheckerFunc: func(ctx context.Context, state *healthcheck.CheckState) error { return nil },
		}

		s3UploadedMock := &mock.S3ClienterMock{
			CheckerFunc: func(ctx context.Context, state *healthcheck.CheckState) error { return nil },
		}

		// healthcheck Stop does not depend on any other service being closed/stopped
		hcMock := &HealthCheckerMock{
			AddCheckFunc: func(name string, checker healthcheck.Checker) error { return nil },
			StartFunc:    func(ctx context.Context) {},
			StopFunc:     func() { hcStopped = true },
		}

		// server Shutdown will fail if healthcheck is not stopped
		serverMock := &HTTPServerMock{
			ListenAndServeFunc: func() error { return nil },
			ShutdownFunc: func(ctx context.Context) error {
				if !hcStopped {
					return errors.New("Server stopped before healthcheck")
				}
				return nil
			},
		}

		Convey("When closing the service", func() {

			initMock := &InitialiserMock{
				DoGetHTTPServerFunc:           func(bindAddr string, router http.Handler) HTTPServer { return serverMock },
				DoGetVaultFunc:                func(ctx context.Context, cfg *config.Config) (upload.VaultClienter, error) { return vaultMock, nil },
				DoGetS3UploadedFunc:           func(ctx context.Context, cfg *config.Config) (upload.S3Clienter, error) { return s3UploadedMock, nil },
				DoGetStaticFileS3UploaderFunc: func(ctx context.Context, cfg *config.Config) (upload.S3Clienter, error) { return s3UploadedMock, nil },
				DoGetHealthCheckFunc: func(cfg *config.Config, buildTime string, gitCommit string, version string) (HealthChecker, error) {
					return hcMock, nil
				},
				DoGetEncryptionKeyGeneratorFunc: funcDoGetEncryptionKeyGenerator,
			}

			svcErrors := make(chan error, 1)
			svcList := NewServiceList(initMock)
			svc, err := Run(ctx, svcList, testBuildTime, testGitCommit, testVersion, svcErrors)
			So(err, ShouldBeNil)
			err = svc.Close(context.Background())
			Convey("Then all the dependencies should close", func() {
				So(err, ShouldBeNil)
				So(len(hcMock.StopCalls()), ShouldEqual, 1)
				So(len(serverMock.ShutdownCalls()), ShouldEqual, 1)
			})
		})

		Convey("When services fail to stop", func() {

			failingserverMock := &HTTPServerMock{
				ListenAndServeFunc: func() error { return nil },
				ShutdownFunc: func(ctx context.Context) error {
					return errors.New("Failed to stop http server")
				},
			}

			initMock := &InitialiserMock{
				DoGetHTTPServerFunc:           func(bindAddr string, router http.Handler) HTTPServer { return failingserverMock },
				DoGetVaultFunc:                func(ctx context.Context, cfg *config.Config) (upload.VaultClienter, error) { return vaultMock, nil },
				DoGetS3UploadedFunc:           func(ctx context.Context, cfg *config.Config) (upload.S3Clienter, error) { return s3UploadedMock, nil },
				DoGetStaticFileS3UploaderFunc: func(ctx context.Context, cfg *config.Config) (upload.S3Clienter, error) { return s3UploadedMock, nil },
				DoGetHealthCheckFunc: func(cfg *config.Config, buildTime string, gitCommit string, version string) (HealthChecker, error) {
					return hcMock, nil
				},
				DoGetEncryptionKeyGeneratorFunc: funcDoGetEncryptionKeyGenerator,
			}

			svcErrors := make(chan error, 1)
			svcList := NewServiceList(initMock)
			svc, err := Run(ctx, svcList, testBuildTime, testGitCommit, testVersion, svcErrors)
			So(err, ShouldBeNil)

			err = svc.Close(context.Background())
			Convey("Then the Close operation tries to close all dependencies and returns an error", func() {
				So(err, ShouldNotBeNil)
				So(len(hcMock.StopCalls()), ShouldEqual, 1)
				So(len(failingserverMock.ShutdownCalls()), ShouldEqual, 1)
			})

		})
	})
}
