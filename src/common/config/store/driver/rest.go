package driver

import (
	"github.com/goharbor/harbor/src/common/http"
	"github.com/goharbor/harbor/src/common/http/modifier"
)

// RESTDriver - config store driver based on REST API
type RESTDriver struct {
	coreURL string
	client  *http.Client
}

// NewRESTDriver - Create RESTDriver
func NewRESTDriver(coreURL string, modifiers ...modifier.Modifier) *RESTDriver {
	return &RESTDriver{coreURL: coreURL, client: http.NewClient(nil, modifiers...)}
}

// Load - load config data from REST server
func (h *RESTDriver) Load() (map[string]interface{}, error) {
	cfgMap := map[string]interface{}{}
	err := h.client.Get(h.coreURL, &cfgMap)
	return cfgMap, err
}

// Save - Save config data to REST server by PUT method
func (h *RESTDriver) Save(cfgMap map[string]interface{}) error {
	return h.client.Put(h.coreURL, cfgMap)
}
