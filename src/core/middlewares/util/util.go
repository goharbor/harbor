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

package util

import (
	"encoding/json"
	"github.com/docker/distribution"
	"github.com/goharbor/harbor/src/common/models"
	"github.com/goharbor/harbor/src/common/quota"
	"github.com/goharbor/harbor/src/common/utils/clair"
	"github.com/goharbor/harbor/src/common/utils/log"
	common_redis "github.com/goharbor/harbor/src/common/utils/redis"
	"github.com/goharbor/harbor/src/core/config"
	"github.com/goharbor/harbor/src/core/promgr"
	"github.com/goharbor/harbor/src/pkg/scan/whitelist"
	"net/http"
	"net/http/httptest"
	"regexp"
	"strings"
)

type contextKey string

const (
	manifestURLPattern = `^/v2/((?:[a-z0-9]+(?:[._-][a-z0-9]+)*/)+)manifests/([\w][\w.:-]{0,127})`
	blobURLPattern     = `^/v2/((?:[a-z0-9]+(?:[._-][a-z0-9]+)*/)+)blobs/uploads/`
	// ImageInfoCtxKey the context key for image information
	ImageInfoCtxKey = contextKey("ImageInfo")
	// TokenUsername ...
	// TODO: temp solution, remove after vmware/harbor#2242 is resolved.
	TokenUsername = "harbor-core"
	// MFInfokKey the context key for image tag redis lock
	MFInfokKey = contextKey("ManifestLock")
)

// ImageInfo ...
type ImageInfo struct {
	Repository  string
	Reference   string
	ProjectName string
	Digest      string
}

// MfInfo ...
type MfInfo struct {
	// basic information of a manifest
	ProjectID  int64
	Repository string
	Tag        string
	Digest     string

	// Exist is to index the existing of the manifest in DB. If false, it's an new image for uploading.
	Exist bool
	// DigestChanged true means the manifest exists but digest is changed.
	// Probably it's a new image with existing repo/tag name or overwrite.
	DigestChanged bool

	// used to block multiple push on same image.
	TagLock    *common_redis.Mutex
	Refrerence []distribution.Descriptor

	// Quota is the resource applied for the manifest upload request.
	Quota *quota.ResourceList
}

// JSONError wraps a concrete Code and Message, it's readable for docker deamon.
type JSONError struct {
	Code    string `json:"code,omitempty"`
	Message string `json:"message,omitempty"`
	Detail  string `json:"detail,omitempty"`
}

// MarshalError ...
func MarshalError(code, msg string) string {
	var tmpErrs struct {
		Errors []JSONError `json:"errors,omitempty"`
	}
	tmpErrs.Errors = append(tmpErrs.Errors, JSONError{
		Code:    code,
		Message: msg,
		Detail:  msg,
	})
	str, err := json.Marshal(tmpErrs)
	if err != nil {
		log.Debugf("failed to marshal json error, %v", err)
		return msg
	}
	return string(str)
}

// MatchManifestURL ...
func MatchManifestURL(req *http.Request) (bool, string, string) {
	re, err := regexp.Compile(manifestURLPattern)
	if err != nil {
		log.Errorf("error to match manifest url, %v", err)
		return false, "", ""
	}
	s := re.FindStringSubmatch(req.URL.Path)
	if len(s) == 3 {
		s[1] = strings.TrimSuffix(s[1], "/")
		return true, s[1], s[2]
	}
	return false, "", ""
}

// MatchPutBlobURL ...
func MatchPutBlobURL(req *http.Request) (bool, string) {
	if req.Method != http.MethodPut {
		return false, ""
	}
	re, err := regexp.Compile(blobURLPattern)
	if err != nil {
		log.Errorf("error to match put blob url, %v", err)
		return false, ""
	}
	s := re.FindStringSubmatch(req.URL.Path)
	if len(s) == 2 {
		s[1] = strings.TrimSuffix(s[1], "/")
		return true, s[1]
	}
	return false, ""
}

// MatchPullManifest checks if the request looks like a request to pull manifest.  If it is returns the image and tag/sha256 digest as 2nd and 3rd return values
func MatchPullManifest(req *http.Request) (bool, string, string) {
	if req.Method != http.MethodGet {
		return false, "", ""
	}
	return MatchManifestURL(req)
}

// MatchPushManifest checks if the request looks like a request to push manifest.  If it is returns the image and tag/sha256 digest as 2nd and 3rd return values
func MatchPushManifest(req *http.Request) (bool, string, string) {
	if req.Method != http.MethodPut {
		return false, "", ""
	}
	return MatchManifestURL(req)
}

// CopyResp ...
func CopyResp(rec *httptest.ResponseRecorder, rw http.ResponseWriter) {
	for k, v := range rec.Header() {
		rw.Header()[k] = v
	}
	rw.WriteHeader(rec.Result().StatusCode)
	rw.Write(rec.Body.Bytes())
}

// PolicyChecker checks the policy of a project by project name, to determine if it's needed to check the image's status under this project.
type PolicyChecker interface {
	// contentTrustEnabled returns whether a project has enabled content trust.
	ContentTrustEnabled(name string) bool
	// vulnerablePolicy  returns whether a project has enabled vulnerable, and the project's severity.
	VulnerablePolicy(name string) (bool, models.Severity, models.CVEWhitelist)
}

// PmsPolicyChecker ...
type PmsPolicyChecker struct {
	pm promgr.ProjectManager
}

// ContentTrustEnabled ...
func (pc PmsPolicyChecker) ContentTrustEnabled(name string) bool {
	project, err := pc.pm.Get(name)
	if err != nil {
		log.Errorf("Unexpected error when getting the project, error: %v", err)
		return true
	}
	return project.ContentTrustEnabled()
}

// VulnerablePolicy ...
func (pc PmsPolicyChecker) VulnerablePolicy(name string) (bool, models.Severity, models.CVEWhitelist) {
	project, err := pc.pm.Get(name)
	wl := models.CVEWhitelist{}
	if err != nil {
		log.Errorf("Unexpected error when getting the project, error: %v", err)
		return true, models.SevUnknown, wl
	}
	mgr := whitelist.NewDefaultManager()
	if project.ReuseSysCVEWhitelist() {
		w, err := mgr.GetSys()
		if err != nil {
			return project.VulPrevented(), clair.ParseClairSev(project.Severity()), wl
		}
		wl = *w
	} else {
		w, err := mgr.Get(project.ProjectID)
		if err != nil {
			return project.VulPrevented(), clair.ParseClairSev(project.Severity()), wl
		}
		wl = *w
	}
	return project.VulPrevented(), clair.ParseClairSev(project.Severity()), wl

}

// NewPMSPolicyChecker returns an instance of an pmsPolicyChecker
func NewPMSPolicyChecker(pm promgr.ProjectManager) PolicyChecker {
	return &PmsPolicyChecker{
		pm: pm,
	}
}

// GetPolicyChecker ...
func GetPolicyChecker() PolicyChecker {
	return NewPMSPolicyChecker(config.GlobalProjectMgr)
}
