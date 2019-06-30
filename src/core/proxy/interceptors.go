package proxy

import (
	"encoding/json"
	"github.com/goharbor/harbor/src/common/dao"
	"github.com/goharbor/harbor/src/common/models"
	"github.com/goharbor/harbor/src/common/utils/clair"
	"github.com/goharbor/harbor/src/common/utils/log"
	"github.com/goharbor/harbor/src/common/utils/notary"
	"github.com/goharbor/harbor/src/core/config"
	"github.com/goharbor/harbor/src/core/promgr"
	coreutils "github.com/goharbor/harbor/src/core/utils"
	"github.com/goharbor/harbor/src/pkg/scan"

	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"regexp"
	"strconv"
	"strings"
)

type contextKey string

const (
	manifestURLPattern = `^/v2/((?:[a-z0-9]+(?:[._-][a-z0-9]+)*/)+)manifests/([\w][\w.:-]{0,127})`
	catalogURLPattern  = `/v2/_catalog`
	imageInfoCtxKey    = contextKey("ImageInfo")
	// TODO: temp solution, remove after vmware/harbor#2242 is resolved.
	tokenUsername = "harbor-core"
)

// Record the docker deamon raw response.
var rec *httptest.ResponseRecorder

// NotaryEndpoint , exported for testing.
var NotaryEndpoint = ""

// MatchPullManifest checks if the request looks like a request to pull manifest.  If it is returns the image and tag/sha256 digest as 2nd and 3rd return values
func MatchPullManifest(req *http.Request) (bool, string, string) {
	// TODO: add user agent check.
	if req.Method != http.MethodGet {
		return false, "", ""
	}
	return matchManifestURL(req)
}

// MatchPushManifest checks if the request looks like a request to push manifest.  If it is returns the image and tag/sha256 digest as 2nd and 3rd return values
func MatchPushManifest(req *http.Request) (bool, string, string) {
	if req.Method != http.MethodPut {
		return false, "", ""
	}
	return matchManifestURL(req)
}

func matchManifestURL(req *http.Request) (bool, string, string) {
	re := regexp.MustCompile(manifestURLPattern)
	s := re.FindStringSubmatch(req.URL.Path)
	if len(s) == 3 {
		s[1] = strings.TrimSuffix(s[1], "/")
		return true, s[1], s[2]
	}
	return false, "", ""
}

// MatchListRepos checks if the request looks like a request to list repositories.
func MatchListRepos(req *http.Request) bool {
	if req.Method != http.MethodGet {
		return false
	}
	re := regexp.MustCompile(catalogURLPattern)
	s := re.FindStringSubmatch(req.URL.Path)
	if len(s) == 1 {
		return true
	}
	return false
}

// policyChecker checks the policy of a project by project name, to determine if it's needed to check the image's status under this project.
type policyChecker interface {
	// contentTrustEnabled returns whether a project has enabled content trust.
	contentTrustEnabled(name string) bool
	// vulnerablePolicy  returns whether a project has enabled vulnerable, and the project's severity.
	vulnerablePolicy(name string) (bool, models.Severity)
}

type pmsPolicyChecker struct {
	pm promgr.ProjectManager
}

func (pc pmsPolicyChecker) contentTrustEnabled(name string) bool {
	project, err := pc.pm.Get(name)
	if err != nil {
		log.Errorf("Unexpected error when getting the project, error: %v", err)
		return true
	}
	return project.ContentTrustEnabled()
}
func (pc pmsPolicyChecker) vulnerablePolicy(name string) (bool, models.Severity) {
	project, err := pc.pm.Get(name)
	if err != nil {
		log.Errorf("Unexpected error when getting the project, error: %v", err)
		return true, models.SevUnknown
	}
	return project.VulPrevented(), clair.ParseClairSev(project.Severity())
}

// newPMSPolicyChecker returns an instance of an pmsPolicyChecker
func newPMSPolicyChecker(pm promgr.ProjectManager) policyChecker {
	return &pmsPolicyChecker{
		pm: pm,
	}
}

func getPolicyChecker() policyChecker {
	return newPMSPolicyChecker(config.GlobalProjectMgr)
}

type imageInfo struct {
	repository  string
	reference   string
	projectName string
	digest      string
}

type urlHandler struct {
	next http.Handler
}

func (uh urlHandler) ServeHTTP(rw http.ResponseWriter, req *http.Request) {
	log.Debugf("in url handler, path: %s", req.URL.Path)
	flag, repository, reference := MatchPullManifest(req)
	if flag {
		components := strings.SplitN(repository, "/", 2)
		if len(components) < 2 {
			http.Error(rw, marshalError("PROJECT_POLICY_VIOLATION", fmt.Sprintf("Bad repository name: %s", repository)), http.StatusBadRequest)
			return
		}

		client, err := coreutils.NewRepositoryClientForUI(tokenUsername, repository)
		if err != nil {
			log.Errorf("Error creating repository Client: %v", err)
			http.Error(rw, marshalError("PROJECT_POLICY_VIOLATION", fmt.Sprintf("Failed due to internal Error: %v", err)), http.StatusInternalServerError)
			return
		}
		digest, _, err := client.ManifestExist(reference)
		if err != nil {
			log.Errorf("Failed to get digest for reference: %s, error: %v", reference, err)
			http.Error(rw, marshalError("PROJECT_POLICY_VIOLATION", fmt.Sprintf("Failed due to internal Error: %v", err)), http.StatusInternalServerError)
			return
		}

		img := imageInfo{
			repository:  repository,
			reference:   reference,
			projectName: components[0],
			digest:      digest,
		}

		log.Debugf("image info of the request: %#v", img)
		ctx := context.WithValue(req.Context(), imageInfoCtxKey, img)
		req = req.WithContext(ctx)
	}
	uh.next.ServeHTTP(rw, req)
}

type readonlyHandler struct {
	next http.Handler
}

func (rh readonlyHandler) ServeHTTP(rw http.ResponseWriter, req *http.Request) {
	if config.ReadOnly() {
		if req.Method == http.MethodDelete || req.Method == http.MethodPost || req.Method == http.MethodPatch || req.Method == http.MethodPut {
			log.Warningf("The request is prohibited in readonly mode, url is: %s", req.URL.Path)
			http.Error(rw, marshalError("DENIED", "The system is in read only mode. Any modification is prohibited."), http.StatusForbidden)
			return
		}
	}
	rh.next.ServeHTTP(rw, req)
}

type multipleManifestHandler struct {
	next http.Handler
}

// The handler is responsible for blocking request to upload manifest list by docker client, which is not supported so far by Harbor.
func (mh multipleManifestHandler) ServeHTTP(rw http.ResponseWriter, req *http.Request) {
	match, _, _ := MatchPushManifest(req)
	if match {
		contentType := req.Header.Get("Content-type")
		// application/vnd.docker.distribution.manifest.list.v2+json
		if strings.Contains(contentType, "manifest.list.v2") {
			log.Debugf("Content-type: %s is not supported, failing the response.", contentType)
			http.Error(rw, marshalError("UNSUPPORTED_MEDIA_TYPE", "Manifest.list is not supported."), http.StatusUnsupportedMediaType)
			return
		}
	}
	mh.next.ServeHTTP(rw, req)
}

type listReposHandler struct {
	next http.Handler
}

func (lrh listReposHandler) ServeHTTP(rw http.ResponseWriter, req *http.Request) {
	listReposFlag := MatchListRepos(req)
	if listReposFlag {
		rec = httptest.NewRecorder()
		lrh.next.ServeHTTP(rec, req)
		if rec.Result().StatusCode != http.StatusOK {
			copyResp(rec, rw)
			return
		}
		var ctlg struct {
			Repositories []string `json:"repositories"`
		}
		decoder := json.NewDecoder(rec.Body)
		if err := decoder.Decode(&ctlg); err != nil {
			log.Errorf("Decode repositories error: %v", err)
			copyResp(rec, rw)
			return
		}
		var entries []string
		for repo := range ctlg.Repositories {
			log.Debugf("the repo in the response %s", ctlg.Repositories[repo])
			exist := dao.RepositoryExists(ctlg.Repositories[repo])
			if exist {
				entries = append(entries, ctlg.Repositories[repo])
			}
		}
		type Repos struct {
			Repositories []string `json:"repositories"`
		}
		resp := &Repos{Repositories: entries}
		respJSON, err := json.Marshal(resp)
		if err != nil {
			log.Errorf("Encode repositories error: %v", err)
			copyResp(rec, rw)
			return
		}

		for k, v := range rec.Header() {
			rw.Header()[k] = v
		}
		clen := len(respJSON)
		rw.Header().Set(http.CanonicalHeaderKey("Content-Length"), strconv.Itoa(clen))
		rw.Write(respJSON)
		return
	}
	lrh.next.ServeHTTP(rw, req)
}

type contentTrustHandler struct {
	next http.Handler
}

func (cth contentTrustHandler) ServeHTTP(rw http.ResponseWriter, req *http.Request) {
	imgRaw := req.Context().Value(imageInfoCtxKey)
	if imgRaw == nil || !config.WithNotary() {
		cth.next.ServeHTTP(rw, req)
		return
	}
	img, _ := req.Context().Value(imageInfoCtxKey).(imageInfo)
	if img.digest == "" {
		cth.next.ServeHTTP(rw, req)
		return
	}
	if !getPolicyChecker().contentTrustEnabled(img.projectName) {
		cth.next.ServeHTTP(rw, req)
		return
	}
	match, err := matchNotaryDigest(img)
	if err != nil {
		http.Error(rw, marshalError("PROJECT_POLICY_VIOLATION", "Failed in communication with Notary please check the log"), http.StatusInternalServerError)
		return
	}
	if !match {
		log.Debugf("digest mismatch, failing the response.")
		http.Error(rw, marshalError("PROJECT_POLICY_VIOLATION", "The image is not signed in Notary."), http.StatusPreconditionFailed)
		return
	}
	cth.next.ServeHTTP(rw, req)
}

type vulnerableHandler struct {
	next http.Handler
}

func (vh vulnerableHandler) ServeHTTP(rw http.ResponseWriter, req *http.Request) {
	imgRaw := req.Context().Value(imageInfoCtxKey)
	if imgRaw == nil || !config.WithClair() {
		vh.next.ServeHTTP(rw, req)
		return
	}
	img, _ := req.Context().Value(imageInfoCtxKey).(imageInfo)
	if img.digest == "" {
		vh.next.ServeHTTP(rw, req)
		return
	}
	projectVulnerableEnabled, projectVulnerableSeverity := getPolicyChecker().vulnerablePolicy(img.projectName)
	if !projectVulnerableEnabled {
		vh.next.ServeHTTP(rw, req)
		return
	}
	// TODO: Get whitelist based on project setting
	wl, err := dao.GetSysCVEWhitelist()
	if err != nil {
		log.Errorf("Failed to get the whitelist, error: %v", err)
		http.Error(rw, marshalError("PROJECT_POLICY_VIOLATION", "Failed to get CVE whitelist."), http.StatusPreconditionFailed)
		return
	}
	vl, err := scan.VulnListByDigest(img.digest)
	if err != nil {
		log.Errorf("Failed to get the vulnerability list, error: %v", err)
		http.Error(rw, marshalError("PROJECT_POLICY_VIOLATION", "Failed to get vulnerabilities."), http.StatusPreconditionFailed)
		return
	}
	filtered := vl.ApplyWhitelist(*wl)
	msg := vh.filterMsg(img, filtered)
	log.Info(msg)
	if int(vl.Severity()) >= int(projectVulnerableSeverity) {
		log.Debugf("the image severity: %q is higher then project setting: %q, failing the response.", vl.Severity(), projectVulnerableSeverity)
		http.Error(rw, marshalError("PROJECT_POLICY_VIOLATION", fmt.Sprintf("The severity of vulnerability of the image: %q is equal or higher than the threshold in project setting: %q.", vl.Severity(), projectVulnerableSeverity)), http.StatusPreconditionFailed)
		return
	}
	vh.next.ServeHTTP(rw, req)
}

func (vh vulnerableHandler) filterMsg(img imageInfo, filtered scan.VulnerabilityList) string {
	filterMsg := fmt.Sprintf("Image: %s/%s:%s, digest: %s, vulnerabilities fitered by whitelist:", img.projectName, img.repository, img.reference, img.digest)
	if len(filtered) == 0 {
		filterMsg = fmt.Sprintf("%s none.", filterMsg)
	}
	for _, v := range filtered {
		filterMsg = fmt.Sprintf("%s ID: %s, severity: %s;", filterMsg, v.ID, v.Severity)
	}
	return filterMsg
}

func matchNotaryDigest(img imageInfo) (bool, error) {
	if NotaryEndpoint == "" {
		NotaryEndpoint = config.InternalNotaryEndpoint()
	}
	targets, err := notary.GetInternalTargets(NotaryEndpoint, tokenUsername, img.repository)
	if err != nil {
		return false, err
	}
	for _, t := range targets {
		if isDigest(img.reference) {
			d, err := notary.DigestFromTarget(t)
			if err != nil {
				return false, err
			}
			if img.digest == d {
				return true, nil
			}
		} else {
			if t.Tag == img.reference {
				log.Debugf("found reference: %s in notary, try to match digest.", img.reference)
				d, err := notary.DigestFromTarget(t)
				if err != nil {
					return false, err
				}
				if img.digest == d {
					return true, nil
				}
			}
		}
	}
	log.Debugf("image: %#v, not found in notary", img)
	return false, nil
}

// A sha256 is a string with 64 characters.
func isDigest(ref string) bool {
	return strings.HasPrefix(ref, "sha256:") && len(ref) == 71
}

func copyResp(rec *httptest.ResponseRecorder, rw http.ResponseWriter) {
	for k, v := range rec.Header() {
		rw.Header()[k] = v
	}
	rw.WriteHeader(rec.Result().StatusCode)
	rw.Write(rec.Body.Bytes())
}

func marshalError(code, msg string) string {
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

// JSONError wraps a concrete Code and Message, it's readable for docker deamon.
type JSONError struct {
	Code    string `json:"code,omitempty"`
	Message string `json:"message,omitempty"`
	Detail  string `json:"detail,omitempty"`
}
