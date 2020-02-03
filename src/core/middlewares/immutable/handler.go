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
	"errors"
	"fmt"
	"github.com/goharbor/harbor/src/common/utils/log"
	"github.com/goharbor/harbor/src/core/middlewares/interceptor"
	middlerware_err "github.com/goharbor/harbor/src/core/middlewares/util/error"
	internal_errors "github.com/goharbor/harbor/src/internal/error"
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
		pkgE := internal_errors.New(fmt.Errorf("error occurred when to handle request in immutable handler: %v", err)).WithCode(internal_errors.GeneralCode)
		msg := internal_errors.NewErrs(pkgE).Error()
		http.Error(rw, msg, http.StatusInternalServerError)
		return
	}

	if interceptor == nil {
		rh.next.ServeHTTP(rw, req)
		return
	}

	if err := interceptor.HandleRequest(req); err != nil {
		log.Warningf("Error occurred when to handle request in immutable handler: %v", err)
		var e *middlerware_err.ErrImmutable
		if errors.As(err, &e) {
			pkgE := internal_errors.New(e).WithCode(internal_errors.PreconditionCode)
			msg := internal_errors.NewErrs(pkgE).Error()
			http.Error(rw, msg, http.StatusPreconditionFailed)
			return
		}

		pkgE := internal_errors.New(fmt.Errorf("error occurred when to handle request in immutable handler: %v", err)).WithCode(internal_errors.GeneralCode)
		msg := internal_errors.NewErrs(pkgE).Error()
		http.Error(rw, msg, http.StatusInternalServerError)
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
