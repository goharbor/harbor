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

package assembler

import (
	"context"

	"github.com/goharbor/harbor/src/controller/scan"
	"github.com/goharbor/harbor/src/lib"
	"github.com/goharbor/harbor/src/lib/log"
	"github.com/goharbor/harbor/src/lib/q"
	v1 "github.com/goharbor/harbor/src/pkg/scan/rest/v1"
	sbomModel "github.com/goharbor/harbor/src/pkg/scan/sbom/model"
	"github.com/goharbor/harbor/src/pkg/task"
	"github.com/goharbor/harbor/src/server/v2.0/handler/model"
)

const (
	vulnerabilitiesAddition = "vulnerabilities"
)

// NewScanReportAssembler returns vul assembler
func NewScanReportAssembler(option *model.OverviewOptions, mimeTypes []string) *ScanReportAssembler {
	return &ScanReportAssembler{
		overviewOption: option,
		scanChecker:    scan.NewChecker(),
		scanCtl:        scan.DefaultController,
		executionMgr:   task.ExecMgr,
		mimeTypes:      mimeTypes,
	}
}

// ScanReportAssembler vul assembler
type ScanReportAssembler struct {
	scanChecker  scan.Checker
	scanCtl      scan.Controller
	executionMgr task.ExecutionManager

	artifacts      []*model.Artifact
	mimeTypes      []string
	overviewOption *model.OverviewOptions
}

// WithArtifacts set artifacts for the assembler
func (assembler *ScanReportAssembler) WithArtifacts(artifacts ...*model.Artifact) *ScanReportAssembler {
	assembler.artifacts = artifacts

	return assembler
}

// Assemble assemble vul for the artifacts
func (assembler *ScanReportAssembler) Assemble(ctx context.Context) error {
	version := lib.GetAPIVersion(ctx)

	for _, artifact := range assembler.artifacts {
		isScannable, err := assembler.scanChecker.IsScannable(ctx, &artifact.Artifact)
		if err != nil {
			log.Errorf("check the scannable status of %s@%s failed, error: %v", artifact.RepositoryName, artifact.Digest, err)
			continue
		}

		if !isScannable {
			continue
		}

		artifact.SetAdditionLink(vulnerabilitiesAddition, version)

		if assembler.overviewOption.WithVuln {
			for _, mimeType := range assembler.mimeTypes {
				overview, err := assembler.scanCtl.GetSummary(ctx, &artifact.Artifact, v1.ScanTypeVulnerability, []string{mimeType})
				if err != nil {
					log.Warningf("get scan summary of artifact %s@%s for %s failed, error:%v", artifact.RepositoryName, artifact.Digest, mimeType, err)
				} else if len(overview) > 0 {
					artifact.ScanOverview = overview
					break
				}
			}
		}

		// set sbom additional link if it is supported, use the empty digest
		artifact.SetSBOMAdditionLink("", version)
		if assembler.overviewOption.WithSBOM {
			overview, err := assembler.scanCtl.GetSummary(ctx, &artifact.Artifact, v1.ScanTypeSbom, []string{v1.MimeTypeSBOMReport})
			if err != nil {
				log.Warningf("get scan summary of artifact %s@%s for %s failed, error:%v", artifact.RepositoryName, artifact.Digest, v1.MimeTypeSBOMReport, err)
			}
			if len(overview) == 0 {
				// only fetch the sbom overview from execution when the overview is empty and the artifact has child references ( image index, cnab etc)
				if len(artifact.References) == 0 {
					continue
				}
				log.Warningf("overview is empty, retrieve sbom status from execution")
				// Get latest execution with digest, repository, and scan type is sbom, the status is the scan status
				query := q.New(
					q.KeyWords{"extra_attrs.artifact.digest": artifact.Digest,
						"extra_attrs.artifact.repository_name":  artifact.RepositoryName,
						"extra_attrs.enabled_capabilities.type": "sbom"})
				// sort by ID desc to get the latest execution
				query.Sorts = []*q.Sort{q.NewSort("ID", true)}
				execs, err := assembler.executionMgr.List(ctx, query)
				if err != nil {
					log.Warningf("get scan summary of artifact %s@%s for %s failed, error:%v", artifact.RepositoryName, artifact.Digest, v1.MimeTypeSBOMReport, err)
					continue
				}
				// if no execs, means this artifact is not scanned, leave the sbom_overview empty
				if len(execs) == 0 {
					continue
				}
				artifact.SBOMOverView = map[string]interface{}{
					sbomModel.ScanStatus: execs[0].Status,
					sbomModel.StartTime:  execs[0].StartTime,
					sbomModel.EndTime:    execs[0].EndTime,
					sbomModel.Duration:   int64(execs[0].EndTime.Sub(execs[0].StartTime).Seconds()),
				}
				continue
			}

			artifact.SBOMOverView = map[string]interface{}{
				sbomModel.StartTime:  overview[sbomModel.StartTime],
				sbomModel.EndTime:    overview[sbomModel.EndTime],
				sbomModel.ScanStatus: overview[sbomModel.ScanStatus],
				sbomModel.SBOMDigest: overview[sbomModel.SBOMDigest],
				sbomModel.Duration:   overview[sbomModel.Duration],
				sbomModel.ReportID:   overview[sbomModel.ReportID],
				sbomModel.Scanner:    overview[sbomModel.Scanner],
			}
			if sbomDgst, ok := overview[sbomModel.SBOMDigest].(string); ok {
				// set additional link for sbom digest
				artifact.SetSBOMAdditionLink(sbomDgst, version)
			}
		}
	}
	return nil
}
