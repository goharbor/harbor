package registry

import (
	"fmt"
	"github.com/docker/distribution/configuration"
	storagedriver "github.com/docker/distribution/registry/storage/driver"
	"os"
)

var StorageDriver storagedriver.StorageDriver

// ResolveConfiguration ...
func ResolveConfiguration(conf string) (*configuration.Configuration, error) {
	fp, err := os.Open(conf)
	if err != nil {
		return nil, err
	}

	defer fp.Close()

	config, err := configuration.Parse(fp)
	if err != nil {
		return nil, fmt.Errorf("error parsing %s: %v", conf, err)
	}

	return config, nil
}
