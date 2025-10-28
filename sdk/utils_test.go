package sdk

import (
	"io"
	"strings"
	"testing"

	"github.com/ONSdigital/dp-upload-service/api"
	. "github.com/smartystreets/goconvey/convey"
)

func TestUnmarshalJsonErrors(t *testing.T) {
	t.Parallel()

	Convey("Given a valid JsonErrors body", t, func() {
		body := `{"errors":[{"code":"DuplicateFile","description":"resource conflict: file already registered"}]}`
		rc := io.NopCloser(strings.NewReader(body))

		Convey("When unmarshalJsonErrors is called", func() {
			jsonErrors, err := unmarshalJsonErrors(rc)

			Convey("Then no error is returned", func() {
				So(err, ShouldBeNil)
			})

			Convey("Then the expected JsonErrors is returned", func() {
				expected := &api.JsonErrors{
					Error: []api.JsonError{
						{
							Code:        "DuplicateFile",
							Description: "resource conflict: file already registered",
						},
					},
				}
				So(jsonErrors, ShouldResemble, expected)
			})
		})
	})

	Convey("Given an invalid JsonErrors body", t, func() {
		body := `invalid json`
		rc := io.NopCloser(strings.NewReader(body))

		Convey("When unmarshalJsonErrors is called", func() {
			jsonErrors, err := unmarshalJsonErrors(rc)

			Convey("Then an error is returned", func() {
				So(err, ShouldNotBeNil)
			})

			Convey("And no JsonErrors is returned", func() {
				So(jsonErrors, ShouldBeNil)
			})
		})
	})

	Convey("Given a nil body", t, func() {
		Convey("When unmarshalJsonErrors is called", func() {
			jsonErrors, err := unmarshalJsonErrors(nil)

			Convey("Then no error is returned", func() {
				So(err, ShouldBeNil)
			})

			Convey("And no JsonErrors is returned", func() {
				So(jsonErrors, ShouldBeNil)
			})
		})
	})

	Convey("Given a reader that returns an error on read", t, func() {
		Convey("When unmarshalJsonErrors is called", func() {
			jsonErrors, err := unmarshalJsonErrors(brokenReader)

			Convey("Then the expected error is returned", func() {
				So(err, ShouldNotBeNil)
				So(err, ShouldEqual, expectedReadErr)
			})

			Convey("And no JsonErrors is returned", func() {
				So(jsonErrors, ShouldBeNil)
			})
		})
	})
}
