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
	models "github.com/goharbor/harbor/src/pkg/scan/dao/scanner"
)

// Checker checker which can check that the artifact is scannable
type Checker interface {
	// IsScannable returns true when the artifact is scannable
	IsScannable(ctx context.Context, artifact *artifact.Artifact) (bool, error)
}

// NewChecker returns checker
func NewChecker() Checker {
	return &checker{
		artifactCtl:   artifact.Ctl,
		scannerCtl:    scanner.DefaultController,
		registrations: map[int64]*models.Registration{},
	}
}

type checker struct {
	artifactCtl   artifact.Controller
	scannerCtl    scanner.Controller
	registrations map[int64]*models.Registration
}

func (c *checker) IsScannable(ctx context.Context, art *artifact.Artifact) (bool, error) {
	// There are two scenarios when artifact is scannable:
	// 1. The scanner has capability for the artifact directly, eg the artifact is docker image.
	// 2. The artifact is image index and the scanner has capability for any artifact which is referenced by the artifact.

	projectID := art.ProjectID

	r, ok := c.registrations[projectID]
	if !ok {
		registration, err := c.scannerCtl.GetRegistrationByProject(projectID)
		if err != nil {
			return false, err
		}

		if registration == nil {
			return false, nil
		}

		r = registration
		c.registrations[projectID] = registration
	}

	var scannable bool

	walkFn := func(a *artifact.Artifact) error {
		if HasCapability(r, a) {
			scannable = true
			return artifact.ErrBreak
		}

		return nil
	}

	if err := c.artifactCtl.Walk(ctx, art, walkFn, nil); err != nil {
		return false, err
	}

	return scannable, nil
}

// HasCapability returns true when scanner has capability for the artifact
// See https://github.com/goharbor/pluggable-scanner-spec/issues/2 to get more info
func HasCapability(r *models.Registration, a *artifact.Artifact) bool {
	if a.Type == "CHART" || a.Type == "UNKNOWN" {
		return false
	}

	return r.HasCapability(a.ManifestMediaType)
}
