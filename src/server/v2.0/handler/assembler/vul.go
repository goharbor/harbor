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

	"github.com/goharbor/harbor/src/api/scan"
	"github.com/goharbor/harbor/src/api/scanner"
	"github.com/goharbor/harbor/src/common/utils/log"
	"github.com/goharbor/harbor/src/internal"
	v1 "github.com/goharbor/harbor/src/pkg/scan/rest/v1"
	"github.com/goharbor/harbor/src/server/v2.0/handler/model"
)

const (
	vulnerabilitiesAddition = "vulnerabilities"
)

// NewVulAssembler returns vul assembler
func NewVulAssembler(withScanOverview bool) *VulAssembler {
	return &VulAssembler{
		withScanOverview: withScanOverview,
		scanCtl:          scan.DefaultController,
		scannerCtl:       scanner.DefaultController,
		scanners:         map[int64]bool{},
	}
}

// VulAssembler vul assembler
type VulAssembler struct {
	artifacts []*model.Artifact

	scanCtl    scan.Controller
	scannerCtl scanner.Controller
	scanners   map[int64]bool

	withScanOverview bool
}

func (assembler *VulAssembler) hasScanner(ctx context.Context, projectID int64) bool {
	value, ok := assembler.scanners[projectID]
	if !ok {
		scanner, err := assembler.scannerCtl.GetRegistrationByProject(projectID)
		if err != nil {
			log.Warningf("get scanner for project %d failed, error: %v", projectID, err)
			return false
		}

		value = scanner != nil
		assembler.scanners[projectID] = value
	}

	return value
}

// WithArtifacts set artifacts for the assembler
func (assembler *VulAssembler) WithArtifacts(artifacts ...*model.Artifact) *VulAssembler {
	assembler.artifacts = artifacts

	return assembler
}

// Assemble assemble vul for the artifacts
func (assembler *VulAssembler) Assemble(ctx context.Context) error {
	version := internal.GetAPIVersion(ctx)

	for _, artifact := range assembler.artifacts {
		hasScanner := assembler.hasScanner(ctx, artifact.ProjectID)

		if !hasScanner {
			continue
		}

		artifact.SetAdditionLink(vulnerabilitiesAddition, version)

		if assembler.withScanOverview {
			art := &v1.Artifact{
				NamespaceID: artifact.ProjectID,
				Repository:  artifact.RepositoryName,
				Digest:      artifact.Digest,
				MimeType:    artifact.ManifestMediaType,
			}

			overview, err := assembler.scanCtl.GetSummary(art, []string{v1.MimeTypeNativeReport})
			if err != nil {
				log.Warningf("get scan summary of artifact %s failed, error:%v", artifact.Digest, err)
			} else if len(overview) > 0 {
				artifact.ScanOverview = overview
			}
		}
	}

	return nil
}
