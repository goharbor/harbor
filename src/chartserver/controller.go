package chartserver

import (
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"os"

	hlog "github.com/goharbor/harbor/src/common/utils/log"
)

const (
	userName    = "chart_controller"
	passwordKey = "CORE_SECRET"
)

// Credential keeps the username and password for the basic auth
type Credential struct {
	Username string
	Password string
}

// Controller is used to handle flows of related requests based on the corresponding handlers
// A reverse proxy will be created and managed to proxy the related traffics between API and
// backend chart server
type Controller struct {
	// Proxy used to to transfer the traffic of requests
	// It's mainly used to talk to the backend chart server
	trafficProxy *ProxyEngine

	// Parse and process the chart version to provide required info data
	chartOperator *ChartOperator

	// HTTP client used to call the realted APIs of the backend chart repositories
	apiClient *ChartClient

	// The access endpoint of the backend chart repository server
	backendServerAddress *url.URL

	// Cache the chart data
	chartCache *ChartCache
}

// NewController is constructor of the chartserver.Controller
func NewController(backendServer *url.URL, middlewares ...func(http.Handler) http.Handler) (*Controller, error) {
	if backendServer == nil {
		return nil, errors.New("failed to create chartserver.Controller: backend sever address is required")
	}

	// Try to create credential
	cred := &Credential{
		Username: userName,
		Password: os.Getenv(passwordKey),
	}

	// Creat cache
	cacheCfg, err := getCacheConfig()
	if err != nil {
		// just log the error
		// will not break the whole flow if failed to create cache
		hlog.Errorf("failed to get cache configuration with error: %s", err)
	}
	cache := NewChartCache(cacheCfg)
	if !cache.IsEnabled() {
		hlog.Info("No cache is enabled for chart caching")
	}

	return &Controller{
		backendServerAddress: backendServer,
		// Use customized reverse proxy
		trafficProxy: NewProxyEngine(backendServer, cred, middlewares...),
		// Initialize chart operator for use
		chartOperator: &ChartOperator{},
		// Create http client with customized timeouts
		apiClient:  NewChartClient(cred),
		chartCache: cache,
	}, nil
}

// APIPrefix returns the API prefix path of calling backend chart service.
func (c *Controller) APIPrefix(namespace string) string {
	return fmt.Sprintf("%s/api/%s/charts", c.backendServerAddress.String(), namespace)
}
