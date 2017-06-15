package proxy

import (
	//	"github.com/vmware/harbor/src/ui/api"
	"github.com/vmware/harbor/src/common/utils/log"
	"github.com/vmware/harbor/src/common/utils/notary"
	"github.com/vmware/harbor/src/ui/config"

	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"regexp"
	"strings"
)

type contextKey string

const (
	manifestURLPattern = `^/v2/((?:[a-z0-9]+(?:[._-][a-z0-9]+)*/)+)manifests/([\w][\w.:-]{0,127})`
	imageInfoCtxKey    = contextKey("ImageInfo")
	//TODO: temp solution, remove after vmware/harbor#2242 is resolved.
	tokenUsername = "admin"
)

// NotaryEndpoint , exported for testing.
var NotaryEndpoint = config.InternalNotaryEndpoint()

// EnvChecker is the instance of envPolicyChecker
var EnvChecker = envPolicyChecker{}

// MatchPullManifest checks if the request looks like a request to pull manifest.  If it is returns the image and tag/sha256 digest as 2nd and 3rd return values
func MatchPullManifest(req *http.Request) (bool, string, string) {
	//TODO: add user agent check.
	if req.Method != http.MethodGet {
		return false, "", ""
	}
	re := regexp.MustCompile(manifestURLPattern)
	s := re.FindStringSubmatch(req.URL.Path)
	if len(s) == 3 {
		s[1] = strings.TrimSuffix(s[1], "/")
		return true, s[1], s[2]
	}
	return false, "", ""
}

// policyChecker checks the policy of a project by project name, to determine if it's needed to check the image's status under this project.
type policyChecker interface {
	// contentTrustEnabled returns whether a project has enabled content trust.
	contentTrustEnabled(name string) bool
	// vulnerableEnabled  returns whether a project has enabled content trust.
	vulnerableEnabled(name string) bool
}

//For testing
type envPolicyChecker struct{}

func (ec envPolicyChecker) contentTrustEnabled(name string) bool {
	return os.Getenv("PROJECT_CONTENT_TRUST") == "1"
}
func (ec envPolicyChecker) vulnerableEnabled(name string) bool {
	// TODO: May need get more information in vulnerable policies.
	return os.Getenv("PROJECT_VULNERABBLE") == "1"
}

//TODO: integrate with PMS to get project policies
func getPolicyChecker() policyChecker {
	return EnvChecker
}

type imageInfo struct {
	repository  string
	tag         string
	projectName string
	//	digest      string
}

type urlHandler struct {
	next http.Handler
}

//TODO: wrap a ResponseWriter to get the status code?

func (uh urlHandler) ServeHTTP(rw http.ResponseWriter, req *http.Request) {
	log.Debugf("in url handler, path: %s", req.URL.Path)
	req.URL.Path = strings.TrimPrefix(req.URL.Path, RegistryProxyPrefix)
	flag, repository, tag := MatchPullManifest(req)
	if flag {
		components := strings.SplitN(repository, "/", 2)
		if len(components) < 2 {
			http.Error(rw, fmt.Sprintf("Bad repository name: %s", repository), http.StatusBadRequest)
			return
		}
		/*
			//Need to get digest of the image.
			endpoint, err := config.RegistryURL()
			if err != nil {
				log.Errorf("Error getting Registry URL: %v", err)
				http.Error(rw, fmt.Sprintf("Failed due to internal Error: %v", err), http.StatusInternalError)
				return
			}
			rc, err := api.NewRepositoryClient(endpoint, false, username, repository, "repository", repository, "pull")
			if err != nil {
				log.Errorf("Error creating repository Client: %v", err)
				http.Error(rw, fmt.Sprintf("Failed due to internal Error: %v", err), http.StatusInternalError)
				return
			}
			digest, exist, err := rc.ManifestExist(tag)
			if err != nil {
				log.Errorf("Failed to get digest for tag: %s, error: %v", tag, err)
				http.Error(rw, fmt.Sprintf("Failed due to internal Error: %v", err), http.StatusInternalError)
				return
			}
		*/

		img := imageInfo{
			repository:  repository,
			tag:         tag,
			projectName: components[0],
		}
		log.Debugf("image info of the request: %#v", img)

		ctx := context.WithValue(req.Context(), imageInfoCtxKey, img)
		req = req.WithContext(ctx)

	}
	uh.next.ServeHTTP(rw, req)
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
	if !getPolicyChecker().contentTrustEnabled(img.projectName) {
		cth.next.ServeHTTP(rw, req)
		return
	}
	//May need to update status code, let's use recorder
	rec := httptest.NewRecorder()
	cth.next.ServeHTTP(rec, req)
	if rec.Result().StatusCode != http.StatusOK {
		copyResp(rec, rw)
		return
	}
	log.Debugf("showing digest")
	digest := rec.Header().Get(http.CanonicalHeaderKey("Docker-Content-Digest"))
	log.Debugf("digest: %s", digest)
	match, err := matchNotaryDigest(img, digest)
	if err != nil {
		http.Error(rw, "Failed in communication with Notary please check the log", http.StatusInternalServerError)
		return
	}
	if match {
		log.Debugf("Passing the response to outter responseWriter")
		copyResp(rec, rw)
	} else {
		log.Debugf("digest mismatch, failing the response.")
		http.Error(rw, "The image is not signed in Notary.", http.StatusPreconditionFailed)
	}
}

func matchNotaryDigest(img imageInfo, digest string) (bool, error) {
	targets, err := notary.GetInternalTargets(NotaryEndpoint, tokenUsername, img.repository)
	if err != nil {
		return false, err
	}
	for _, t := range targets {
		if t.Tag == img.tag {
			log.Debugf("found tag: %s in notary, try to match digest.")
			d, err := notary.DigestFromTarget(t)
			if err != nil {
				return false, err
			}
			return digest == d, nil
		}
	}
	log.Debugf("image: %#v, not found in notary", img)
	return false, nil
}

func copyResp(rec *httptest.ResponseRecorder, rw http.ResponseWriter) {
	for k, v := range rec.Header() {
		rw.Header()[k] = v
	}
	rw.WriteHeader(rec.Result().StatusCode)
	rw.Write(rec.Body.Bytes())
}
