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

package security

import (
	"fmt"
	"net/http"
	"regexp"
	"strings"

	"github.com/goharbor/harbor/src/common"
	"github.com/goharbor/harbor/src/common/api"
	"github.com/goharbor/harbor/src/common/security"
	"github.com/goharbor/harbor/src/common/security/local"
	"github.com/goharbor/harbor/src/controller/user"
	"github.com/goharbor/harbor/src/lib"
	"github.com/goharbor/harbor/src/lib/log"
	"github.com/goharbor/harbor/src/pkg/oidc"
)

var (
	base           = fmt.Sprintf("/api/%s", api.APIVersion)
	sysInfoAPI     = base + "/systeminfo"
	apiVersionAPI  = "/api/version"
	labelsAPI      = base + "/labels"
	projectsAPI    = base + "/projects"
	reposAPIRe     = regexp.MustCompile(fmt.Sprintf(`^%s/projects/.*/repositories$`, regexp.QuoteMeta(base)))
	artifactsAPIRe = regexp.MustCompile(fmt.Sprintf(`^%s/projects/.*/repositories/.*/artifacts$`, regexp.QuoteMeta(base)))
	tagsAPIRe      = regexp.MustCompile(fmt.Sprintf(`^%s/projects/.*/repositories/.*/artifacts/.*/tags/.*$`, regexp.QuoteMeta(base)))
	uctl           = user.Ctl
)

type oidcCli struct{}

func (o *oidcCli) Generate(req *http.Request) security.Context {
	ctx := req.Context()
	if lib.GetAuthMode(ctx) != common.OIDCAuth {
		return nil
	}
	logger := log.G(ctx)
	username, secret, ok := req.BasicAuth()
	if !ok {
		return nil
	}
	if !o.valid(req) {
		return nil
	}
	info, err := oidc.VerifySecret(ctx, username, secret)
	if err != nil {
		logger.Errorf("failed to verify secret, username: %s, error: %v", username, err)
		return nil
	}
	u, err := uctl.GetByName(ctx, username)
	if err != nil {
		logger.Errorf("failed to get user model, username: %s, error: %v", username, err)
		return nil
	}
	oidc.InjectGroupsToUser(info, u)
	logger.Debugf("an OIDC CLI security context generated for request %s %s", req.Method, req.URL.Path)
	return local.NewSecurityContext(u)
}

func (o *oidcCli) valid(req *http.Request) bool {

	path := strings.TrimSuffix(req.URL.Path, "/")

	if path == "/service/token" ||
		strings.HasPrefix(path, "/v2") ||
		strings.HasPrefix(path, "/chartrepo") ||
		strings.HasPrefix(path, "/api/chartrepo") {
		// The request was sent by CLI to upload/download artifacts
		return true
	}

	// additional requests for replication
	if req.Method == http.MethodPost && path == projectsAPI { // creating project
		return true
	}

	if req.Method == http.MethodGet && (path == projectsAPI || // list projects
		path == sysInfoAPI || // query sys info
		path == apiVersionAPI || // api version
		path == labelsAPI || // list labels
		reposAPIRe.MatchString(path) || // list repos
		artifactsAPIRe.MatchString(path)) { // list artifacts
		return true
	}

	if req.Method == http.MethodDelete && tagsAPIRe.MatchString(path) { // deleting tags
		return true
	}
	return false
}
