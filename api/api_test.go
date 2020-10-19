package api

import (
	"context"
	"net/http/httptest"
	"testing"

	"github.com/gorilla/mux"
	. "github.com/smartystreets/goconvey/convey"
)

func TestSetup(t *testing.T) {
	Convey("Given an API instance", t, func() {
		r := &mux.Router{}
		ctx := context.Background()
		api := Setup(ctx, &VaultClienterMock{}, r, &S3ClienterMock{})

		Convey("When created the following routes should have been added", func() {
			So(api, ShouldNotBeNil)
			// Replace the check below with any newly added api endpoints
			So(hasRoute(api.Router, "/hello", "GET"), ShouldBeTrue)
		})
	})
}

func hasRoute(r *mux.Router, path, method string) bool {
	req := httptest.NewRequest(method, path, nil)
	match := &mux.RouteMatch{}
	return r.Match(req, match)
}
