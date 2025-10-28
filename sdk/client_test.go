package sdk

import (
	"context"
	"net/http"
	"testing"
	"time"

	"github.com/ONSdigital/dp-api-clients-go/v2/health"
	"github.com/ONSdigital/dp-healthcheck/healthcheck"
	dphttp "github.com/ONSdigital/dp-net/v3/http"
	. "github.com/smartystreets/goconvey/convey"
)

const (
	uploadServiceURL = "http://localhost:25100"
)

func newMockClienter(r *http.Response, err error) *dphttp.ClienterMock {
	return &dphttp.ClienterMock{
		SetPathsWithNoRetriesFunc: func(_ []string) {
		},
		DoFunc: func(_ context.Context, _ *http.Request) (*http.Response, error) {
			return r, err
		},
		GetPathsWithNoRetriesFunc: func() []string {
			return []string{"/health"}
		},
	}
}

func newMockUploadServiceClient(mockClienter *dphttp.ClienterMock) *Client {
	return NewWithHealthClient(health.NewClientWithClienter(serviceName, uploadServiceURL, mockClienter))
}

func TestNew(t *testing.T) {
	t.Parallel()

	Convey("Given an upload service Client created with New", t, func() {
		client := New(uploadServiceURL)

		Convey("When URL() is called", func() {
			url := client.URL()

			Convey("Then the correct URL is returned", func() {
				So(url, ShouldEqual, uploadServiceURL)
			})
		})

		Convey("When Health() is called", func() {
			healthClient := client.Health()

			Convey("Then the correct health.Client is returned", func() {
				So(healthClient.URL, ShouldEqual, uploadServiceURL)
				So(healthClient.Name, ShouldEqual, serviceName)
			})
		})

		Convey("When Checker() is called", func() {
			timeBeforeCheck := time.Now()
			checkState := healthcheck.NewCheckState(serviceName)
			err := client.Checker(context.Background(), checkState)

			Convey("Then no error is returned", func() {
				So(err, ShouldBeNil)
			})

			Convey("And the check state is updated correctly", func() {
				// This test is to check that the .Checker function works as expected.
				// The "CRITICAL" status and connection refused message are expected since the service is not running.
				So(checkState.Name(), ShouldEqual, serviceName)
				So(checkState.Status(), ShouldEqual, healthcheck.StatusCritical)
				So(checkState.StatusCode(), ShouldEqual, 0)
				So(checkState.Message(), ShouldContainSubstring, "connect: connection refused")
				So(*checkState.LastChecked(), ShouldHappenAfter, timeBeforeCheck)
				So(*checkState.LastFailure(), ShouldHappenAfter, timeBeforeCheck)
				So(checkState.LastSuccess(), ShouldBeNil)
			})
		})
	})
}

func TestNewWithHealthClient(t *testing.T) {
	t.Parallel()

	Convey("Given an upload service Client created with an existing health.Client using NewWithHealthClient", t, func() {
		mockClienter := newMockClienter(&http.Response{StatusCode: http.StatusOK}, nil)
		client := newMockUploadServiceClient(mockClienter)

		Convey("When URL() is called", func() {
			url := client.URL()

			Convey("Then the correct URL is returned", func() {
				So(url, ShouldEqual, uploadServiceURL)
			})
		})

		Convey("When Health() is called", func() {
			healthClient := client.Health()

			Convey("Then the correct health.Client is returned", func() {
				So(healthClient.URL, ShouldEqual, uploadServiceURL)
				So(healthClient.Name, ShouldEqual, serviceName)
				So(healthClient.Client, ShouldEqual, mockClienter)
			})
		})

		Convey("When Checker() is called", func() {
			timeBeforeCheck := time.Now()
			checkState := healthcheck.NewCheckState(serviceName)
			err := client.Checker(context.Background(), checkState)

			Convey("Then no error is returned", func() {
				So(err, ShouldBeNil)
			})

			Convey("And the check state is updated correctly", func() {
				So(checkState.Name(), ShouldEqual, serviceName)
				So(checkState.Status(), ShouldEqual, healthcheck.StatusOK)
				So(checkState.StatusCode(), ShouldEqual, 200)
				So(checkState.Message(), ShouldEqual, serviceName+health.StatusMessage[healthcheck.StatusOK])
				So(*checkState.LastChecked(), ShouldHappenAfter, timeBeforeCheck)
				So(*checkState.LastSuccess(), ShouldHappenAfter, timeBeforeCheck)
				So(checkState.LastFailure(), ShouldBeNil)
			})
		})
	})
}
