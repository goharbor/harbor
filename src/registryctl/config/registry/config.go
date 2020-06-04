package registry

import (
	"fmt"
	"github.com/docker/distribution/configuration"
	storagedriver "github.com/docker/distribution/registry/storage/driver"
	"os"
)

// StorageDriver the storage driver bases on the registry configurations, like filesystem, oss, gcs, S3, and etc.
var StorageDriver storagedriver.StorageDriver

// ResolveConfiguration loads the mounted registry configuration file, which is shared with the registry controller.
func ResolveConfiguration(configPath string) (*configuration.Configuration, error) {
	fp, err := os.Open(configPath)
	if err != nil {
		return nil, err
	}

	defer fp.Close()

	config, err := configuration.Parse(fp)
	if err != nil {
		return nil, fmt.Errorf("error parsing %s: %v", configPath, err)
	}

	return config, nil
}
