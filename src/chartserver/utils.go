package chartserver

import (
	"encoding/json"
	"net/http"
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
