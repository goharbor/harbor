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

package scan

import (
	"context"

	"github.com/goharbor/harbor/src/api/artifact"
	"github.com/goharbor/harbor/src/api/scanner"
	v1 "github.com/goharbor/harbor/src/pkg/scan/rest/v1"
)

// Checker checker which can check that the artifact is scannable
type Checker interface {
	// IsScannable returns true when the artifact is scannable
	IsScannable(ctx context.Context, artifact *artifact.Artifact) (bool, error)
}

// NewChecker returns checker
func NewChecker() Checker {
	return &checker{
		artifactCtl:      artifact.Ctl,
		scannerCtl:       scanner.DefaultController,
		scannerMetadatas: map[int64]*v1.ScannerAdapterMetadata{},
	}
}

type checker struct {
	artifactCtl      artifact.Controller
	scannerCtl       scanner.Controller
	scannerMetadatas map[int64]*v1.ScannerAdapterMetadata
}

func (c *checker) IsScannable(ctx context.Context, art *artifact.Artifact) (bool, error) {
	projectID := art.ProjectID

	metadata, ok := c.scannerMetadatas[projectID]
	if !ok {
		registration, err := c.scannerCtl.GetRegistrationByProject(projectID, scanner.WithPing(false))
		if err != nil {
			return false, err
		}

		if registration == nil {
			return false, nil
		}

		md, err := c.scannerCtl.Ping(registration)
		if err != nil {
			return false, err
		}

		metadata = md
		c.scannerMetadatas[projectID] = md
	}

	var scannable bool

	walkFn := func(a *artifact.Artifact) error {
		scannable = metadata.HasCapability(a.ManifestMediaType)
		if scannable {
			return artifact.ErrBreak
		}

		return nil
	}

	if err := c.artifactCtl.Walk(ctx, art, walkFn, nil); err != nil {
		return false, err
	}

	return scannable, nil
}
