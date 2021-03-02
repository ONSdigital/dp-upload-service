// Code generated by moq; DO NOT EDIT.
// github.com/matryer/moq

package service

import (
	"context"
	"github.com/ONSdigital/dp-upload-service/config"
	"github.com/ONSdigital/dp-upload-service/upload"
	"net/http"
	"sync"
)

// Ensure, that InitialiserMock does implement Initialiser.
// If this is not the case, regenerate this file with moq.
var _ Initialiser = &InitialiserMock{}

// InitialiserMock is a mock implementation of Initialiser.
//
//     func TestSomethingThatUsesInitialiser(t *testing.T) {
//
//         // make and configure a mocked Initialiser
//         mockedInitialiser := &InitialiserMock{
//             DoGetHTTPServerFunc: func(bindAddr string, router http.Handler) HTTPServer {
// 	               panic("mock out the DoGetHTTPServer method")
//             },
//             DoGetHealthCheckFunc: func(cfg *config.Config, buildTime string, gitCommit string, version string) (HealthChecker, error) {
// 	               panic("mock out the DoGetHealthCheck method")
//             },
//             DoGetS3UploadedFunc: func(ctx context.Context, cfg *config.Config) (api.S3Clienter, error) {
// 	               panic("mock out the DoGetS3Uploaded method")
//             },
//             DoGetVaultFunc: func(ctx context.Context, cfg *config.Config) (api.VaultClienter, error) {
// 	               panic("mock out the DoGetVault method")
//             },
//         }
//
//         // use mockedInitialiser in code that requires Initialiser
//         // and then make assertions.
//
//     }
type InitialiserMock struct {
	// DoGetHTTPServerFunc mocks the DoGetHTTPServer method.
	DoGetHTTPServerFunc func(bindAddr string, router http.Handler) HTTPServer

	// DoGetHealthCheckFunc mocks the DoGetHealthCheck method.
	DoGetHealthCheckFunc func(cfg *config.Config, buildTime string, gitCommit string, version string) (HealthChecker, error)

	// DoGetS3UploadedFunc mocks the DoGetS3Uploaded method.
	DoGetS3UploadedFunc func(ctx context.Context, cfg *config.Config) (upload.S3Clienter, error)

	// DoGetVaultFunc mocks the DoGetVault method.
	DoGetVaultFunc func(ctx context.Context, cfg *config.Config) (upload.VaultClienter, error)

	// calls tracks calls to the methods.
	calls struct {
		// DoGetHTTPServer holds details about calls to the DoGetHTTPServer method.
		DoGetHTTPServer []struct {
			// BindAddr is the bindAddr argument value.
			BindAddr string
			// Router is the router argument value.
			Router http.Handler
		}
		// DoGetHealthCheck holds details about calls to the DoGetHealthCheck method.
		DoGetHealthCheck []struct {
			// Cfg is the cfg argument value.
			Cfg *config.Config
			// BuildTime is the buildTime argument value.
			BuildTime string
			// GitCommit is the gitCommit argument value.
			GitCommit string
			// Version is the version argument value.
			Version string
		}
		// DoGetS3Uploaded holds details about calls to the DoGetS3Uploaded method.
		DoGetS3Uploaded []struct {
			// Ctx is the ctx argument value.
			Ctx context.Context
			// Cfg is the cfg argument value.
			Cfg *config.Config
		}
		// DoGetVault holds details about calls to the DoGetVault method.
		DoGetVault []struct {
			// Ctx is the ctx argument value.
			Ctx context.Context
			// Cfg is the cfg argument value.
			Cfg *config.Config
		}
	}
	lockDoGetHTTPServer  sync.RWMutex
	lockDoGetHealthCheck sync.RWMutex
	lockDoGetS3Uploaded  sync.RWMutex
	lockDoGetVault       sync.RWMutex
}

// DoGetHTTPServer calls DoGetHTTPServerFunc.
func (mock *InitialiserMock) DoGetHTTPServer(bindAddr string, router http.Handler) HTTPServer {
	if mock.DoGetHTTPServerFunc == nil {
		panic("InitialiserMock.DoGetHTTPServerFunc: method is nil but Initialiser.DoGetHTTPServer was just called")
	}
	callInfo := struct {
		BindAddr string
		Router   http.Handler
	}{
		BindAddr: bindAddr,
		Router:   router,
	}
	mock.lockDoGetHTTPServer.Lock()
	mock.calls.DoGetHTTPServer = append(mock.calls.DoGetHTTPServer, callInfo)
	mock.lockDoGetHTTPServer.Unlock()
	return mock.DoGetHTTPServerFunc(bindAddr, router)
}

// DoGetHTTPServerCalls gets all the calls that were made to DoGetHTTPServer.
// Check the length with:
//     len(mockedInitialiser.DoGetHTTPServerCalls())
func (mock *InitialiserMock) DoGetHTTPServerCalls() []struct {
	BindAddr string
	Router   http.Handler
} {
	var calls []struct {
		BindAddr string
		Router   http.Handler
	}
	mock.lockDoGetHTTPServer.RLock()
	calls = mock.calls.DoGetHTTPServer
	mock.lockDoGetHTTPServer.RUnlock()
	return calls
}

// DoGetHealthCheck calls DoGetHealthCheckFunc.
func (mock *InitialiserMock) DoGetHealthCheck(cfg *config.Config, buildTime string, gitCommit string, version string) (HealthChecker, error) {
	if mock.DoGetHealthCheckFunc == nil {
		panic("InitialiserMock.DoGetHealthCheckFunc: method is nil but Initialiser.DoGetHealthCheck was just called")
	}
	callInfo := struct {
		Cfg       *config.Config
		BuildTime string
		GitCommit string
		Version   string
	}{
		Cfg:       cfg,
		BuildTime: buildTime,
		GitCommit: gitCommit,
		Version:   version,
	}
	mock.lockDoGetHealthCheck.Lock()
	mock.calls.DoGetHealthCheck = append(mock.calls.DoGetHealthCheck, callInfo)
	mock.lockDoGetHealthCheck.Unlock()
	return mock.DoGetHealthCheckFunc(cfg, buildTime, gitCommit, version)
}

// DoGetHealthCheckCalls gets all the calls that were made to DoGetHealthCheck.
// Check the length with:
//     len(mockedInitialiser.DoGetHealthCheckCalls())
func (mock *InitialiserMock) DoGetHealthCheckCalls() []struct {
	Cfg       *config.Config
	BuildTime string
	GitCommit string
	Version   string
} {
	var calls []struct {
		Cfg       *config.Config
		BuildTime string
		GitCommit string
		Version   string
	}
	mock.lockDoGetHealthCheck.RLock()
	calls = mock.calls.DoGetHealthCheck
	mock.lockDoGetHealthCheck.RUnlock()
	return calls
}

// DoGetS3Uploaded calls DoGetS3UploadedFunc.
func (mock *InitialiserMock) DoGetS3Uploaded(ctx context.Context, cfg *config.Config) (upload.S3Clienter, error) {
	if mock.DoGetS3UploadedFunc == nil {
		panic("InitialiserMock.DoGetS3UploadedFunc: method is nil but Initialiser.DoGetS3Uploaded was just called")
	}
	callInfo := struct {
		Ctx context.Context
		Cfg *config.Config
	}{
		Ctx: ctx,
		Cfg: cfg,
	}
	mock.lockDoGetS3Uploaded.Lock()
	mock.calls.DoGetS3Uploaded = append(mock.calls.DoGetS3Uploaded, callInfo)
	mock.lockDoGetS3Uploaded.Unlock()
	return mock.DoGetS3UploadedFunc(ctx, cfg)
}

// DoGetS3UploadedCalls gets all the calls that were made to DoGetS3Uploaded.
// Check the length with:
//     len(mockedInitialiser.DoGetS3UploadedCalls())
func (mock *InitialiserMock) DoGetS3UploadedCalls() []struct {
	Ctx context.Context
	Cfg *config.Config
} {
	var calls []struct {
		Ctx context.Context
		Cfg *config.Config
	}
	mock.lockDoGetS3Uploaded.RLock()
	calls = mock.calls.DoGetS3Uploaded
	mock.lockDoGetS3Uploaded.RUnlock()
	return calls
}

// DoGetVault calls DoGetVaultFunc.
func (mock *InitialiserMock) DoGetVault(ctx context.Context, cfg *config.Config) (upload.VaultClienter, error) {
	if mock.DoGetVaultFunc == nil {
		panic("InitialiserMock.DoGetVaultFunc: method is nil but Initialiser.DoGetVault was just called")
	}
	callInfo := struct {
		Ctx context.Context
		Cfg *config.Config
	}{
		Ctx: ctx,
		Cfg: cfg,
	}
	mock.lockDoGetVault.Lock()
	mock.calls.DoGetVault = append(mock.calls.DoGetVault, callInfo)
	mock.lockDoGetVault.Unlock()
	return mock.DoGetVaultFunc(ctx, cfg)
}

// DoGetVaultCalls gets all the calls that were made to DoGetVault.
// Check the length with:
//     len(mockedInitialiser.DoGetVaultCalls())
func (mock *InitialiserMock) DoGetVaultCalls() []struct {
	Ctx context.Context
	Cfg *config.Config
} {
	var calls []struct {
		Ctx context.Context
		Cfg *config.Config
	}
	mock.lockDoGetVault.RLock()
	calls = mock.calls.DoGetVault
	mock.lockDoGetVault.RUnlock()
	return calls
}
