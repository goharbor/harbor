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
	"net/http"
)

func handleInternalServerError(w http.ResponseWriter) {
	http.Error(w, http.StatusText(http.StatusInternalServerError),
		http.StatusInternalServerError)
}

func handleUnauthorized(w http.ResponseWriter) {
	http.Error(w, http.StatusText(http.StatusUnauthorized),
		http.StatusUnauthorized)
}

// response status code will be written automatically if there is an error
func writeJSON(w http.ResponseWriter, v interface{}) error {
	b, err := json.Marshal(v)
	if err != nil {
		handleInternalServerError(w)
		return err
	}

	if _, err = w.Write(b); err != nil {
		return err
	}
	return nil
}
