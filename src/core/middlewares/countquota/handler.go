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

package countquota

import (
	"fmt"
	"github.com/goharbor/harbor/src/common/utils/log"
	"github.com/goharbor/harbor/src/core/middlewares/util"
	"net/http"
)

type countQuotaHandler struct {
	next http.Handler
}

// New ...
func New(next http.Handler) http.Handler {
	return &countQuotaHandler{
		next: next,
	}
}

// ServeHTTP manifest ...
func (cqh *countQuotaHandler) ServeHTTP(rw http.ResponseWriter, req *http.Request) {
	countInteceptor := getInteceptor(req)
	if countInteceptor == nil {
		cqh.next.ServeHTTP(rw, req)
		return
	}
	// handler request
	if err := countInteceptor.HandleRequest(req); err != nil {
		log.Warningf("Error occurred when to handle request in count quota handler: %v", err)
		http.Error(rw, util.MarshalError("InternalError", fmt.Sprintf("Error occurred when to handle request in count quota handler: %v", err)),
			http.StatusInternalServerError)
		return
	}
	cqh.next.ServeHTTP(rw, req)

	// handler response
	countInteceptor.HandleResponse(*rw.(*util.CustomResponseWriter), req)
}

func getInteceptor(req *http.Request) util.RegInterceptor {
	// PUT /v2/<name>/manifests/<reference>
	matchPushMF, repository, tag := util.MatchPushManifest(req)
	if matchPushMF {
		mfInfo := util.MfInfo{}
		mfInfo.Repository = repository
		mfInfo.Tag = tag
		return NewPutManifestInterceptor(&mfInfo)
	}
	return nil
}
