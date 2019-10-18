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
	"net/http"

	"github.com/goharbor/harbor/src/common/utils/log"
	"github.com/goharbor/harbor/src/core/middlewares/util"
	sc "github.com/goharbor/harbor/src/pkg/scan/api/scan"
	"github.com/goharbor/harbor/src/pkg/scan/report"
	v1 "github.com/goharbor/harbor/src/pkg/scan/rest/v1"
	"github.com/goharbor/harbor/src/pkg/scan/vuln"
	"github.com/pkg/errors"
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
	if imgRaw == nil {
		vh.next.ServeHTTP(rw, req)
		return
	}

	// Expected artifact specified?
	img, ok := imgRaw.(util.ImageInfo)
	if !ok || len(img.Digest) == 0 {
		vh.next.ServeHTTP(rw, req)
		return
	}

	// Token bypass policy check
	policyCheck := req.Context().Value(util.PolicyCheckCtxKey)
	if policyCheck != nil && !policyCheck.(bool) {
		vh.next.ServeHTTP(rw, req)
		return
	}

	// Is vulnerable policy set?
	projectVulnerableEnabled, projectVulnerableSeverity, wl := util.GetPolicyChecker().VulnerablePolicy(img.ProjectName)
	if !projectVulnerableEnabled {
		vh.next.ServeHTTP(rw, req)
		return
	}

	// Invalid project ID
	if wl.ProjectID == 0 {
		err := errors.Errorf("project verification error: project %s", img.ProjectName)
		vh.sendError(err, rw)

		return
	}

	// Get the vulnerability summary
	artifact := &v1.Artifact{
		NamespaceID: wl.ProjectID,
		Repository:  img.Repository,
		Tag:         img.Reference,
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
		vh.sendError(err, rw)

		return
	}

	rawSummary, ok := summaries[v1.MimeTypeNativeReport]
	// No report yet?
	if !ok {
		err = errors.Errorf("no scan report existing for the artifact: %s:%s@%s", img.Repository, img.Reference, img.Digest)
		vh.sendError(err, rw)

		return
	}

	summary := rawSummary.(*vuln.NativeReportSummary)

	// Do judgement
	if summary.Severity.Code() >= projectVulnerableSeverity.Code() {
		err = errors.Errorf("the pulling image severity %q is higher than or equal with the project setting %q, reject the response.", summary.Severity, projectVulnerableSeverity)
		vh.sendError(err, rw)

		return
	}

	// Print bypass CVE list
	if len(summary.CVEBypassed) > 0 {
		for _, cve := range summary.CVEBypassed {
			log.Infof("Vulnerable policy check: bypass CVE %s", cve)
		}
	}

	vh.next.ServeHTTP(rw, req)
}

func (vh vulnerableHandler) sendError(err error, rw http.ResponseWriter) {
	log.Error(err)
	http.Error(rw, util.MarshalError("PROJECT_POLICY_VIOLATION", err.Error()), http.StatusPreconditionFailed)
}
