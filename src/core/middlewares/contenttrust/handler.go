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

package contenttrust

import (
	"github.com/goharbor/harbor/src/common/utils"
	"github.com/goharbor/harbor/src/common/utils/log"
	"github.com/goharbor/harbor/src/common/utils/notary"
	"github.com/goharbor/harbor/src/core/config"
	"github.com/goharbor/harbor/src/core/middlewares/util"
	"net/http"
)

// NotaryEndpoint ...
var NotaryEndpoint = ""

type contentTrustHandler struct {
	next http.Handler
}

// New ...
func New(next http.Handler) http.Handler {
	return &contentTrustHandler{
		next: next,
	}
}

// ServeHTTP ...
func (cth contentTrustHandler) ServeHTTP(rw http.ResponseWriter, req *http.Request) {
	imgRaw := req.Context().Value(util.ImageInfoCtxKey)
	if imgRaw == nil || !config.WithNotary() {
		cth.next.ServeHTTP(rw, req)
		return
	}
	img, _ := req.Context().Value(util.ImageInfoCtxKey).(util.ImageInfo)
	if img.Digest == "" {
		cth.next.ServeHTTP(rw, req)
		return
	}
	if scannerPull, ok := util.ScannerPullFromContext(req.Context()); ok && scannerPull {
		cth.next.ServeHTTP(rw, req)
		return
	}
	if !util.GetPolicyChecker().ContentTrustEnabled(img.ProjectName) {
		cth.next.ServeHTTP(rw, req)
		return
	}
	match, err := matchNotaryDigest(img)
	if err != nil {
		http.Error(rw, util.MarshalError("PROJECT_POLICY_VIOLATION", "Failed in communication with Notary please check the log"), http.StatusInternalServerError)
		return
	}
	if !match {
		log.Debugf("digest mismatch, failing the response.")
		http.Error(rw, util.MarshalError("PROJECT_POLICY_VIOLATION", "The image is not signed in Notary."), http.StatusPreconditionFailed)
		return
	}
	cth.next.ServeHTTP(rw, req)
}

func matchNotaryDigest(img util.ImageInfo) (bool, error) {
	if NotaryEndpoint == "" {
		NotaryEndpoint = config.InternalNotaryEndpoint()
	}
	targets, err := notary.GetInternalTargets(NotaryEndpoint, util.TokenUsername, img.Repository)
	if err != nil {
		return false, err
	}
	for _, t := range targets {
		if utils.IsDigest(img.Reference) {
			d, err := notary.DigestFromTarget(t)
			if err != nil {
				return false, err
			}
			if img.Digest == d {
				return true, nil
			}
		} else {
			if t.Tag == img.Reference {
				log.Debugf("found reference: %s in notary, try to match digest.", img.Reference)
				d, err := notary.DigestFromTarget(t)
				if err != nil {
					return false, err
				}
				if img.Digest == d {
					return true, nil
				}
			}
		}
	}
	log.Debugf("image: %#v, not found in notary", img)
	return false, nil
}
