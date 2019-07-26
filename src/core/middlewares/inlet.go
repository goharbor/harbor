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

package middlewares

import (
	"errors"
	"github.com/goharbor/harbor/src/core/middlewares/registryproxy"
	"github.com/goharbor/harbor/src/core/middlewares/util"
	"net/http"
)

var head http.Handler

// Init initialize the Proxy instance and handler chain.
func Init() error {
	ph := registryproxy.New()
	if ph == nil {
		return errors.New("get nil when to create proxy")
	}
	handlerChain := New(Middlewares).Create()
	head = handlerChain.Then(ph)
	return nil
}

// Handle handles the request.
func Handle(rw http.ResponseWriter, req *http.Request) {
	customResW := util.NewCustomResponseWriter(rw)
	head.ServeHTTP(customResW, req)
}
