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
	"github.com/goharbor/harbor/src/controller/artifact"
	"github.com/goharbor/harbor/src/controller/artifact/processor/image"
	"github.com/goharbor/harbor/src/controller/scanner"
	"github.com/goharbor/harbor/src/lib/q"
	"github.com/goharbor/harbor/src/pkg/accessory"
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
		accMgr:        accessory.Mgr,
		scannerCtl:    scanner.DefaultController,
		registrations: map[int64]*models.Registration{},
	}
}

type checker struct {
	artifactCtl   artifact.Controller
	accMgr        accessory.Manager
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
		registration, err := c.scannerCtl.GetRegistrationByProject(ctx, projectID)
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
		ok, err := c.isAccessory(ctx, a)
		if err != nil {
			return err
		}
		if ok {
			return nil
		}

		if hasCapability(r, a) {
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

func (c *checker) isAccessory(ctx context.Context, art *artifact.Artifact) (bool, error) {
	ac, err := c.accMgr.List(ctx, q.New(q.KeyWords{"ArtifactID": art.Artifact.ID, "digest": art.Artifact.Digest}))
	if err != nil {
		return false, err
	}
	if len(ac) > 0 {
		return true, nil
	}
	return false, nil
}

// hasCapability returns true when scanner has capability for the artifact
// See https://github.com/goharbor/pluggable-scanner-spec/issues/2 to get more info
func hasCapability(r *models.Registration, a *artifact.Artifact) bool {
	// use allowlist here because currently only docker image is supported by the scanner
	// https://github.com/goharbor/pluggable-scanner-spec/issues/2
	allowlist := []string{image.ArtifactTypeImage}
	for _, t := range allowlist {
		if a.Type == t {
			return r.HasCapability(a.ManifestMediaType)
		}
	}

	return false
}
