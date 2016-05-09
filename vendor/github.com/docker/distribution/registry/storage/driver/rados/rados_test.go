// +build include_rados

package rados

import (
	"os"
	"testing"

	storagedriver "github.com/docker/distribution/registry/storage/driver"
	"github.com/docker/distribution/registry/storage/driver/testsuites"

	"gopkg.in/check.v1"
)

// Hook up gocheck into the "go test" runner.
func Test(t *testing.T) { check.TestingT(t) }

func init() {
	poolname := os.Getenv("RADOS_POOL")
	username := os.Getenv("RADOS_USER")

	driverConstructor := func() (storagedriver.StorageDriver, error) {
		parameters := DriverParameters{
			poolname,
			username,
			defaultChunkSize,
		}

		return New(parameters)
	}

	skipCheck := func() string {
		if poolname == "" {
			return "RADOS_POOL must be set to run Rado tests"
		}
		return ""
	}

	testsuites.RegisterSuite(driverConstructor, skipCheck)
}
