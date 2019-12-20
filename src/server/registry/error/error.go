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

package error

import (
	"github.com/goharbor/harbor/src/common/utils/log"
	serror "github.com/goharbor/harbor/src/server/error"
	"net/http"
)

// Handle generates the HTTP status code and error payload and writes them to the response
func Handle(w http.ResponseWriter, req *http.Request, err error) {
	log.Errorf("failed to handle the request %s: %v", req.URL.Path, err)
	statusCode, payload := serror.APIError(err)
	w.WriteHeader(statusCode)
	w.Write([]byte(payload))
}
