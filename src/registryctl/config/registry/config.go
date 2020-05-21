package registry

import (
	"fmt"
	"github.com/docker/distribution/configuration"
	storagedriver "github.com/docker/distribution/registry/storage/driver"
	"os"
)

const DefaultRegConf = "/etc/registry/config.yml"

var StorageDriver storagedriver.StorageDriver

// ResolveConfiguration ...
func ResolveConfiguration(configPath ...string) (*configuration.Configuration, error) {
	if len(configPath) == 0 {
		configPath[0] = DefaultRegConf
	}
	fp, err := os.Open(configPath[0])
	if err != nil {
		return nil, err
	}

	defer fp.Close()

	config, err := configuration.Parse(fp)
	if err != nil {
		return nil, fmt.Errorf("error parsing %s: %v", DefaultRegConf, err)
	}

	return config, nil
}
