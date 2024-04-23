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

	"github.com/goharbor/harbor/src/common"
	"github.com/goharbor/harbor/src/lib/config"
	scanModel "github.com/goharbor/harbor/src/pkg/scan/dao/scan"
	sbom "github.com/goharbor/harbor/src/pkg/scan/sbom/model"

	"github.com/goharbor/harbor/src/common/rbac"
	"github.com/goharbor/harbor/src/jobservice/job"
	"github.com/goharbor/harbor/src/pkg/permission/types"
	"github.com/goharbor/harbor/src/pkg/robot/model"
	"github.com/goharbor/harbor/src/pkg/scan"

	v1 "github.com/goharbor/harbor/src/pkg/scan/rest/v1"
	"github.com/goharbor/harbor/src/pkg/scan/vuln"
)

const (
	sbomMimeType      = "application/vnd.goharbor.harbor.sbom.v1"
	sbomMediaTypeSpdx = "application/spdx+json"
)

func init() {
	scan.RegisterScanHanlder(v1.ScanTypeSbom, &scanHandler{GenAccessoryFunc: scan.GenAccessoryArt, RegistryServer: registryFQDN})
}

// ScanHandler defines the Handler to generate sbom
type scanHandler struct {
	GenAccessoryFunc func(scanRep v1.ScanRequest, sbomContent []byte, labels map[string]string, mediaType string, robot *model.Robot) (string, error)
	RegistryServer   func(ctx context.Context) string
}

// RequestProducesMineTypes defines the mine types produced by the scan handler
func (v *scanHandler) RequestProducesMineTypes() []string {
	return []string{v1.MimeTypeSBOMReport}
}

// RequestParameters defines the parameters for scan request
func (v *scanHandler) RequestParameters() map[string]interface{} {
	return map[string]interface{}{"sbom_media_types": []string{sbomMediaTypeSpdx}}
}

// ReportURLParameter defines the parameters for scan report url
func (v *scanHandler) ReportURLParameter(_ *v1.ScanRequest) (string, error) {
	return fmt.Sprintf("sbom_media_type=%s", url.QueryEscape(sbomMediaTypeSpdx)), nil
}

// RequiredPermissions defines the permission used by the scan robot account
func (v *scanHandler) RequiredPermissions() []*types.Policy {
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

// PostScan defines task specific operations after the scan is complete
func (v *scanHandler) PostScan(ctx job.Context, sr *v1.ScanRequest, _ *scanModel.Report, rawReport string, startTime time.Time, robot *model.Robot) (string, error) {
	sbomContent, err := retrieveSBOMContent(rawReport)
	if err != nil {
		return "", err
	}
	scanReq := v1.ScanRequest{
		Registry: sr.Registry,
		Artifact: sr.Artifact,
	}
	// the registry server url is core by default, need to replace it with real registry server url
	scanReq.Registry.URL = v.RegistryServer(ctx.SystemContext())
	if len(scanReq.Registry.URL) == 0 {
		return "", fmt.Errorf("empty registry server")
	}
	myLogger := ctx.GetLogger()
	myLogger.Debugf("Pushing accessory artifact to %s/%s", scanReq.Registry.URL, scanReq.Artifact.Repository)
	dgst, err := v.GenAccessoryFunc(scanReq, sbomContent, v.annotations(), sbomMimeType, robot)
	if err != nil {
		myLogger.Errorf("error when create accessory from image %v", err)
		return "", err
	}
	return v.generateReport(startTime, sr.Artifact.Repository, dgst, "Success")
}

// annotations defines the annotations for the accessory artifact
func (v *scanHandler) annotations() map[string]string {
	t := time.Now().Format(time.RFC3339)
	return map[string]string{
		"created":                             t,
		"created-by":                          "Harbor",
		"org.opencontainers.artifact.created": t,
		"org.opencontainers.artifact.description": "SPDX JSON SBOM",
	}
}

func (v *scanHandler) generateReport(startTime time.Time, repository, digest, status string) (string, error) {
	summary := sbom.Summary{}
	endTime := time.Now()
	summary[sbom.StartTime] = startTime
	summary[sbom.EndTime] = endTime
	summary[sbom.Duration] = int64(endTime.Sub(startTime).Seconds())
	summary[sbom.SBOMRepository] = repository
	summary[sbom.SBOMDigest] = digest
	summary[sbom.ScanStatus] = status
	rep, err := json.Marshal(summary)
	if err != nil {
		return "", err
	}
	return string(rep), nil
}

// extract server name from config, and remove the protocol prefix
func registryFQDN(ctx context.Context) string {
	cfgMgr, ok := config.FromContext(ctx)
	if ok {
		extURL := cfgMgr.Get(context.Background(), common.ExtEndpoint).GetString()
		server := strings.TrimPrefix(extURL, "https://")
		server = strings.TrimPrefix(server, "http://")
		return server
	}
	return ""
}

// retrieveSBOMContent retrieves the "sbom" field from the raw report
func retrieveSBOMContent(rawReport string) ([]byte, error) {
	rpt := vuln.Report{}
	err := json.Unmarshal([]byte(rawReport), &rpt)
	if err != nil {
		return nil, err
	}
	sbomContent, err := json.Marshal(rpt.SBOM)
	if err != nil {
		return nil, err
	}
	return sbomContent, nil
}
