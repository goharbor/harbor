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
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/docker/distribution"
	"github.com/garyburd/redigo/redis"
	"github.com/goharbor/harbor/src/common/dao"
	"github.com/goharbor/harbor/src/common/models"
	"github.com/goharbor/harbor/src/common/quota"
	"github.com/goharbor/harbor/src/common/utils/clair"
	"github.com/goharbor/harbor/src/common/utils/log"
	common_redis "github.com/goharbor/harbor/src/common/utils/redis"
	"github.com/goharbor/harbor/src/core/config"
	"github.com/goharbor/harbor/src/core/promgr"
	"github.com/goharbor/harbor/src/pkg/scan/whitelist"
)

type contextKey string

// ErrRequireQuota ...
var ErrRequireQuota = errors.New("cannot get quota on project for request")

const (
	manifestURLPattern = `^/v2/((?:[a-z0-9]+(?:[._-][a-z0-9]+)*/)+)manifests/([\w][\w.:-]{0,127})`
	blobURLPattern     = `^/v2/((?:[a-z0-9]+(?:[._-][a-z0-9]+)*/)+)blobs/uploads/`

	chartVersionInfoKey = contextKey("ChartVersionInfo")

	// ImageInfoCtxKey the context key for image information
	ImageInfoCtxKey = contextKey("ImageInfo")
	// TokenUsername ...
	// TODO: temp solution, remove after vmware/harbor#2242 is resolved.
	TokenUsername = "harbor-core"
	// MFInfokKey the context key for image tag redis lock
	MFInfokKey = contextKey("ManifestInfo")
	// BBInfokKey the context key for image tag redis lock
	BBInfokKey = contextKey("BlobInfo")

	// DialConnectionTimeout ...
	DialConnectionTimeout = 30 * time.Second
	// DialReadTimeout ...
	DialReadTimeout = time.Minute + 10*time.Second
	// DialWriteTimeout ...
	DialWriteTimeout = 10 * time.Second
)

// ChartVersionInfo ...
type ChartVersionInfo struct {
	ProjectID int64
	Namespace string
	ChartName string
	Version   string
}

// ImageInfo ...
type ImageInfo struct {
	Repository  string
	Reference   string
	ProjectName string
	Digest      string
}

// BlobInfo ...
type BlobInfo struct {
	UUID        string
	ProjectID   int64
	ContentType string
	Size        int64
	Repository  string
	Tag         string

	// Exist is to index the existing of the manifest in DB. If false, it's an new image for uploading.
	Exist bool

	Digest     string
	DigestLock *common_redis.Mutex
	// Quota is the resource applied for the manifest upload request.
	Quota *quota.ResourceList
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

	// ArtifactID is the ID of the artifact which query by repository and tag
	ArtifactID int64

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

// MatchPatchBlobURL ...
func MatchPatchBlobURL(req *http.Request) (bool, string) {
	if req.Method != http.MethodPatch {
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

// MatchMountBlobURL POST /v2/<name>/blobs/uploads/?mount=<digest>&from=<repository name>
// If match, will return repo, mount and from as the 2nd, 3th and 4th.
func MatchMountBlobURL(req *http.Request) (bool, string, string, string) {
	if req.Method != http.MethodPost {
		return false, "", "", ""
	}
	re, err := regexp.Compile(blobURLPattern)
	if err != nil {
		log.Errorf("error to match post blob url, %v", err)
		return false, "", "", ""
	}
	s := re.FindStringSubmatch(req.URL.Path)
	if len(s) == 2 {
		s[1] = strings.TrimSuffix(s[1], "/")
		mount := req.FormValue("mount")
		if mount == "" {
			return false, "", "", ""
		}
		from := req.FormValue("from")
		if from == "" {
			return false, "", "", ""
		}
		return true, s[1], mount, from
	}
	return false, "", "", ""
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

// TryRequireQuota ...
func TryRequireQuota(projectID int64, quotaRes *quota.ResourceList) error {
	quotaMgr, err := quota.NewManager("project", strconv.FormatInt(projectID, 10))
	if err != nil {
		log.Errorf("Error occurred when to new quota manager %v", err)
		return err
	}
	if err := quotaMgr.AddResources(*quotaRes); err != nil {
		log.Errorf("cannot get quota for the project resource: %d, err: %v", projectID, err)
		return ErrRequireQuota
	}
	return nil
}

// TryFreeQuota used to release resource for failure case
func TryFreeQuota(projectID int64, qres *quota.ResourceList) bool {
	quotaMgr, err := quota.NewManager("project", strconv.FormatInt(projectID, 10))
	if err != nil {
		log.Errorf("Error occurred when to new quota manager %v", err)
		return false
	}

	if err := quotaMgr.SubtractResources(*qres); err != nil {
		log.Errorf("cannot release quota for the project resource: %d, err: %v", projectID, err)
		return false
	}
	return true
}

// GetBlobSize blob size with UUID in redis
func GetBlobSize(conn redis.Conn, uuid string) (int64, error) {
	exists, err := redis.Int(conn.Do("EXISTS", uuid))
	if err != nil {
		return 0, err
	}
	if exists == 1 {
		size, err := redis.Int64(conn.Do("GET", uuid))
		if err != nil {
			return 0, err
		}
		return size, nil
	}
	return 0, nil
}

// SetBunkSize sets the temp size for blob bunk with its uuid.
func SetBunkSize(conn redis.Conn, uuid string, size int64) (bool, error) {
	setRes, err := redis.String(conn.Do("SET", uuid, size))
	if err != nil {
		return false, err
	}
	return setRes == "OK", nil
}

// GetProjectID ...
func GetProjectID(name string) (int64, error) {
	project, err := dao.GetProjectByName(name)
	if err != nil {
		return 0, err
	}
	if project != nil {
		return project.ProjectID, nil
	}
	return 0, fmt.Errorf("project %s is not found", name)
}

// GetRegRedisCon ...
func GetRegRedisCon() (redis.Conn, error) {
	// FOR UT
	if os.Getenv("UTTEST") == "true" {
		return redis.Dial(
			"tcp",
			fmt.Sprintf("%s:%d", os.Getenv("REDIS_HOST"), 6379),
			redis.DialConnectTimeout(DialConnectionTimeout),
			redis.DialReadTimeout(DialReadTimeout),
			redis.DialWriteTimeout(DialWriteTimeout),
		)
	}
	return redis.DialURL(
		config.GetRedisOfRegURL(),
		redis.DialConnectTimeout(DialConnectionTimeout),
		redis.DialReadTimeout(DialReadTimeout),
		redis.DialWriteTimeout(DialWriteTimeout),
	)
}

// BlobInfoFromContext returns blob info from context
func BlobInfoFromContext(ctx context.Context) (*BlobInfo, bool) {
	info, ok := ctx.Value(BBInfokKey).(*BlobInfo)
	return info, ok
}

// ChartVersionInfoFromContext returns chart info from context
func ChartVersionInfoFromContext(ctx context.Context) (*ChartVersionInfo, bool) {
	info, ok := ctx.Value(chartVersionInfoKey).(*ChartVersionInfo)
	return info, ok
}

// ImageInfoFromContext returns image info from context
func ImageInfoFromContext(ctx context.Context) (*ImageInfo, bool) {
	info, ok := ctx.Value(ImageInfoCtxKey).(*ImageInfo)
	return info, ok
}

// ManifestInfoFromContext returns manifest info from context
func ManifestInfoFromContext(ctx context.Context) (*MfInfo, bool) {
	info, ok := ctx.Value(MFInfokKey).(*MfInfo)
	return info, ok
}

// NewBlobInfoContext returns context with blob info
func NewBlobInfoContext(ctx context.Context, info *BlobInfo) context.Context {
	return context.WithValue(ctx, BBInfokKey, info)
}

// NewChartVersionInfoContext returns context with blob info
func NewChartVersionInfoContext(ctx context.Context, info *ChartVersionInfo) context.Context {
	return context.WithValue(ctx, chartVersionInfoKey, info)
}

// NewImageInfoContext returns context with image info
func NewImageInfoContext(ctx context.Context, info *ImageInfo) context.Context {
	return context.WithValue(ctx, ImageInfoCtxKey, info)
}

// NewManifestInfoContext returns context with manifest info
func NewManifestInfoContext(ctx context.Context, info *MfInfo) context.Context {
	return context.WithValue(ctx, MFInfokKey, info)
}
