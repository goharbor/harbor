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

package signature

import (
	"github.com/goharbor/harbor/src/common/security"
	"github.com/goharbor/harbor/src/core/config"
	"github.com/goharbor/harbor/src/lib/log"
	"github.com/goharbor/harbor/src/pkg/signature/notary"
	"github.com/goharbor/harbor/src/pkg/signature/notary/model"
	"golang.org/x/net/context"
)

// Checker checks the signature status of artifact
type Checker struct {
	signatures map[string]string
}

// IsTagSigned checks if the tag of the artifact is signed, it also checks the signed artifact has the same digest as parm.
func (sc Checker) IsTagSigned(tag, digest string) bool {
	d, ok := sc.signatures[tag]
	if len(digest) == 0 {
		return ok
	}
	return digest == d
}

// IsArtifactSigned checks if the artifact with given digest is signed.
func (sc Checker) IsArtifactSigned(digest string) bool {
	for _, v := range sc.signatures {
		if v == digest {
			return true
		}
	}
	return false
}

// Manager interface for handling signatures of artifacts
type Manager interface {
	// GetCheckerByRepo returns a Checker for checking signature
	GetCheckerByRepo(ctx context.Context, repo string) (*Checker, error)
}

type mgr struct {
}

// GetCheckerByRepo ...
func (m *mgr) GetCheckerByRepo(ctx context.Context, repo string) (*Checker, error) {
	if !config.WithNotary() { // return a checker that always return false
		return &Checker{}, nil
	}
	s := make(map[string]string)
	targets, err := m.getTargetsByRepo(ctx, repo)
	if err != nil {
		return nil, err
	}
	for _, t := range targets {
		if d, err := notary.DigestFromTarget(t); err != nil {
			log.Warningf("Failed to get signed digest for tag %s, error: %v, skip", t.Tag, err)
		} else {
			s[t.Tag] = d
		}
	}
	return &Checker{s}, nil
}

func (m *mgr) getTargetsByRepo(ctx context.Context, repo string) ([]model.Target, error) {
	name := "unknown"
	if sc, ok := security.FromContext(ctx); !ok || sc == nil {
		log.Warningf("Unable to get security context")
	} else {
		name = sc.GetUsername()
	}
	return notary.GetInternalTargets(config.InternalNotaryEndpoint(), name, repo)
}

var instance = &mgr{}

// GetManager ...
func GetManager() Manager {
	return instance
}
