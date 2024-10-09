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

package sbom

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"
	"strings"
	"time"

	"github.com/goharbor/harbor/src/common/rbac"
	"github.com/goharbor/harbor/src/controller/artifact"
	scanCtl "github.com/goharbor/harbor/src/controller/scan"
	"github.com/goharbor/harbor/src/jobservice/job"
	"github.com/goharbor/harbor/src/jobservice/logger"
	"github.com/goharbor/harbor/src/lib/errors"
	"github.com/goharbor/harbor/src/lib/log"
	"github.com/goharbor/harbor/src/lib/orm"
	accessoryModel "github.com/goharbor/harbor/src/pkg/accessory/model"
	"github.com/goharbor/harbor/src/pkg/permission/types"
	"github.com/goharbor/harbor/src/pkg/robot/model"
	"github.com/goharbor/harbor/src/pkg/scan"
	scanModel "github.com/goharbor/harbor/src/pkg/scan/dao/scan"
	"github.com/goharbor/harbor/src/pkg/scan/dao/scanner"
	sbom "github.com/goharbor/harbor/src/pkg/scan/sbom/model"
	"github.com/goharbor/harbor/src/pkg/task"

	sc "github.com/goharbor/harbor/src/controller/scanner"

	v1 "github.com/goharbor/harbor/src/pkg/scan/rest/v1"
)

const (
	sbomMimeType      = "application/vnd.goharbor.harbor.sbom.v1"
	sbomMediaTypeSpdx = "application/spdx+json"
)

func init() {
	scan.RegisterScanHanlder(v1.ScanTypeSbom, &scanHandler{
		GenAccessoryFunc:       scan.GenAccessoryArt,
		SBOMMgrFunc:            func() Manager { return Mgr },
		TaskMgrFunc:            func() task.Manager { return task.Mgr },
		ArtifactControllerFunc: func() artifact.Controller { return artifact.Ctl },
		ScanControllerFunc:     func() scanCtl.Controller { return scanCtl.DefaultController },
		ScannerControllerFunc:  func() sc.Controller { return sc.DefaultController },
		cloneCtx:               orm.Clone,
	})
}

// scanHandler defines the Handler to generate sbom
type scanHandler struct {
	GenAccessoryFunc       func(scanRep v1.ScanRequest, sbomContent []byte, labels map[string]string, mediaType string, robot *model.Robot) (string, error)
	SBOMMgrFunc            func() Manager
	TaskMgrFunc            func() task.Manager
	ArtifactControllerFunc func() artifact.Controller
	ScanControllerFunc     func() scanCtl.Controller
	ScannerControllerFunc  func() sc.Controller
	cloneCtx               func(ctx context.Context) context.Context
}

// RequestProducesMineTypes defines the mine types produced by the scan handler
func (h *scanHandler) RequestProducesMineTypes() []string {
	return []string{v1.MimeTypeSBOMReport}
}

// RequestParameters defines the parameters for scan request
func (h *scanHandler) RequestParameters() map[string]interface{} {
	return map[string]interface{}{"sbom_media_types": []string{sbomMediaTypeSpdx}}
}

// PostScan defines task specific operations after the scan is complete
func (h *scanHandler) PostScan(ctx job.Context, sr *v1.ScanRequest, _ *scanModel.Report, rawReport string, startTime time.Time, robot *model.Robot) (string, error) {
	sbomContent, s, err := retrieveSBOMContent(rawReport)
	if err != nil {
		return "", err
	}
	scanReq := v1.ScanRequest{
		Registry: sr.Registry,
		Artifact: sr.Artifact,
	}
	scanReq.Registry.Insecure = strings.HasPrefix(scanReq.Registry.URL, "http://")
	// the registry URL should not contain http:// or https:// prefix
	scanReq.Registry.URL = strings.TrimPrefix(scanReq.Registry.URL, "http://")
	scanReq.Registry.URL = strings.TrimPrefix(scanReq.Registry.URL, "https://")
	if len(scanReq.Registry.URL) == 0 {
		return "", fmt.Errorf("empty registry server")
	}
	myLogger := ctx.GetLogger()
	myLogger.Debugf("Pushing accessory artifact to %s/%s", scanReq.Registry.URL, scanReq.Artifact.Repository)
	dgst, err := h.GenAccessoryFunc(scanReq, sbomContent, h.annotations(), sbomMimeType, robot)
	if err != nil {
		myLogger.Errorf("error when create accessory from image %v", err)
		return "", err
	}
	return h.generateReport(startTime, sr.Artifact.Repository, dgst, "Success", s)
}

// URLParameter defines the parameters for scan report url
func (h *scanHandler) URLParameter(_ *v1.ScanRequest) (string, error) {
	return fmt.Sprintf("sbom_media_type=%s", url.QueryEscape(sbomMediaTypeSpdx)), nil
}

// RequiredPermissions defines the permission used by the scan robot account
func (h *scanHandler) RequiredPermissions() []*types.Policy {
	return []*types.Policy{
		{
			Resource: rbac.ResourceRepository,
			Action:   rbac.ActionPull,
		},
		{
			Resource: rbac.ResourceRepository,
			Action:   rbac.ActionScannerPull,
		},
		{
			Resource: rbac.ResourceRepository,
			Action:   rbac.ActionPush,
		},
	}
}

// annotations defines the annotations for the accessory artifact
func (h *scanHandler) annotations() map[string]string {
	t := time.Now().Format(time.RFC3339)
	return map[string]string{
		"created":                             t,
		"created-by":                          "Harbor",
		"org.opencontainers.artifact.created": t,
		"org.opencontainers.artifact.description": "SPDX JSON SBOM",
	}
}

func (h *scanHandler) generateReport(startTime time.Time, repository, digest, status string, scanner *v1.Scanner) (string, error) {
	summary := sbom.Summary{}
	endTime := time.Now()
	summary[sbom.StartTime] = startTime
	summary[sbom.EndTime] = endTime
	summary[sbom.Duration] = int64(endTime.Sub(startTime).Seconds())
	summary[sbom.SBOMRepository] = repository
	summary[sbom.SBOMDigest] = digest
	summary[sbom.ScanStatus] = status
	summary[sbom.Scanner] = scanner
	rep, err := json.Marshal(summary)
	if err != nil {
		return "", err
	}
	return string(rep), nil
}

func (h *scanHandler) Update(ctx context.Context, uuid string, report string) error {
	mgr := h.SBOMMgrFunc()
	if err := mgr.UpdateReportData(ctx, uuid, report); err != nil {
		return err
	}
	return nil
}

// retrieveSBOMContent retrieves the "sbom" field from the raw report
func retrieveSBOMContent(rawReport string) ([]byte, *v1.Scanner, error) {
	rpt := sbom.RawSBOMReport{}
	err := json.Unmarshal([]byte(rawReport), &rpt)
	if err != nil {
		return nil, nil, err
	}
	sbomContent, err := json.Marshal(rpt.SBOM)
	if err != nil {
		return nil, nil, err
	}
	return sbomContent, rpt.Scanner, nil
}

func (h *scanHandler) MakePlaceHolder(ctx context.Context, art *artifact.Artifact, r *scanner.Registration) (rps []*scanModel.Report, err error) {
	mgr := h.SBOMMgrFunc()
	mimeTypes := r.GetProducesMimeTypes(art.ManifestMediaType, v1.ScanTypeSbom)
	if len(mimeTypes) == 0 {
		return nil, errors.New("no mime types to make report placeholders")
	}
	if err := h.delete(ctx, art, mimeTypes[0], r); err != nil {
		return nil, err
	}
	var reports []*scanModel.Report
	for _, mt := range mimeTypes {
		report := &sbom.Report{
			ArtifactID:       art.ID,
			RegistrationUUID: r.UUID,
			MimeType:         mt,
			MediaType:        sbomMediaTypeSpdx,
		}

		create := func(ctx context.Context) error {
			reportUUID, err := mgr.Create(ctx, report)
			if err != nil {
				return err
			}
			report.UUID = reportUUID
			return nil
		}

		if err := orm.WithTransaction(create)(orm.SetTransactionOpNameToContext(ctx, "tx-make-report-placeholder-sbom")); err != nil {
			return nil, err
		}
		reports = append(reports, &scanModel.Report{
			RegistrationUUID: r.UUID,
			MimeType:         mt,
			UUID:             report.UUID,
		})
	}

	return reports, nil
}

// delete deletes the sbom report and accessory
func (h *scanHandler) delete(ctx context.Context, art *artifact.Artifact, mimeTypes string, r *scanner.Registration) error {
	mgr := h.SBOMMgrFunc()
	sbomReports, err := mgr.GetBy(h.cloneCtx(ctx), art.ID, r.UUID, mimeTypes, sbomMediaTypeSpdx)
	if err != nil {
		return err
	}
	// check if any report has running task associate with it
	taskMgr := h.TaskMgrFunc()
	for _, rpt := range sbomReports {
		if !taskMgr.IsTaskFinished(ctx, rpt.UUID) {
			return errors.ConflictError(nil).WithMessage("a previous sbom generate process is running")
		}
	}

	for _, rpt := range sbomReports {
		if rpt.MimeType != v1.MimeTypeSBOMReport {
			continue
		}
		if err := mgr.Delete(ctx, rpt.UUID); err != nil {
			return err
		}
	}
	if err := h.deleteSBOMAccessory(ctx, art.ID); err != nil {
		return err
	}
	return nil
}

// deleteSBOMAccessory check if current report has sbom accessory info, if there is, delete it
func (h *scanHandler) deleteSBOMAccessory(ctx context.Context, artID int64) error {
	artifactCtl := h.ArtifactControllerFunc()
	art, err := artifactCtl.Get(ctx, artID, &artifact.Option{
		WithAccessory: true,
	})
	if err != nil {
		return err
	}
	if art == nil {
		return nil
	}
	for _, acc := range art.Accessories {
		if acc.GetData().Type == accessoryModel.TypeHarborSBOM {
			if err := artifactCtl.Delete(ctx, acc.GetData().ArtifactID); err != nil {
				return err
			}
		}
	}
	return nil
}

func (h *scanHandler) GetPlaceHolder(ctx context.Context, artRepo string, artDigest, scannerUUID string, mimeType string) (rp *scanModel.Report, err error) {
	artifactCtl := h.ArtifactControllerFunc()
	a, err := artifactCtl.GetByReference(ctx, artRepo, artDigest, nil)
	if err != nil {
		return nil, err
	}
	mgr := h.SBOMMgrFunc()
	rpts, err := mgr.GetBy(ctx, a.ID, scannerUUID, mimeType, sbomMediaTypeSpdx)
	if err != nil {
		logger.Errorf("Failed to get report for artifact %s@%s of mimetype %s, error %v", artRepo, artDigest, mimeType, err)
		return nil, err
	}
	if len(rpts) == 0 {
		logger.Errorf("No report found for artifact %s@%s of mimetype %s, error %v", artRepo, artDigest, mimeType, err)
		return nil, errors.NotFoundError(nil).WithMessage("no report found to update data")
	}
	return &scanModel.Report{
		UUID:     rpts[0].UUID,
		MimeType: rpts[0].MimeType,
	}, nil
}

func (h *scanHandler) GetSummary(ctx context.Context, art *artifact.Artifact, mimeTypes []string) (map[string]interface{}, error) {
	if len(mimeTypes) == 0 {
		return nil, errors.New("no mime types to get report summaries")
	}
	if art == nil {
		return nil, errors.New("no way to get report summaries for nil artifact")
	}
	ds := h.ScannerControllerFunc()
	r, err := ds.GetRegistrationByProject(ctx, art.ProjectID)
	if err != nil {
		return nil, errors.Wrap(err, "get sbom summary failed")
	}
	reports, err := h.SBOMMgrFunc().GetBy(ctx, art.ID, r.UUID, mimeTypes[0], sbomMediaTypeSpdx)
	if err != nil {
		return nil, err
	}
	if len(reports) == 0 {
		return map[string]interface{}{}, nil
	}
	reportContent := reports[0].ReportSummary
	result := map[string]interface{}{}
	if len(reportContent) == 0 {
		status := h.TaskMgrFunc().RetrieveStatusFromTask(ctx, reports[0].UUID)
		if len(status) > 0 {
			result[sbom.ReportID] = reports[0].UUID
			result[sbom.ScanStatus] = status
		}
		log.Debug("no content for current report")
		return result, nil
	}
	err = json.Unmarshal([]byte(reportContent), &result)
	return result, err
}

func (h *scanHandler) JobVendorType() string {
	return job.SBOMJobVendorType
}
