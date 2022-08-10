// Code generated by moq; DO NOT EDIT.
// github.com/matryer/moq

package mock

import (
	"context"
	"github.com/ONSdigital/dp-api-clients-go/v2/files"
	"github.com/ONSdigital/dp-upload-service/upload"
	"sync"
)

// Ensure, that FilesClienterMock does implement upload.FilesClienter.
// If this is not the case, regenerate this file with moq.
var _ upload.FilesClienter = &FilesClienterMock{}

// FilesClienterMock is a mock implementation of upload.FilesClienter.
//
// 	func TestSomethingThatUsesFilesClienter(t *testing.T) {
//
// 		// make and configure a mocked upload.FilesClienter
// 		mockedFilesClienter := &FilesClienterMock{
// 			MarkFileUploadedFunc: func(ctx context.Context, path string, etag string) error {
// 				panic("mock out the MarkFileUploaded method")
// 			},
// 			RegisterFileFunc: func(ctx context.Context, metadata files.FileMetaData) error {
// 				panic("mock out the RegisterFile method")
// 			},
// 		}
//
// 		// use mockedFilesClienter in code that requires upload.FilesClienter
// 		// and then make assertions.
//
// 	}
type FilesClienterMock struct {
	// MarkFileUploadedFunc mocks the MarkFileUploaded method.
	MarkFileUploadedFunc func(ctx context.Context, path string, etag string) error

	// RegisterFileFunc mocks the RegisterFile method.
	RegisterFileFunc func(ctx context.Context, metadata files.FileMetaData) error

	// calls tracks calls to the methods.
	calls struct {
		// MarkFileUploaded holds details about calls to the MarkFileUploaded method.
		MarkFileUploaded []struct {
			// Ctx is the ctx argument value.
			Ctx context.Context
			// Path is the path argument value.
			Path string
			// Etag is the etag argument value.
			Etag string
		}
		// RegisterFile holds details about calls to the RegisterFile method.
		RegisterFile []struct {
			// Ctx is the ctx argument value.
			Ctx context.Context
			// Metadata is the metadata argument value.
			Metadata files.FileMetaData
		}
	}
	lockMarkFileUploaded sync.RWMutex
	lockRegisterFile     sync.RWMutex
}

// MarkFileUploaded calls MarkFileUploadedFunc.
func (mock *FilesClienterMock) MarkFileUploaded(ctx context.Context, path string, etag string) error {
	if mock.MarkFileUploadedFunc == nil {
		panic("FilesClienterMock.MarkFileUploadedFunc: method is nil but FilesClienter.MarkFileUploaded was just called")
	}
	callInfo := struct {
		Ctx  context.Context
		Path string
		Etag string
	}{
		Ctx:  ctx,
		Path: path,
		Etag: etag,
	}
	mock.lockMarkFileUploaded.Lock()
	mock.calls.MarkFileUploaded = append(mock.calls.MarkFileUploaded, callInfo)
	mock.lockMarkFileUploaded.Unlock()
	return mock.MarkFileUploadedFunc(ctx, path, etag)
}

// MarkFileUploadedCalls gets all the calls that were made to MarkFileUploaded.
// Check the length with:
//     len(mockedFilesClienter.MarkFileUploadedCalls())
func (mock *FilesClienterMock) MarkFileUploadedCalls() []struct {
	Ctx  context.Context
	Path string
	Etag string
} {
	var calls []struct {
		Ctx  context.Context
		Path string
		Etag string
	}
	mock.lockMarkFileUploaded.RLock()
	calls = mock.calls.MarkFileUploaded
	mock.lockMarkFileUploaded.RUnlock()
	return calls
}

// RegisterFile calls RegisterFileFunc.
func (mock *FilesClienterMock) RegisterFile(ctx context.Context, metadata files.FileMetaData) error {
	if mock.RegisterFileFunc == nil {
		panic("FilesClienterMock.RegisterFileFunc: method is nil but FilesClienter.RegisterFile was just called")
	}
	callInfo := struct {
		Ctx      context.Context
		Metadata files.FileMetaData
	}{
		Ctx:      ctx,
		Metadata: metadata,
	}
	mock.lockRegisterFile.Lock()
	mock.calls.RegisterFile = append(mock.calls.RegisterFile, callInfo)
	mock.lockRegisterFile.Unlock()
	return mock.RegisterFileFunc(ctx, metadata)
}

// RegisterFileCalls gets all the calls that were made to RegisterFile.
// Check the length with:
//     len(mockedFilesClienter.RegisterFileCalls())
func (mock *FilesClienterMock) RegisterFileCalls() []struct {
	Ctx      context.Context
	Metadata files.FileMetaData
} {
	var calls []struct {
		Ctx      context.Context
		Metadata files.FileMetaData
	}
	mock.lockRegisterFile.RLock()
	calls = mock.calls.RegisterFile
	mock.lockRegisterFile.RUnlock()
	return calls
}
