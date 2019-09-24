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

	"github.com/goharbor/harbor/src/pkg/scan/report"

	"github.com/goharbor/harbor/src/common/models"
	"github.com/goharbor/harbor/src/common/rbac"
	"github.com/goharbor/harbor/src/common/utils"
	coreutils "github.com/goharbor/harbor/src/core/utils"
	"github.com/goharbor/harbor/src/jobservice/logger"
	"github.com/goharbor/harbor/src/pkg/scan/api/scan"
	v1 "github.com/goharbor/harbor/src/pkg/scan/rest/v1"
	"github.com/pkg/errors"
)

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
	if !sa.SecurityCtx.IsAuthenticated() {
		sa.SendUnAuthorizedError(errors.New("Unauthorized"))
		return
	}

	// Assemble artifact object
	digest, err := getDigest(repoName, tag, sa.SecurityCtx.GetUsername())
	if err != nil {
		sa.SendInternalServerError(errors.Wrap(err, "scan API: prepare"))
		return
	}

	sa.artifact = &v1.Artifact{
		NamespaceID: pro.ProjectID,
		Repository:  repoName,
		Digest:      digest,
		MimeType:    v1.MimeTypeDockerArtifact,
	}

	logger.Debugf("scan artifact: %#v", sa.artifact)
}

// Scan artifact
func (sa *ScanAPI) Scan() {
	// Check access permissions
	resource := rbac.NewProjectNamespace(sa.pro.ProjectID).Resource(rbac.ResourceScan)
	if !sa.SecurityCtx.Can(rbac.ActionCreate, resource) {
		sa.SendForbiddenError(errors.New(sa.SecurityCtx.GetUsername()))
		return
	}

	if err := scan.DefaultController.Scan(sa.artifact); err != nil {
		sa.SendInternalServerError(errors.Wrap(err, "scan API: scan"))
		return
	}
}

// Report returns the required reports with the given mime types.
func (sa *ScanAPI) Report() {
	// Check access permissions
	resource := rbac.NewProjectNamespace(sa.pro.ProjectID).Resource(rbac.ResourceScan)
	if !sa.SecurityCtx.Can(rbac.ActionRead, resource) {
		sa.SendForbiddenError(errors.New(sa.SecurityCtx.GetUsername()))
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
		sa.SendInternalServerError(errors.Wrap(err, "scan API: get report"))
		return
	}

	vulItems := make(map[string]interface{}, len(reports))
	for _, rp := range reports {
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
	resource := rbac.NewProjectNamespace(sa.pro.ProjectID).Resource(rbac.ResourceScan)
	if !sa.SecurityCtx.Can(rbac.ActionRead, resource) {
		sa.SendForbiddenError(errors.New(sa.SecurityCtx.GetUsername()))
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
