package remote

import (
	"fmt"

	"github.com/goharbor/harbor/src/common/config"
	"github.com/goharbor/harbor/src/common/http"
	"github.com/goharbor/harbor/src/common/http/modifier/auth"
	"github.com/goharbor/harbor/src/common/models"
	"github.com/goharbor/harbor/src/common/utils/log"
)

// ConfigureDriver - use http://core:8080/api/configurations to manage configuration,
// commonly used outside core api container
type ConfigureDriver struct {
	config.ConfigureStore
	// ConfigURL -- URL of configure server
	ConfigURL  string
	httpClient *http.Client
}

// Client - only used for remote client
type Client interface {
	GetDatabaseCfg() (*models.Database, error)
	GetCfgs() (map[string]interface{}, error)
}

// Config contains configurations needed for client
type Config struct {
	Secret string
}

// NewRemoteConfigDriver ... Create a Remote Configure Driver
func NewRemoteConfigDriver(coreurl string, cfg *Config) (*ConfigureDriver, error) {
	remoteDriver := &ConfigureDriver{ConfigURL: coreurl}
	config.MetaData.InitMetaDataFromArray(config.ConfigList)
	if cfg != nil {
		authorizer := auth.NewSecretAuthorizer(cfg.Secret)
		log.Errorf("The jobservice secret is %v, url:%v", cfg.Secret, remoteDriver.ConfigURL)
		remoteDriver.httpClient = http.NewClient(nil, authorizer)
	} else {
		log.Error("Config is nil")
	}
	return remoteDriver, nil
}

// Load ... load configures from URL
func (cd *ConfigureDriver) Load() error {

	url := cd.ConfigURL + "/api/configs"

	cfgs := map[string]interface{}{}

	if err := cd.httpClient.Get(url, &cfgs); err != nil {
		return err
	}
	stringCfgs := map[string]string{}
	for k, v := range cfgs {
		stringCfgs[k] = fmt.Sprintf("%v", v)
	}
	// Get all configure entry from configure store
	cd.LoadFromMap(stringCfgs)
	return nil
}

// GetDatabaseCfg ... Get database configure
func (cd *ConfigureDriver) GetDatabaseCfg() (*models.Database, error) {
	if err := cd.Load(); err != nil {
		return nil, err
	}
	return cd.ConfigureStore.GetDatabaseCfg(), nil
}

// GetCfgs ...
func (cd *ConfigureDriver) GetCfgs() (map[string]interface{}, error) {
	if err := cd.Load(); err != nil {
		return nil, err
	}
	return cd.GetCfgs()
}
