package chartserver

import (
	"errors"
	"fmt"
	"net/url"
	"os"
	"strings"

	hlog "github.com/goharbor/harbor/src/common/utils/log"
)

const (
	userName    = "chart_controller"
	passwordKey = "UI_SECRET"
)

//Credential keeps the username and password for the basic auth
type Credential struct {
	Username string
	Password string
}

//Controller is used to handle flows of related requests based on the corresponding handlers
//A reverse proxy will be created and managed to proxy the related traffics between API and
//backend chart server
type Controller struct {
	//The access endpoint of the backend chart repository server
	backendServerAddr *url.URL

	//To cover the server info and status requests
	baseHandler *BaseHandler

	//To cover the chart repository requests
	repositoryHandler *RepositoryHandler

	//To cover all the manipulation requests
	manipulationHandler *ManipulationHandler

	//To cover the other utility requests
	utilityHandler *UtilityHandler
}

//NewController is constructor of the chartserver.Controller
func NewController(backendServer *url.URL) (*Controller, error) {
	if backendServer == nil {
		return nil, errors.New("failed to create chartserver.Controller: backend sever address is required")
	}

	//Try to create credential
	cred := &Credential{
		Username: userName,
		Password: os.Getenv(passwordKey),
	}

	//Use customized reverse proxy
	proxy := NewProxyEngine(backendServer, cred)

	//Create http client with customized timeouts
	client := NewChartClient(cred)

	//Initialize chart operator for use
	operator := &ChartOperator{}

	//Creat cache
	cacheCfg, err := getCacheConfig()
	if err != nil {
		//just log the error
		//will not break the whole flow if failed to create cache
		hlog.Errorf("failed to get cache configuration with error: %s", err)
	}
	cache := NewChartCache(cacheCfg)
	if !cache.IsEnabled() {
		hlog.Info("No cache is enabled for chart caching")
	}

	return &Controller{
		backendServerAddr: backendServer,
		baseHandler:       &BaseHandler{proxy},
		repositoryHandler: &RepositoryHandler{
			trafficProxy:         proxy,
			apiClient:            client,
			backendServerAddress: backendServer,
		},
		manipulationHandler: &ManipulationHandler{
			trafficProxy:         proxy,
			chartOperator:        operator,
			apiClient:            client,
			backendServerAddress: backendServer,
			chartCache:           cache,
		},
		utilityHandler: &UtilityHandler{
			apiClient:            client,
			backendServerAddress: backendServer,
			chartOperator:        operator,
		},
	}, nil
}

//GetBaseHandler returns the reference of BaseHandler
func (c *Controller) GetBaseHandler() *BaseHandler {
	return c.baseHandler
}

//GetRepositoryHandler returns the reference of RepositoryHandler
func (c *Controller) GetRepositoryHandler() *RepositoryHandler {
	return c.repositoryHandler
}

//GetManipulationHandler returns the reference of ManipulationHandler
func (c *Controller) GetManipulationHandler() *ManipulationHandler {
	return c.manipulationHandler
}

//GetUtilityHandler returns the reference of UtilityHandler
func (c *Controller) GetUtilityHandler() *UtilityHandler {
	return c.utilityHandler
}

//What's the cache driver if it is set
func parseCacheDriver() (string, bool) {
	driver, ok := os.LookupEnv(cacheDriverENVKey)
	return strings.ToLower(driver), ok
}

//Get and parse the configuration for the chart cache
func getCacheConfig() (*ChartCacheConfig, error) {
	driver, isSet := parseCacheDriver()
	if !isSet {
		return nil, nil
	}

	if driver != cacheDriverMem && driver != cacheDriverRedis {
		return nil, fmt.Errorf("cache driver '%s' is not supported, only support 'memory' and 'redis'", driver)
	}

	if driver == cacheDriverMem {
		return &ChartCacheConfig{
			DriverType: driver,
		}, nil
	}

	redisConfigV := os.Getenv(redisENVKey)
	redisCfg, err := parseRedisConfig(redisConfigV)
	if err != nil {
		return nil, fmt.Errorf("failed to parse redis configurations from '%s' with error: %s", redisCfg, err)
	}

	return &ChartCacheConfig{
		DriverType: driver,
		Config:     redisCfg,
	}, nil
}
