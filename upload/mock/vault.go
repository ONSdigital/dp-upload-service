// Code generated by moq; DO NOT EDIT.
// github.com/matryer/moq

package mock

import (
	"context"
	"github.com/ONSdigital/dp-healthcheck/healthcheck"
	"github.com/ONSdigital/dp-upload-service/upload"
	"sync"
)

// Ensure, that VaultClienterMock does implement upload.VaultClienter.
// If this is not the case, regenerate this file with moq.
var _ upload.VaultClienter = &VaultClienterMock{}

// VaultClienterMock is a mock implementation of upload.VaultClienter.
//
// 	func TestSomethingThatUsesVaultClienter(t *testing.T) {
//
// 		// make and configure a mocked upload.VaultClienter
// 		mockedVaultClienter := &VaultClienterMock{
// 			CheckerFunc: func(ctx context.Context, state *healthcheck.CheckState) error {
// 				panic("mock out the Checker method")
// 			},
// 			ReadKeyFunc: func(path string, key string) (string, error) {
// 				panic("mock out the ReadKey method")
// 			},
// 			WriteKeyFunc: func(path string, key string, value string) error {
// 				panic("mock out the WriteKey method")
// 			},
// 		}
//
// 		// use mockedVaultClienter in code that requires upload.VaultClienter
// 		// and then make assertions.
//
// 	}
type VaultClienterMock struct {
	// CheckerFunc mocks the Checker method.
	CheckerFunc func(ctx context.Context, state *healthcheck.CheckState) error

	// ReadKeyFunc mocks the ReadKey method.
	ReadKeyFunc func(path string, key string) (string, error)

	// WriteKeyFunc mocks the WriteKey method.
	WriteKeyFunc func(path string, key string, value string) error

	// calls tracks calls to the methods.
	calls struct {
		// Checker holds details about calls to the Checker method.
		Checker []struct {
			// Ctx is the ctx argument value.
			Ctx context.Context
			// State is the state argument value.
			State *healthcheck.CheckState
		}
		// ReadKey holds details about calls to the ReadKey method.
		ReadKey []struct {
			// Path is the path argument value.
			Path string
			// Key is the key argument value.
			Key string
		}
		// WriteKey holds details about calls to the WriteKey method.
		WriteKey []struct {
			// Path is the path argument value.
			Path string
			// Key is the key argument value.
			Key string
			// Value is the value argument value.
			Value string
		}
	}
	lockChecker  sync.RWMutex
	lockReadKey  sync.RWMutex
	lockWriteKey sync.RWMutex
}

// Checker calls CheckerFunc.
func (mock *VaultClienterMock) Checker(ctx context.Context, state *healthcheck.CheckState) error {
	if mock.CheckerFunc == nil {
		panic("VaultClienterMock.CheckerFunc: method is nil but VaultClienter.Checker was just called")
	}
	callInfo := struct {
		Ctx   context.Context
		State *healthcheck.CheckState
	}{
		Ctx:   ctx,
		State: state,
	}
	mock.lockChecker.Lock()
	mock.calls.Checker = append(mock.calls.Checker, callInfo)
	mock.lockChecker.Unlock()
	return mock.CheckerFunc(ctx, state)
}

// CheckerCalls gets all the calls that were made to Checker.
// Check the length with:
//     len(mockedVaultClienter.CheckerCalls())
func (mock *VaultClienterMock) CheckerCalls() []struct {
	Ctx   context.Context
	State *healthcheck.CheckState
} {
	var calls []struct {
		Ctx   context.Context
		State *healthcheck.CheckState
	}
	mock.lockChecker.RLock()
	calls = mock.calls.Checker
	mock.lockChecker.RUnlock()
	return calls
}

// ReadKey calls ReadKeyFunc.
func (mock *VaultClienterMock) ReadKey(path string, key string) (string, error) {
	if mock.ReadKeyFunc == nil {
		panic("VaultClienterMock.ReadKeyFunc: method is nil but VaultClienter.ReadKey was just called")
	}
	callInfo := struct {
		Path string
		Key  string
	}{
		Path: path,
		Key:  key,
	}
	mock.lockReadKey.Lock()
	mock.calls.ReadKey = append(mock.calls.ReadKey, callInfo)
	mock.lockReadKey.Unlock()
	return mock.ReadKeyFunc(path, key)
}

// ReadKeyCalls gets all the calls that were made to ReadKey.
// Check the length with:
//     len(mockedVaultClienter.ReadKeyCalls())
func (mock *VaultClienterMock) ReadKeyCalls() []struct {
	Path string
	Key  string
} {
	var calls []struct {
		Path string
		Key  string
	}
	mock.lockReadKey.RLock()
	calls = mock.calls.ReadKey
	mock.lockReadKey.RUnlock()
	return calls
}

// WriteKey calls WriteKeyFunc.
func (mock *VaultClienterMock) WriteKey(path string, key string, value string) error {
	if mock.WriteKeyFunc == nil {
		panic("VaultClienterMock.WriteKeyFunc: method is nil but VaultClienter.WriteKey was just called")
	}
	callInfo := struct {
		Path  string
		Key   string
		Value string
	}{
		Path:  path,
		Key:   key,
		Value: value,
	}
	mock.lockWriteKey.Lock()
	mock.calls.WriteKey = append(mock.calls.WriteKey, callInfo)
	mock.lockWriteKey.Unlock()
	return mock.WriteKeyFunc(path, key, value)
}

// WriteKeyCalls gets all the calls that were made to WriteKey.
// Check the length with:
//     len(mockedVaultClienter.WriteKeyCalls())
func (mock *VaultClienterMock) WriteKeyCalls() []struct {
	Path  string
	Key   string
	Value string
} {
	var calls []struct {
		Path  string
		Key   string
		Value string
	}
	mock.lockWriteKey.RLock()
	calls = mock.calls.WriteKey
	mock.lockWriteKey.RUnlock()
	return calls
}