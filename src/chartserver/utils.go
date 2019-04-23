package chartserver

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/url"
	"os"
	"strings"
)

const (
	contentTypeHeader = "content-type"
	contentTypeJSON   = "application/json"
)

// Extract error object '{"error": "****---***"}' from the content if existing
// nil error will be returned if it does exist
func extractError(content []byte) (text string, err error) {
	if len(content) == 0 {
		return "", nil
	}

	errorObj := make(map[string]string)
	err = json.Unmarshal(content, &errorObj)
	if err != nil {
		return "", err
	}

	if errText, ok := errorObj["error"]; ok {
		return errText, nil
	}

	return "", nil
}

// Parse the redis configuration to the beego cache pattern
// Config pattern is "address:port[,weight,password,db_index]"
func parseRedisConfig(redisConfigV string) (string, error) {
	if len(redisConfigV) == 0 {
		return "", errors.New("empty redis config")
	}

	redisConfig := make(map[string]string)
	redisConfig["key"] = cacheCollectionName

	// Try best to parse the configuration segments.
	// If the related parts are missing, assign default value.
	// The default database index for UI process is 0.
	configSegments := strings.Split(redisConfigV, ",")
	for i, segment := range configSegments {
		if i > 3 {
			// ignore useless segments
			break
		}

		switch i {
		// address:port
		case 0:
			redisConfig["conn"] = segment
		// password, may not exist
		case 2:
			redisConfig["password"] = segment
		// database index, may not exist
		case 3:
			redisConfig["dbNum"] = segment
		}
	}

	// Assign default value
	if len(redisConfig["dbNum"]) == 0 {
		redisConfig["dbNum"] = "0"
	}

	// Try to validate the connection address
	fullAddr := redisConfig["conn"]
	if strings.Index(fullAddr, "://") == -1 {
		// Append schema
		fullAddr = fmt.Sprintf("redis://%s", fullAddr)
	}
	// Validate it by url
	_, err := url.Parse(fullAddr)
	if err != nil {
		return "", err
	}

	// Convert config map to string
	cfgData, err := json.Marshal(redisConfig)
	if err != nil {
		return "", err
	}

	return string(cfgData), nil
}

// What's the cache driver if it is set
func parseCacheDriver() (string, bool) {
	driver, ok := os.LookupEnv(cacheDriverENVKey)
	return strings.ToLower(driver), ok
}

// Get and parse the configuration for the chart cache
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
