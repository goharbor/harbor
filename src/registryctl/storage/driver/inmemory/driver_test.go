package inmemory

import (
	"testing"

	storagedriver "github.com/goharbor/harbor/src/registryctl/storage/driver"
	"github.com/goharbor/harbor/src/registryctl/storage/driver/testsuites"
	"gopkg.in/check.v1"
)

// Hook up gocheck into the "go test" runner.
func Test(t *testing.T) { check.TestingT(t) }

func init() {
	inmemoryDriverConstructor := func() (storagedriver.StorageDriver, error) {
		return New(), nil
	}
	testsuites.RegisterSuite(inmemoryDriverConstructor, testsuites.NeverSkip)
}
