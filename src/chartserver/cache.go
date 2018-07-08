package chartserver

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/url"
	"os"
	"strings"
	"time"

	beego_cache "github.com/astaxie/beego/cache"
	hlog "github.com/vmware/harbor/src/common/utils/log"

	//Enable redis cache adaptor
	_ "github.com/astaxie/beego/cache/redis"
)

const (
	standardExpireTime  = 3600 * time.Second
	redisENVKey         = "_REDIS_URL"
	cacheDriverENVKey   = "CHART_CACHE_DRIVER" //"memory" or "redis"
	cacheDriverMem      = "memory"
	cacheDriverRedis    = "redis"
	cacheCollectionName = "helm_chart_cache"
)

//ChartCache is designed to cache some processed data for repeated accessing
//to improve the performance
type ChartCache struct {
	//cache driver
	cache beego_cache.Cache

	//Flag to indicate if cache driver is configured
	isEnabled bool

	//Backend driver type
	driverType string
}

//NewChartCache is constructor of ChartCache
func NewChartCache() *ChartCache {
	driverType, isSet := parseCacheDriver()

	chartCache := &ChartCache{
		isEnabled: isSet,
	}
	if !chartCache.isEnabled {
		hlog.Info("No cache driver is configured, chart cache will be disabled")
		return chartCache
	}

	cache, enabledDriverType := initCacheDriver(driverType)
	chartCache.cache = cache
	chartCache.driverType = enabledDriverType

	return chartCache
}

//IsEnabled to indicate if the chart cache is configured
func (chc *ChartCache) IsEnabled() bool {
	return chc.isEnabled
}

//PutChart caches the detailed data of chart version
func (chc *ChartCache) PutChart(chart *ChartVersionDetails) {
	if !chc.isEnabled {
		return
	}

	//As it's a valid json data anymore when retrieving back from redis cache,
	//here we use separate methods to handle the data according to the driver type
	if chart != nil {
		var err error

		switch chc.driverType {
		case cacheDriverMem:
			//Directly put object in
			err = chc.cache.Put(chart.Metadata.Digest, chart, standardExpireTime)
		case cacheDriverRedis:
			//Marshal to json data before saving
			var jsonData []byte
			if jsonData, err = json.Marshal(chart); err == nil {
				err = chc.cache.Put(chart.Metadata.Digest, jsonData, standardExpireTime)
			}
		default:
			//Should not reach here, but still put guard code here
			err = chc.cache.Put(chart.Metadata.Digest, chart, standardExpireTime)
		}

		if err != nil {
			//Just logged
			hlog.Errorf("Failed to cache chart object with error: %s\n", err)
			hlog.Warningf("If cache driver is using 'redis', please check the related configurations or the network connection")
		}
	}
}

//GetChart trys to retrieve it from the cache
//If hit, return the cached item;
//otherwise, nil object is returned
func (chc *ChartCache) GetChart(chartDigest string) *ChartVersionDetails {
	if !chc.isEnabled {
		return nil
	}

	object := chc.cache.Get(chartDigest)
	if object != nil {
		//Try to convert data
		//First try the normal way
		if chartDetails, ok := object.(*ChartVersionDetails); ok {
			return chartDetails
		}

		//Maybe json bytes
		if bytes, yes := object.([]byte); yes {
			chartDetails := &ChartVersionDetails{}
			err := json.Unmarshal(bytes, chartDetails)
			if err == nil {
				return chartDetails
			}
			//Just logged the error
			hlog.Errorf("Failed to retrieve chart from cache with error: %s", err)
		}
	}

	return nil
}

//What's the cache driver if it is set
func parseCacheDriver() (string, bool) {
	driver, ok := os.LookupEnv(cacheDriverENVKey)
	return strings.ToLower(driver), ok
}

//Initialize the cache driver based on the config
func initCacheDriver(driverType string) (beego_cache.Cache, string) {
	switch driverType {
	case cacheDriverMem:
		hlog.Info("Enable memory cache for chart caching")
		return beego_cache.NewMemoryCache(), cacheDriverMem
	case cacheDriverRedis:
		redisConfig, err := parseRedisConfig()
		if err != nil {
			//Just logged
			hlog.Errorf("Failed to read redis configurations with error: %s", err)
			break
		}

		redisCache, err := beego_cache.NewCache(cacheDriverRedis, redisConfig)
		if err != nil {
			//Just logged
			hlog.Errorf("Failed to initialize redis cache: %s", err)
			break
		}

		hlog.Info("Enable reids cache for chart caching")
		return redisCache, cacheDriverRedis
	default:
		break
	}

	hlog.Info("Driver type %s is not suppotred, enable memory cache by default for chart caching")
	//Any other cases, use memory cache
	return beego_cache.NewMemoryCache(), cacheDriverMem
}

//Parse the redis configuration to the beego cache pattern
//Config pattern is "address:port[,weight,password,db_index]"
func parseRedisConfig() (string, error) {
	redisConfigV := os.Getenv(redisENVKey)
	if len(redisConfigV) == 0 {
		return "", errors.New("empty redis config")
	}

	redisConfig := make(map[string]string)
	redisConfig["key"] = cacheCollectionName

	//The full pattern
	if strings.Index(redisConfigV, ",") != -1 {
		//Read only the previous 4 segments
		configSegments := strings.SplitN(redisConfigV, ",", 4)
		if len(configSegments) != 4 {
			return "", errors.New("invalid redis config, it should be address:port[,weight,password,db_index]")
		}

		redisConfig["conn"] = configSegments[0]
		redisConfig["password"] = configSegments[2]
		redisConfig["dbNum"] = configSegments[3]
	} else {
		//The short pattern
		redisConfig["conn"] = redisConfigV
		redisConfig["dbNum"] = "0"
		redisConfig["password"] = ""
	}

	//Try to validate the connection address
	fullAddr := redisConfig["conn"]
	if strings.Index(fullAddr, "://") == -1 {
		//Append schema
		fullAddr = fmt.Sprintf("redis://%s", fullAddr)
	}
	//Validate it by url
	_, err := url.Parse(fullAddr)
	if err != nil {
		return "", err
	}

	//Convert config map to string
	cfgData, err := json.Marshal(redisConfig)
	if err != nil {
		return "", err
	}

	return string(cfgData), nil
}
