package manifestinfo

import (
	"fmt"
	"github.com/goharbor/harbor/src/common/utils"
	ierror "github.com/goharbor/harbor/src/internal/error"
	project2 "github.com/goharbor/harbor/src/pkg/project"
	"github.com/goharbor/harbor/src/server/middleware"
	reg_err "github.com/goharbor/harbor/src/server/registry/error"
	"github.com/opencontainers/go-digest"
	"net/http"
	"regexp"
	"strings"
)

var (
	manifestURLRe = regexp.MustCompile(`^/v2/((?:[a-z0-9]+(?:[._-][a-z0-9]+)*/)+)manifests/([\w][\w.:-]{0,127})`)
)

// Middleware gets the manifest information from request and inject it into the context
func Middleware() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
			mf, err := parseManifestInfoFromPath(req)
			if err != nil {
				reg_err.Handle(rw, req, err)
				return
			}
			*req = *(req.WithContext(middleware.NewManifestInfoContext(req.Context(), mf)))
			next.ServeHTTP(rw, req)
		})
	}
}

// parseManifestInfoFromPath parse manifest from request path
func parseManifestInfoFromPath(req *http.Request) (*middleware.ManifestInfo, error) {
	match, repository, reference := MatchManifestURL(req)
	if !match {
		return nil, fmt.Errorf("not match url %s for manifest", req.URL.Path)
	}

	projectName, _ := utils.ParseRepository(repository)
	project, err := project2.Mgr.Get(projectName)
	if err != nil {
		return nil, fmt.Errorf("failed to get project %s, error: %v", projectName, err)
	}
	if project == nil {
		return nil, ierror.NotFoundError(nil).WithMessage("project %s not found", projectName)
	}

	info := &middleware.ManifestInfo{
		ProjectID:  project.ProjectID,
		Repository: repository,
	}

	dgt, err := digest.Parse(reference)
	if err != nil {
		info.Tag = reference
	} else {
		info.Digest = dgt.String()
	}

	return info, nil
}

// MatchManifestURL ...
func MatchManifestURL(req *http.Request) (bool, string, string) {
	s := manifestURLRe.FindStringSubmatch(req.URL.Path)
	if len(s) == 3 {
		s[1] = strings.TrimSuffix(s[1], "/")
		return true, s[1], s[2]
	}
	return false, "", ""
}
