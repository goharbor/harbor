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
	v1 "github.com/goharbor/harbor/src/pkg/scan/rest/v1"
	sbomModel "github.com/goharbor/harbor/src/pkg/scan/sbom/model"
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
		mimeTypes:      mimeTypes,
	}
}

// ScanReportAssembler vul assembler
type ScanReportAssembler struct {
	scanChecker scan.Checker
	scanCtl     scan.Controller

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
				overview, err := assembler.scanCtl.GetSummary(ctx, &artifact.Artifact, []string{mimeType})
				if err != nil {
					log.Warningf("get scan summary of artifact %s@%s for %s failed, error:%v", artifact.RepositoryName, artifact.Digest, mimeType, err)
				} else if len(overview) > 0 {
					artifact.ScanOverview = overview
					break
				}
			}
		}
		if assembler.overviewOption.WithSBOM {
			overview, err := assembler.scanCtl.GetSummary(ctx, &artifact.Artifact, []string{v1.MimeTypeSBOMReport})
			if err != nil {
				log.Warningf("get scan summary of artifact %s@%s for %s failed, error:%v", artifact.RepositoryName, artifact.Digest, v1.MimeTypeSBOMReport, err)
			}
			if len(overview) > 0 {
				artifact.SBOMOverView = map[string]interface{}{
					sbomModel.StartTime:  overview[sbomModel.StartTime],
					sbomModel.EndTime:    overview[sbomModel.EndTime],
					sbomModel.ScanStatus: overview[sbomModel.ScanStatus],
					sbomModel.SBOMDigest: overview[sbomModel.SBOMDigest],
					sbomModel.Duration:   overview[sbomModel.Duration],
					sbomModel.ReportID:   overview[sbomModel.ReportID],
					sbomModel.Scanner:    overview[sbomModel.Scanner],
				}
			}
		}
	}
	return nil
}
