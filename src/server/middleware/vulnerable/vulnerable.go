package vulnerable

import (
	"github.com/goharbor/harbor/src/common/models"
	"github.com/goharbor/harbor/src/common/utils/log"
	internal_errors "github.com/goharbor/harbor/src/internal/error"
	sc "github.com/goharbor/harbor/src/pkg/scan/api/scan"
	"github.com/goharbor/harbor/src/pkg/scan/report"
	v1 "github.com/goharbor/harbor/src/pkg/scan/rest/v1"
	"github.com/goharbor/harbor/src/pkg/scan/vuln"
	serror "github.com/goharbor/harbor/src/server/error"
	"github.com/goharbor/harbor/src/server/middleware"
	"github.com/pkg/errors"
	"net/http"
	"net/http/httptest"
)

// Middleware handle docker pull vulnerable check
func Middleware() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
			doVulCheck, img, projectVulnerableSeverity, wl := validate(req)
			if !doVulCheck {
				next.ServeHTTP(rw, req)
				return
			}
			rec := httptest.NewRecorder()
			next.ServeHTTP(rec, req)
			// only enable vul policy check the response 200
			if rec.Result().StatusCode == http.StatusOK {
				// Invalid project ID
				if wl.ProjectID == 0 {
					err := errors.Errorf("project verification error: project %s", img.ProjectName)
					pkgE := internal_errors.New(err).WithCode(internal_errors.PROJECTPOLICYVIOLATION)
					serror.SendError(rw, pkgE)
					return
				}

				// Get the vulnerability summary
				artifact := &v1.Artifact{
					NamespaceID: wl.ProjectID,
					Repository:  img.Repository,
					Tag:         img.Tag,
					Digest:      img.Digest,
					MimeType:    v1.MimeTypeDockerArtifact,
				}

				cve := report.CVESet(wl.CVESet())
				summaries, err := sc.DefaultController.GetSummary(
					artifact,
					[]string{v1.MimeTypeNativeReport},
					report.WithCVEWhitelist(&cve),
				)

				if err != nil {
					err = errors.Wrap(err, "middleware: vulnerable handler")
					pkgE := internal_errors.New(err).WithCode(internal_errors.PROJECTPOLICYVIOLATION)
					serror.SendError(rw, pkgE)
					return
				}

				rawSummary, ok := summaries[v1.MimeTypeNativeReport]
				// No report yet?
				if !ok {
					err = errors.Errorf("no scan report existing for the artifact: %s:%s@%s", img.Repository, img.Tag, img.Digest)
					pkgE := internal_errors.New(err).WithCode(internal_errors.PROJECTPOLICYVIOLATION)
					serror.SendError(rw, pkgE)
					return
				}

				summary := rawSummary.(*vuln.NativeReportSummary)

				// Do judgement
				if summary.Severity.Code() >= projectVulnerableSeverity.Code() {
					err = errors.Errorf("current image with '%q vulnerable' cannot be pulled due to configured policy in 'Prevent images with vulnerability severity of %q from running.' "+
						"Please contact your project administrator for help'", summary.Severity, projectVulnerableSeverity)
					pkgE := internal_errors.New(err).WithCode(internal_errors.PROJECTPOLICYVIOLATION)
					serror.SendError(rw, pkgE)
					return
				}

				// Print scannerPull CVE list
				if len(summary.CVEBypassed) > 0 {
					for _, cve := range summary.CVEBypassed {
						log.Infof("Vulnerable policy check: scannerPull CVE %s", cve)
					}
				}
			}
			middleware.CopyResp(rec, rw)
		})
	}
}

func validate(req *http.Request) (bool, *middleware.ManifestInfo, vuln.Severity, models.CVEWhitelist) {
	var vs vuln.Severity
	var wl models.CVEWhitelist
	var mf *middleware.ManifestInfo
	mf, ok := middleware.ManifestInfoFromContext(req.Context())
	if !ok {
		return false, nil, vs, wl
	}

	exist, err := mf.ManifestExists(req.Context())
	if err != nil || !exist {
		return false, nil, vs, wl
	}

	if scannerPull, ok := middleware.ScannerPullFromContext(req.Context()); ok && scannerPull {
		return false, mf, vs, wl
	}
	// Is vulnerable policy set?
	projectVulnerableEnabled, projectVulnerableSeverity, wl := middleware.GetPolicyChecker().VulnerablePolicy(mf.ProjectName)
	if !projectVulnerableEnabled {
		return false, mf, vs, wl
	}
	return true, mf, projectVulnerableSeverity, wl
}
