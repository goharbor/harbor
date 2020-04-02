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
	"net/http"

	"github.com/goharbor/harbor/src/common/rbac"
	"github.com/goharbor/harbor/src/common/security"
	"github.com/goharbor/harbor/src/controller/artifact"
	"github.com/goharbor/harbor/src/controller/project"
	"github.com/goharbor/harbor/src/controller/scan"
	"github.com/goharbor/harbor/src/lib"
	"github.com/goharbor/harbor/src/lib/errors"
	"github.com/goharbor/harbor/src/lib/log"
	"github.com/goharbor/harbor/src/pkg/scan/report"
	v1 "github.com/goharbor/harbor/src/pkg/scan/rest/v1"
	"github.com/goharbor/harbor/src/pkg/scan/vuln"
	"github.com/goharbor/harbor/src/server/middleware"
)

var (
	scanChecker = func() scan.Checker {
		return scan.NewChecker()
	}
)

// Middleware middleware which does the vulnerability prevention checking for the artifact in GET /v2/<name>/manifests/<reference> API
func Middleware() func(http.Handler) http.Handler {
	return middleware.BeforeRequest(func(r *http.Request) error {
		ctx := r.Context()

		logger := log.G(ctx).WithFields(log.Fields{"middleware": "vulnerable"})

		none := lib.ArtifactInfo{}
		info := lib.GetArtifactInfo(ctx)
		if info == none {
			return fmt.Errorf("artifactinfo middleware required before this middleware")
		}

		art, err := artifactController.GetByReference(ctx, info.Repository, info.Reference, nil)
		if err != nil {
			if !errors.IsNotFoundErr(err) {
				logger.Errorf("get artifact failed, error %v", err)
			}
			return err
		}

		proj, err := projectController.Get(ctx, art.ProjectID, project.CVEWhitelist(true))
		if err != nil {
			logger.Errorf("get the project %d failed, error: %v", art.ProjectID, err)
			return err
		}

		if !proj.VulPrevented() {
			// vulnerability prevention disabled, skip the checking
			logger.Debugf("project %s vulnerability prevention disabled, skip the checking", proj.Name)
			return nil
		}

		securityCtx, ok := security.FromContext(ctx)
		if ok &&
			securityCtx.Name() == "robot" &&
			securityCtx.Can(rbac.ActionScannerPull, rbac.NewProjectNamespace(proj.ProjectID).Resource(rbac.ResourceRepository)) {
			// the artifact is pulling by the scanner, skip the checking
			logger.Debugf("artifact %s@%s is pulling by the scanner, skip the checking", art.RepositoryName, art.Digest)
			return nil
		}

		checker := scanChecker()

		scannable, err := checker.IsScannable(ctx, art)
		if err != nil {
			logger.Errorf("check the scannable status of the artifact %s@%s failed, error: %v", art.RepositoryName, art.Digest, err)
			return err
		}

		if !scannable {
			// the artifact is not scannable, skip the checking
			logger.Debugf("artifact %s@%s is not scannable, skip the checking", art.RepositoryName, art.Digest)
			return nil
		}

		whitelist := report.CVESet(proj.CVEWhitelist.CVESet())
		summaries, err := scanController.GetSummary(ctx, art, []string{v1.MimeTypeNativeReport}, report.WithCVEWhitelist(&whitelist))
		if err != nil {
			logger.Errorf("get vulnerability summary of the artifact %s@%s failed, error: %v", art.RepositoryName, art.Digest, err)
			return err
		}

		rawSummary, ok := summaries[v1.MimeTypeNativeReport]
		if !ok {
			// No report yet?
			msg := "vulnerability prevention enabled, but no scan report existing for the artifact"
			return errors.New(nil).WithCode(errors.PROJECTPOLICYVIOLATION).WithMessage(msg)
		}

		summary, ok := rawSummary.(*vuln.NativeReportSummary)
		if !ok {
			return fmt.Errorf("report summary is invalid")
		}

		if art.IsImageIndex() {
			// artifact is image index, skip the checking when it is in the whitelist
			skippingWhitelist := []string{artifact.ImageType, artifact.CNABType}
			for _, t := range skippingWhitelist {
				if art.Type == t {
					logger.Debugf("artifact %s@%s is image index and its type is %s in skipping whitelist, "+
						"skip the vulnerability prevention checking", art.RepositoryName, art.Digest, art.Type)
					return nil
				}
			}
		}

		// Do judgement
		severity := vuln.ParseSeverityVersion3(proj.Severity())
		if summary.Severity.Code() >= severity.Code() {
			msg := fmt.Sprintf("current image with '%q vulnerable' cannot be pulled due to configured policy in 'Prevent images with vulnerability severity of %q from running.' "+
				"Please contact your project administrator for help'", summary.Severity, severity)
			return errors.New(nil).WithCode(errors.PROJECTPOLICYVIOLATION).WithMessage(msg)
		}

		// Print scannerPull CVE list
		if len(summary.CVEBypassed) > 0 {
			for _, cve := range summary.CVEBypassed {
				logger.Infof("Vulnerable policy check: scannerPull CVE %s", cve)
			}
		}

		return nil
	})
}
