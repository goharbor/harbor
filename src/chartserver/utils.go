package chartserver

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strings"
)

const (
	contentTypeHeader = "content-type"
	contentTypeJSON   = "application/json"
)

//Write error to http client
func writeError(w http.ResponseWriter, code int, err error) {
	errorObj := make(map[string]string)
	errorObj["error"] = err.Error()
	errorContent, _ := json.Marshal(errorObj)

	w.WriteHeader(code)
	w.Write(errorContent)
}

//StatusCode == 500
func writeInternalError(w http.ResponseWriter, err error) {
	writeError(w, http.StatusInternalServerError, err)
}

//Write JSON data to http client
func writeJSONData(w http.ResponseWriter, data []byte) {
	w.Header().Set(contentTypeHeader, contentTypeJSON)
	w.WriteHeader(http.StatusOK)
	w.Write(data)
}

//Parse the redis configuration to the beego cache pattern
//Config pattern is "address:port[,weight,password,db_index]"
func parseRedisConfig(redisConfigV string) (string, error) {
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
