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

package quota

import (
	"errors"
	"fmt"
	"net/http"

	"github.com/goharbor/harbor/src/common/utils/log"
	"github.com/goharbor/harbor/src/internal"
	"github.com/goharbor/harbor/src/pkg/types"
	serror "github.com/goharbor/harbor/src/server/error"
	"github.com/goharbor/harbor/src/server/middleware"
)

var (
	errNonSuccess = errors.New("non success status code")
)

// RequestConfig request resources middleware config
type RequestConfig struct {
	// ReferenceObject returns reference object which resources will be requested
	ReferenceObject func(r *http.Request) (reference string, referenceID string, err error)

	// Resources returns request resources for the reference object
	Resources func(r *http.Request, reference, referenceID string) (types.ResourceList, error)
}

// RequestMiddleware middleware which request resources
func RequestMiddleware(config RequestConfig, skippers ...middleware.Skipper) func(http.Handler) http.Handler {
	return middleware.New(func(w http.ResponseWriter, r *http.Request, next http.Handler) {
		logger := log.G(r.Context()).WithFields(log.Fields{"middleware": "quota", "action": "request", "url": r.URL.Path})

		if config.ReferenceObject == nil || config.Resources == nil {
			serror.SendError(w, fmt.Errorf("invald config the for middleware"))
			return
		}

		reference, referenceID, err := config.ReferenceObject(r)
		if err != nil {
			logger.Errorf("get reference object failed, error: %v", err)

			serror.SendError(w, err)
			return
		}

		enabled, err := quotaController.IsEnabled(r.Context(), reference, referenceID)
		if err != nil {
			logger.Errorf("check whether quota enabled for %s %s failed, error: %v", reference, referenceID, err)
			serror.SendError(w, err)
			return
		}

		if !enabled {
			// quota is disabled for the reference object, so direct to next handler
			logger.Infof("quota is disabled for %s %s, so direct to next handler", reference, referenceID)
			next.ServeHTTP(w, r)
			return
		}

		resources, err := config.Resources(r, reference, referenceID)
		if err != nil {
			logger.Errorf("get resources failed, error: %v", err)

			serror.SendError(w, err)
			return
		}

		if len(resources) == 0 {
			// no resources request for this http request, so direct to next handler
			logger.Info("no resources request for this http request, so direct to next handler")
			next.ServeHTTP(w, r)
			return
		}

		res, ok := w.(*internal.ResponseBuffer)
		if !ok {
			res = internal.NewResponseBuffer(w)
			defer res.Flush()
		}

		err = quotaController.Request(r.Context(), reference, referenceID, resources, func() error {
			next.ServeHTTP(res, r)
			if !res.Success() {
				return errNonSuccess
			}

			return nil
		})

		if err != nil && err != errNonSuccess {
			res.Reset()
			serror.SendError(res, err)
		}

	}, skippers...)
}

// RefreshConfig refresh quota usage middleware config
type RefreshConfig struct {
	// ReferenceObject returns reference object its quota usage will refresh by reference and reference id
	ReferenceObject func(*http.Request) (reference string, referenceID string, err error)
}

// RefreshMiddleware middleware which refresh the quota usage after the response success
func RefreshMiddleware(config RefreshConfig, skipers ...middleware.Skipper) func(http.Handler) http.Handler {
	return middleware.AfterResponse(func(w http.ResponseWriter, r *http.Request, statusCode int) error {
		// skip to refresh quota usage when response is not success
		if !isSuccess(statusCode) {
			return nil
		}

		if config.ReferenceObject == nil {
			return fmt.Errorf("invald config the for middleware")
		}

		logger := log.G(r.Context()).WithFields(log.Fields{"middleware": "quota", "action": "refresh", "url": r.URL.Path})

		reference, referenceID, err := config.ReferenceObject(r)
		if err != nil {
			logger.Errorf("get reference object to refresh quota usage failed, error: %v", err)
			return err
		}

		enabled, err := quotaController.IsEnabled(r.Context(), reference, referenceID)
		if err != nil {
			logger.Errorf("check whether quota enabled for %s %s failed, error: %v", reference, referenceID, err)
			return err
		}

		if !enabled {
			logger.Infof("quota is disabled for %s %s, so return directly", reference, referenceID)
			return nil
		}

		if err = quotaController.Refresh(r.Context(), reference, referenceID); err != nil {
			logger.Errorf("refresh quota for %s %s failed, error: %v", reference, referenceID, err)
			return err
		}

		return nil
	}, skipers...)
}

func isSuccess(statusCode int) bool {
	return statusCode >= http.StatusOK && statusCode < http.StatusBadRequest
}
