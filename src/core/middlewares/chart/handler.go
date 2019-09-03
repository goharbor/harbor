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

package chart

import (
	"fmt"
	"net/http"

	"github.com/goharbor/harbor/src/common/quota"
	"github.com/goharbor/harbor/src/common/utils/log"
	"github.com/goharbor/harbor/src/core/middlewares/interceptor"
	"github.com/goharbor/harbor/src/core/middlewares/util"
)

type chartHandler struct {
	builders []interceptor.Builder
	next     http.Handler
}

// New ...
func New(next http.Handler, builders ...interceptor.Builder) http.Handler {
	if len(builders) == 0 {
		builders = defaultBuilders
	}

	return &chartHandler{
		builders: builders,
		next:     next,
	}
}

// ServeHTTP manifest ...
func (h *chartHandler) ServeHTTP(rw http.ResponseWriter, req *http.Request) {
	interceptor, err := h.getInterceptor(req)
	if err != nil {
		http.Error(rw, fmt.Sprintf("Error occurred when to handle request in chart count quota handler: %v", err),
			http.StatusInternalServerError)
		return
	}

	if interceptor == nil {
		h.next.ServeHTTP(rw, req)
		return
	}

	if err := interceptor.HandleRequest(req); err != nil {
		log.Warningf("Error occurred when to handle request in count quota handler: %v", err)
		if _, ok := err.(quota.Errors); ok {
			http.Error(rw, fmt.Sprintf("Quota exceeded when processing the request of %v", err), http.StatusForbidden)
			return
		}
		http.Error(rw, fmt.Sprintf("Error occurred when to handle request in chart count quota handler: %v", err),
			http.StatusInternalServerError)
		return
	}

	w := util.NewCustomResponseWriter(rw)
	h.next.ServeHTTP(w, req)

	interceptor.HandleResponse(w, req)
}

func (h *chartHandler) getInterceptor(req *http.Request) (interceptor.Interceptor, error) {
	for _, builder := range h.builders {
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
