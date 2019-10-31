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

package api

import (
	"net/http"
	"strconv"

	"github.com/goharbor/harbor/src/common/models"
	"github.com/goharbor/harbor/src/common/rbac"
	"github.com/goharbor/harbor/src/common/utils"
	coreutils "github.com/goharbor/harbor/src/core/utils"
	"github.com/goharbor/harbor/src/jobservice/logger"
	"github.com/goharbor/harbor/src/pkg/errs"
	"github.com/goharbor/harbor/src/pkg/scan/api/scan"
	"github.com/goharbor/harbor/src/pkg/scan/report"
	v1 "github.com/goharbor/harbor/src/pkg/scan/rest/v1"
	"github.com/pkg/errors"
)

var digestFunc digestGetter = getDigest

// ScanAPI handles the scan related actions
type ScanAPI struct {
	BaseController

	// Target artifact
	artifact *v1.Artifact
	// Project reference
	pro *models.Project
}

// Prepare sth. for the subsequent actions
func (sa *ScanAPI) Prepare() {
	// Call super prepare method
	sa.BaseController.Prepare()

	// Parse parameters
	repoName := sa.GetString(":splat")
	tag := sa.GetString(":tag")
	projectName, _ := utils.ParseRepository(repoName)

	pro, err := sa.ProjectMgr.Get(projectName)
	if err != nil {
		sa.SendInternalServerError(errors.Wrap(err, "scan API: prepare"))
		return
	}
	if pro == nil {
		sa.SendNotFoundError(errors.Errorf("project %s not found", projectName))
		return
	}
	sa.pro = pro

	// Check authentication
	if !sa.RequireAuthenticated() {
		return
	}

	// Assemble artifact object
	digest, err := digestFunc(repoName, tag, sa.SecurityCtx.GetUsername())
	if err != nil {
		sa.SendInternalServerError(errors.Wrap(err, "scan API: prepare"))
		return
	}

	sa.artifact = &v1.Artifact{
		NamespaceID: pro.ProjectID,
		Repository:  repoName,
		Tag:         tag,
		Digest:      digest,
		MimeType:    v1.MimeTypeDockerArtifact,
	}

	logger.Debugf("Scan API receives artifact: %#v", sa.artifact)
}

// Scan artifact
func (sa *ScanAPI) Scan() {
	// Check access permissions
	if !sa.RequireProjectAccess(sa.pro.ProjectID, rbac.ActionCreate, rbac.ResourceScan) {
		return
	}

	if err := scan.DefaultController.Scan(sa.artifact); err != nil {
		e := errors.Wrap(err, "scan API: scan")

		if errs.AsError(err, errs.PreconditionFailed) {
			sa.SendPreconditionFailedError(e)
			return
		}

		if errs.AsError(err, errs.Conflict) {
			sa.SendConflictError(e)
			return
		}

		sa.SendInternalServerError(e)
		return
	}

	sa.Ctx.ResponseWriter.WriteHeader(http.StatusAccepted)
}

// Report returns the required reports with the given mime types.
func (sa *ScanAPI) Report() {
	// Check access permissions
	if !sa.RequireProjectAccess(sa.pro.ProjectID, rbac.ActionRead, rbac.ResourceScan) {
		return
	}

	// Extract mime types
	producesMimes := make([]string, 0)
	if hl, ok := sa.Ctx.Request.Header[v1.HTTPAcceptHeader]; ok && len(hl) > 0 {
		producesMimes = append(producesMimes, hl...)
	}

	// Get the reports
	reports, err := scan.DefaultController.GetReport(sa.artifact, producesMimes)
	if err != nil {
		e := errors.Wrap(err, "scan API: get report")

		if errs.AsError(err, errs.PreconditionFailed) {
			sa.SendPreconditionFailedError(e)
			return
		}

		sa.SendInternalServerError(e)
		return
	}

	vulItems := make(map[string]interface{})
	for _, rp := range reports {
		// Resolve scan report data only when it is ready
		if len(rp.Report) == 0 {
			continue
		}

		vrp, err := report.ResolveData(rp.MimeType, []byte(rp.Report))
		if err != nil {
			sa.SendInternalServerError(errors.Wrap(err, "scan API: get report"))
			return
		}

		vulItems[rp.MimeType] = vrp
	}

	sa.Data["json"] = vulItems
	sa.ServeJSON()
}

// Log returns the log stream
func (sa *ScanAPI) Log() {
	// Check access permissions
	if !sa.RequireProjectAccess(sa.pro.ProjectID, rbac.ActionRead, rbac.ResourceScan) {
		return
	}

	uuid := sa.GetString(":uuid")
	bytes, err := scan.DefaultController.GetScanLog(uuid)
	if err != nil {
		sa.SendInternalServerError(errors.Wrap(err, "scan API: log"))
		return
	}

	if bytes == nil {
		// Not found
		sa.SendNotFoundError(errors.Errorf("report with uuid %s does not exist", uuid))
		return
	}

	sa.Ctx.ResponseWriter.Header().Set(http.CanonicalHeaderKey("Content-Length"), strconv.Itoa(len(bytes)))
	sa.Ctx.ResponseWriter.Header().Set(http.CanonicalHeaderKey("Content-Type"), "text/plain")
	_, err = sa.Ctx.ResponseWriter.Write(bytes)
	if err != nil {
		sa.SendInternalServerError(errors.Wrap(err, "scan API: log"))
	}
}

// digestGetter is a function template for getting digest.
// TODO: This can be removed if the registry access interface is ready.
type digestGetter func(repo, tag string, username string) (string, error)

func getDigest(repo, tag string, username string) (string, error) {
	client, err := coreutils.NewRepositoryClientForUI(username, repo)
	if err != nil {
		return "", err
	}

	digest, exists, err := client.ManifestExist(tag)
	if err != nil {
		return "", err
	}

	if !exists {
		return "", errors.Errorf("tag %s does exist", tag)
	}

	return digest, nil
}
