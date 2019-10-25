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

package immutable

import (
	"fmt"
	"github.com/goharbor/harbor/src/common/utils/log"
	"github.com/goharbor/harbor/src/core/middlewares/interceptor"
	"github.com/goharbor/harbor/src/core/middlewares/util"
	middlerware_err "github.com/goharbor/harbor/src/core/middlewares/util/error"
	"net/http"
)

type immutableHandler struct {
	builders []interceptor.Builder
	next     http.Handler
}

// New ...
func New(next http.Handler, builders ...interceptor.Builder) http.Handler {
	if len(builders) == 0 {
		builders = defaultBuilders
	}

	return &immutableHandler{
		builders: builders,
		next:     next,
	}
}

// ServeHTTP ...
func (rh *immutableHandler) ServeHTTP(rw http.ResponseWriter, req *http.Request) {

	interceptor, err := rh.getInterceptor(req)
	if err != nil {
		log.Warningf("Error occurred when to handle request in immutable handler: %v", err)
		http.Error(rw, util.MarshalError("InternalError", fmt.Sprintf("Error occurred when to handle request in immutable handler: %v", err)),
			http.StatusInternalServerError)
		return
	}

	if interceptor == nil {
		rh.next.ServeHTTP(rw, req)
		return
	}

	if err := interceptor.HandleRequest(req); err != nil {
		log.Warningf("Error occurred when to handle request in immutable handler: %v", err)
		if _, ok := err.(middlerware_err.ErrImmutable); ok {
			http.Error(rw, util.MarshalError("DENIED",
				fmt.Sprintf("%v", err)), http.StatusPreconditionFailed)
			return
		}
		http.Error(rw, util.MarshalError("InternalError", fmt.Sprintf("Error occurred when to handle request in immutable handler: %v", err)),
			http.StatusInternalServerError)
		return
	}

	rh.next.ServeHTTP(rw, req)

	interceptor.HandleResponse(rw, req)
}

func (rh *immutableHandler) getInterceptor(req *http.Request) (interceptor.Interceptor, error) {
	for _, builder := range rh.builders {
		interceptor, err := builder.Build(req)
		if err != nil {
			return nil, err
		}

		if interceptor != nil {
			return interceptor, nil
		}
	}

	return nil, nil
}
