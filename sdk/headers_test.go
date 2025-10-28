package sdk

import (
	"net/http"
	"testing"

	dprequest "github.com/ONSdigital/dp-net/v3/request"
	. "github.com/smartystreets/goconvey/convey"
)

const (
	testServiceAuthToken = "testServiceAuthToken"
)

func TestHeaders_Add(t *testing.T) {
	t.Parallel()

	Convey("Given a Headers struct with a ServiceAuthToken", t, func() {
		headers := &Headers{
			ServiceAuthToken: testServiceAuthToken,
		}

		Convey("When Add is called on an http.Request", func() {
			req := &http.Request{Header: http.Header{}}
			headers.Add(req)

			Convey("Then the Authorization header is added to the request", func() {
				So(req.Header.Get(dprequest.AuthHeaderKey), ShouldEqual, dprequest.BearerPrefix+testServiceAuthToken)
			})
		})
	})

	Convey("Given an empty Headers struct", t, func() {
		headers := &Headers{}

		Convey("When Add is called on an http.Request", func() {
			req := &http.Request{Header: http.Header{}}
			headers.Add(req)

			Convey("Then the Authorization header is not added to the request", func() {
				So(req.Header.Get(dprequest.AuthHeaderKey), ShouldEqual, "")
			})
		})
	})
}
