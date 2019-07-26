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

package vulnerable

import (
	"fmt"
	"github.com/goharbor/harbor/src/common/utils/log"
	"github.com/goharbor/harbor/src/core/config"
	"github.com/goharbor/harbor/src/core/middlewares/util"
	"github.com/goharbor/harbor/src/pkg/scan"
	"net/http"
)

type vulnerableHandler struct {
	next http.Handler
}

// New ...
func New(next http.Handler) http.Handler {
	return &vulnerableHandler{
		next: next,
	}
}

// ServeHTTP ...
func (vh vulnerableHandler) ServeHTTP(rw http.ResponseWriter, req *http.Request) {
	imgRaw := req.Context().Value(util.ImageInfoCtxKey)
	if imgRaw == nil || !config.WithClair() {
		vh.next.ServeHTTP(rw, req)
		return
	}
	img, _ := req.Context().Value(util.ImageInfoCtxKey).(util.ImageInfo)
	if img.Digest == "" {
		vh.next.ServeHTTP(rw, req)
		return
	}
	projectVulnerableEnabled, projectVulnerableSeverity, wl := util.GetPolicyChecker().VulnerablePolicy(img.ProjectName)
	if !projectVulnerableEnabled {
		vh.next.ServeHTTP(rw, req)
		return
	}
	vl, err := scan.VulnListByDigest(img.Digest)
	if err != nil {
		log.Errorf("Failed to get the vulnerability list, error: %v", err)
		http.Error(rw, util.MarshalError("PROJECT_POLICY_VIOLATION", "Failed to get vulnerabilities."), http.StatusPreconditionFailed)
		return
	}
	filtered := vl.ApplyWhitelist(wl)
	msg := vh.filterMsg(img, filtered)
	log.Info(msg)
	if int(vl.Severity()) >= int(projectVulnerableSeverity) {
		log.Debugf("the image severity: %q is higher then project setting: %q, failing the response.", vl.Severity(), projectVulnerableSeverity)
		http.Error(rw, util.MarshalError("PROJECT_POLICY_VIOLATION", fmt.Sprintf("The severity of vulnerability of the image: %q is equal or higher than the threshold in project setting: %q.", vl.Severity(), projectVulnerableSeverity)), http.StatusPreconditionFailed)
		return
	}
	vh.next.ServeHTTP(rw, req)
}

func (vh vulnerableHandler) filterMsg(img util.ImageInfo, filtered scan.VulnerabilityList) string {
	filterMsg := fmt.Sprintf("Image: %s/%s:%s, digest: %s, vulnerabilities fitered by whitelist:", img.ProjectName, img.Repository, img.Reference, img.Digest)
	if len(filtered) == 0 {
		filterMsg = fmt.Sprintf("%s none.", filterMsg)
	}
	for _, v := range filtered {
		filterMsg = fmt.Sprintf("%s ID: %s, severity: %s;", filterMsg, v.ID, v.Severity)
	}
	return filterMsg
}
