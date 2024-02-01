package config

import (
	"os"
	"testing"
	"time"

	. "github.com/smartystreets/goconvey/convey"
)

func TestConfig(t *testing.T) {
	os.Clearenv()
	testCfg, err := Get()
	Convey("Given an environment with no environment variables set", t, func() {
		Convey("When the config values are retrieved", func() {
			Convey("Then testCfg should not be nil", func() {
				So(testCfg, ShouldNotBeNil)
			})

			Convey("Then there should be no error returned", func() {
				So(err, ShouldBeNil)
			})

			Convey("Then the values should be set to the expected defaults", func() {
				So(testCfg.BindAddr, ShouldEqual, ":25100")
				So(testCfg.AwsRegion, ShouldEqual, "eu-west-2")
				So(testCfg.UploadBucketName, ShouldEqual, "deprecated")
				So(testCfg.StaticFilesEncryptedBucketName, ShouldEqual, "testing")
				So(testCfg.GracefulShutdownTimeout, ShouldEqual, 5*time.Second)
				So(testCfg.HealthCheckInterval, ShouldEqual, 30*time.Second)
				So(testCfg.HealthCheckCriticalTimeout, ShouldEqual, 90*time.Second)
				So(testCfg.ServiceAuthToken, ShouldEqual, "c60198e9-1864-4b68-ad0b-1e858e5b46a4")
			})

			Convey("Then a second call to config should return the same config", func() {
				newCfg, newErr := Get()
				So(newErr, ShouldBeNil)
				So(newCfg, ShouldResemble, testCfg)
			})
		})
	})
}
