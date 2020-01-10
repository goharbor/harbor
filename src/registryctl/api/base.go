// Copyright Project Harbor Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//    http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package api

import (
	"encoding/json"
	"errors"
	"github.com/goharbor/harbor/src/jobservice/logger"
	"net/http"
)

func HandleInternalServerError(w http.ResponseWriter, req *http.Request) {
	handleError(w, req, http.StatusInternalServerError, errors.New("internalerror"))
}

func HandleForbidden(w http.ResponseWriter, req *http.Request) {
	handleError(w, req, http.StatusForbidden, errors.New("forbidden"))
}

func HandleBadRequest(w http.ResponseWriter, req *http.Request, err error) {
	handleError(w, req, http.StatusBadRequest, err)
}

func handleError(w http.ResponseWriter, req *http.Request, code int, err error) {
	logger.Errorf("Serve http request '%s %s' error: %d %s", req.Method, req.URL.String(), code, err.Error())
	w.WriteHeader(code)
	writeDate(w, []byte(err.Error()))
}

func writeDate(w http.ResponseWriter, bytes []byte) {
	if _, err := w.Write(bytes); err != nil {
		logger.Errorf("writer write error: %s", err)
	}
}

// WriteJSON response status code will be written automatically if there is an error
func WriteJSON(w http.ResponseWriter, req *http.Request, v interface{}) error {
	b, err := json.Marshal(v)
	if err != nil {
		HandleInternalServerError(w, req)
		return err
	}

	if _, err = w.Write(b); err != nil {
		return err
	}
	return nil
}
