package middleware

import (
	"context"
	"github.com/goharbor/harbor/src/common/models"
	"github.com/goharbor/harbor/src/common/utils/log"
	"github.com/goharbor/harbor/src/core/config"
	"github.com/goharbor/harbor/src/core/promgr"
	"github.com/goharbor/harbor/src/pkg/artifact"
	"github.com/goharbor/harbor/src/pkg/q"
	"github.com/goharbor/harbor/src/pkg/repository"
	"github.com/goharbor/harbor/src/pkg/scan/vuln"
	"github.com/goharbor/harbor/src/pkg/scan/whitelist"
	"github.com/goharbor/harbor/src/pkg/tag"
	"github.com/pkg/errors"
	"net/http"
	"net/http/httptest"
	"sync"
)

type contextKey string

const (
	// manifestInfoKey the context key for manifest info
	manifestInfoKey = contextKey("ManifestInfo")
	// ScannerPullCtxKey the context key for robot account to bypass the pull policy check.
	ScannerPullCtxKey = contextKey("ScannerPullCheck")
)

// ManifestInfo ...
type ManifestInfo struct {
	ProjectID   int64
	ProjectName string
	Repository  string
	Tag         string
	Digest      string

	manifestExist     bool
	manifestExistErr  error
	manifestExistOnce sync.Once
}

// ManifestExists ...
func (info *ManifestInfo) ManifestExists(ctx context.Context) (bool, error) {
	info.manifestExistOnce.Do(func() {

		// ToDo: use the artifact controller method
		total, repos, err := repository.Mgr.List(ctx, &q.Query{
			Keywords: map[string]interface{}{
				"Name": info.Repository,
			},
		})
		if err != nil {
			info.manifestExistErr = err
			return
		}
		if total == 0 {
			return
		}

		total, tags, err := tag.Mgr.List(ctx, &q.Query{
			Keywords: map[string]interface{}{
				"Name":         info.Tag,
				"RepositoryID": repos[0].RepositoryID,
			},
		})
		if err != nil {
			info.manifestExistErr = err
			return
		}
		if total == 0 {
			return
		}

		total, afs, err := artifact.Mgr.List(ctx, &q.Query{
			Keywords: map[string]interface{}{
				"ID": tags[0].ArtifactID,
			},
		})
		if err != nil {
			info.manifestExistErr = err
			return
		}
		if total == 0 {
			return
		}

		info.Digest = afs[0].Digest
		info.manifestExist = total > 0
	})

	return info.manifestExist, info.manifestExistErr
}

// NewManifestInfoContext returns context with manifest info
func NewManifestInfoContext(ctx context.Context, info *ManifestInfo) context.Context {
	return context.WithValue(ctx, manifestInfoKey, info)
}

// ManifestInfoFromContext returns manifest info from context
func ManifestInfoFromContext(ctx context.Context) (*ManifestInfo, bool) {
	info, ok := ctx.Value(manifestInfoKey).(*ManifestInfo)
	return info, ok
}

// NewScannerPullContext returns context with policy check info
func NewScannerPullContext(ctx context.Context, scannerPull bool) context.Context {
	return context.WithValue(ctx, ScannerPullCtxKey, scannerPull)
}

// ScannerPullFromContext returns whether to bypass policy check
func ScannerPullFromContext(ctx context.Context) (bool, bool) {
	info, ok := ctx.Value(ScannerPullCtxKey).(bool)
	return info, ok
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
	VulnerablePolicy(name string) (bool, vuln.Severity, models.CVEWhitelist)
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
	if project == nil {
		log.Debugf("project %s not found", name)
		return false
	}
	return project.ContentTrustEnabled()
}

// VulnerablePolicy ...
func (pc PmsPolicyChecker) VulnerablePolicy(name string) (bool, vuln.Severity, models.CVEWhitelist) {
	project, err := pc.pm.Get(name)
	wl := models.CVEWhitelist{}
	if err != nil {
		log.Errorf("Unexpected error when getting the project, error: %v", err)
		return true, vuln.Unknown, wl
	}

	mgr := whitelist.NewDefaultManager()
	if project.ReuseSysCVEWhitelist() {
		w, err := mgr.GetSys()
		if err != nil {
			log.Error(errors.Wrap(err, "policy checker: vulnerable policy"))
		} else {
			wl = *w

			// Use the real project ID
			wl.ProjectID = project.ProjectID
		}
	} else {
		w, err := mgr.Get(project.ProjectID)
		if err != nil {
			log.Error(errors.Wrap(err, "policy checker: vulnerable policy"))
		} else {
			wl = *w
		}
	}

	return project.VulPrevented(), vuln.ParseSeverityVersion3(project.Severity()), wl
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
