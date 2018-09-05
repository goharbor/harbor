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

// WriteError writes error to http client
func WriteError(w http.ResponseWriter, code int, err error) {
	errorObj := make(map[string]string)
	errorObj["error"] = err.Error()
	errorContent, errorMarshal := json.Marshal(errorObj)
	if errorMarshal != nil {
		errorContent = []byte(err.Error())
	}
	w.WriteHeader(code)
	w.Write(errorContent)
}

// WriteInternalError writes error with statusCode == 500
func WriteInternalError(w http.ResponseWriter, err error) {
	WriteError(w, http.StatusInternalServerError, err)
}

// Write JSON data to http client
func writeJSONData(w http.ResponseWriter, data []byte) {
	w.Header().Set(contentTypeHeader, contentTypeJSON)
	w.WriteHeader(http.StatusOK)
	w.Write(data)
}

// Extract error object '{"error": "****---***"}' from the content if existing
// nil error will be returned if it does exist
func extractError(content []byte) error {
	if len(content) == 0 {
		return nil
	}

	errorObj := make(map[string]string)
	err := json.Unmarshal(content, &errorObj)
	if err != nil {
		return nil
	}

	if errText, ok := errorObj["error"]; ok {
		return errors.New(errText)
	}

	return nil
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
