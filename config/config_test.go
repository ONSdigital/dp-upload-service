package config

import (
	"os"
	"testing"
	"time"

	. "github.com/smartystreets/goconvey/convey"
)

func TestConfigNil(t *testing.T) {
	Convey("Given an environment with no environment variables set", t, func() {
		Convey("Then cfg should be nil", func() {
			So(cfg, ShouldBeNil)
		})
	})
}

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
				So(testCfg.BindAddr, ShouldEqual, "localhost:25100")
				So(testCfg.AwsRegion, ShouldEqual, "eu-west-1")
				So(testCfg.UploadBucketName, ShouldEqual, "dp-frontend-florence-file-uploads")
				So(testCfg.EncryptionDisabled, ShouldBeFalse)
				So(testCfg.GracefulShutdownTimeout, ShouldEqual, 5*time.Second)
				So(testCfg.HealthCheckInterval, ShouldEqual, 30*time.Second)
				So(testCfg.HealthCheckCriticalTimeout, ShouldEqual, 90*time.Second)
				So(testCfg.VaultAddress, ShouldEqual, "http://localhost:8200")
				So(testCfg.VaultPath, ShouldEqual, "secret/shared/psk")
				So(testCfg.VaultToken, ShouldEqual, "")
			})

			Convey("Then a second call to config should return the same config", func() {
				newCfg, newErr := Get()
				So(newErr, ShouldBeNil)
				So(newCfg, ShouldResemble, testCfg)
			})
		})
	})
}
