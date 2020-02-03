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
	"net/http"

	"github.com/goharbor/harbor/src/common/security"
	"github.com/goharbor/harbor/src/common/utils/log"
	"github.com/goharbor/harbor/src/core/middlewares/registryproxy"
	"github.com/goharbor/harbor/src/core/middlewares/util"
)

var head http.Handler
var proxy http.Handler

// Init initialize the Proxy instance and handler chain.
func Init() error {
	proxy = registryproxy.New()
	if proxy == nil {
		return errors.New("get nil when to create proxy")
	}
	return nil
}

// Handle handles the request.
func Handle(rw http.ResponseWriter, req *http.Request) {
	securityCtx, ok := security.FromContext(req.Context())
	if !ok {
		log.Errorf("failed to get security context in middlerware")
		// error to get security context, use the default chain.
		head = New(Middlewares).Create().Then(proxy)
	} else {
		// true: the request is from 127.0.0.1, only quota middlewares are applied to request
		// false: the request is from outside, all of middlewares are applied to the request.
		if securityCtx.IsSolutionUser() {
			head = New(MiddlewaresLocal).Create().Then(proxy)
		} else {
			head = New(Middlewares).Create().Then(proxy)
		}
	}

	customResW := util.NewCustomResponseWriter(rw)
	head.ServeHTTP(customResW, req)
}
