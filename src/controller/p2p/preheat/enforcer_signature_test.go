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

package preheat

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	car "github.com/goharbor/harbor/src/controller/artifact"
	"github.com/goharbor/harbor/src/controller/scan"
	"github.com/goharbor/harbor/src/controller/tag"
	"github.com/goharbor/harbor/src/lib/selector"
	accmodel "github.com/goharbor/harbor/src/pkg/accessory/model"
	"github.com/goharbor/harbor/src/pkg/accessory/model/cosign"
	"github.com/goharbor/harbor/src/pkg/accessory/model/notation"
	ar "github.com/goharbor/harbor/src/pkg/artifact"
	po "github.com/goharbor/harbor/src/pkg/p2p/preheat/models/policy"
	pr "github.com/goharbor/harbor/src/pkg/p2p/preheat/models/provider"
	"github.com/goharbor/harbor/src/pkg/p2p/preheat/policy"
	"github.com/goharbor/harbor/src/pkg/p2p/preheat/provider"
	"github.com/goharbor/harbor/src/pkg/p2p/preheat/provider/auth"
	proModels "github.com/goharbor/harbor/src/pkg/project/models"
	"github.com/goharbor/harbor/src/pkg/scan/vuln"
	ta "github.com/goharbor/harbor/src/pkg/tag/model/tag"
	"github.com/goharbor/harbor/src/testing/controller/artifact"
	"github.com/goharbor/harbor/src/testing/controller/project"
	scantesting "github.com/goharbor/harbor/src/testing/controller/scan"
	"github.com/goharbor/harbor/src/testing/mock"
	"github.com/goharbor/harbor/src/testing/pkg/p2p/preheat/instance"
	prepolicy "github.com/goharbor/harbor/src/testing/pkg/p2p/preheat/policy"
	"github.com/goharbor/harbor/src/testing/pkg/task"
)

// signedImage builds an image artifact that carries the given signature accessories.
func signedImage(id int64, repo, tagName string, signatureTypes ...string) *car.Artifact {
	art := &car.Artifact{
		Artifact: ar.Artifact{
			ID:             id,
			Type:           "image",
			ProjectID:      1,
			RepositoryName: "library/" + repo,
			Digest:         "sha256:" + repo,
		},
		Tags: []*tag.Tag{{Tag: ta.Tag{Name: tagName}}},
	}
	for _, typ := range signatureTypes {
		data := accmodel.AccessoryData{ArtifactID: id + 1000, SubArtifactID: id, Type: typ}
		switch typ {
		case accmodel.TypeCosignSignature:
			art.Accessories = append(art.Accessories, cosign.New(data))
		case accmodel.TypeNotationSignature:
			art.Accessories = append(art.Accessories, notation.New(data))
		}
	}
	return art
}

func cosignTrustProject() *proModels.Project {
	return &proModels.Project{
		ProjectID: 1,
		Name:      "library",
		Metadata:  map[string]string{proModels.ProMetaEnableContentTrustCosign: "true"},
	}
}

func scanCtlWithLowSeverity() *scantesting.Controller {
	low := vuln.Low
	scanCtl := &scantesting.Controller{}
	scanCtl.On("GetVulnerable",
		context.TODO(),
		mock.AnythingOfType("*artifact.Artifact"),
		mock.AnythingOfType("models.CVESet"),
		mock.AnythingOfType("bool"),
	).Return(&scan.Vulnerable{Severity: &low, ScanStatus: "Success"}, nil)
	return scanCtl
}

// TestToCandidatesSignatures checks the candidates tell which tags are signed
func TestToCandidatesSignatures(t *testing.T) {
	enforcer := &defaultEnforcer{scanCtl: scanCtlWithLowSeverity()}
	arts := []*car.Artifact{
		signedImage(1, "cosigned", "v1", accmodel.TypeCosignSignature),
		signedImage(2, "unsigned", "v1"),
		signedImage(3, "notation-only", "v1", accmodel.TypeNotationSignature),
	}

	candidates, err := enforcer.toCandidates(context.TODO(), cosignTrustProject(), arts)
	require.NoError(t, err)
	require.Len(t, candidates, 3)

	signed := map[string]bool{}
	for _, c := range candidates {
		signed[c.Repository] = c.Signatures[c.Tags[0]]
	}
	assert.Equal(t, map[string]bool{
		"cosigned":      true,
		"unsigned":      false,
		"notation-only": false, // the project asks for cosign signatures
	}, signed)
}

// TestOverrideSecuritySettingsContentTrust checks the project content trust settings switch on
// the signature criteria of the policy.
func TestOverrideSecuritySettingsContentTrust(t *testing.T) {
	signatureFilter := func(p *po.Schema) *po.Filter {
		for _, f := range p.Filters {
			if f.Type == po.FilterTypeSignature {
				return f
			}
		}
		return nil
	}

	cases := []struct {
		name     string
		metadata map[string]string
		// the signature filter that the policy had before the override
		policyFilter *po.Filter
		want         bool
	}{
		{"cosign content trust", map[string]string{proModels.ProMetaEnableContentTrustCosign: "true"}, nil, true},
		{"notation content trust", map[string]string{proModels.ProMetaEnableContentTrust: "true"}, nil, true},
		{"no content trust", map[string]string{}, nil, false},
		{"no content trust drops the stored criteria", map[string]string{}, &po.Filter{Type: po.FilterTypeSignature, Value: true}, false},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			pl := &po.Schema{Name: "p", Filters: []*po.Filter{{Type: po.FilterTypeRepository, Value: "**"}}}
			if c.policyFilter != nil {
				pl.Filters = append(pl.Filters, c.policyFilter)
			}
			pro := &proModels.Project{ProjectID: 1, Name: "library", Metadata: c.metadata}

			overrideSecuritySettings(pl, pro)

			f := signatureFilter(pl)
			require.Equal(t, c.want, f != nil)
			if c.want {
				assert.Equal(t, true, f.Value)
			}
		})
	}
}

// TestSignatureCriteriaFiltersUnsignedArtifacts runs the whole filtering chain of a policy for a
// project that requires cosign signatures: the unsigned artifacts must not be preheated.
func TestSignatureCriteriaFiltersUnsignedArtifacts(t *testing.T) {
	enforcer := &defaultEnforcer{scanCtl: scanCtlWithLowSeverity()}
	pro := cosignTrustProject()
	candidates, err := enforcer.toCandidates(context.TODO(), pro, []*car.Artifact{
		signedImage(1, "signed", "v1", accmodel.TypeCosignSignature),
		signedImage(2, "unsigned", "v1"),
	})
	require.NoError(t, err)

	pl := &po.Schema{Name: "p", Filters: []*po.Filter{{Type: po.FilterTypeRepository, Value: "**"}}}
	overrideSecuritySettings(pl, pro)

	filtered, err := policy.NewFilter().BuildFrom(pl).Filter(candidates)
	require.NoError(t, err)
	var repos []string
	for _, c := range filtered {
		repos = append(repos, c.Repository)
	}
	assert.Equal(t, []string{"signed"}, repos)
}

// TestPreheatArtifactSignatureCriteria checks the event based preheating of a project that
// requires cosign signatures. The artifact carried by the events has no accessories, so the
// enforcer has to look them up to tell whether the artifact is signed.
func TestPreheatArtifactSignatureCriteria(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	cases := []struct {
		name string
		// the artifact as stored, with its accessories
		stored *car.Artifact
		want   int
	}{
		{"signed artifact is preheated", signedImage(1, "app", "v1", accmodel.TypeCosignSignature), 1},
		{"unsigned artifact is skipped", signedImage(1, "app", "v1"), 0},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			pl := &po.Schema{
				ID:         1,
				Name:       "event_based_policy",
				ProjectID:  1,
				ProviderID: 1,
				Filters:    []*po.Filter{{Type: po.FilterTypeRepository, Value: "**"}},
				Trigger:    &po.Trigger{Type: po.TriggerTypeEventBased},
				Enabled:    true,
			}
			policyMgr := &prepolicy.Manager{}
			policyMgr.On("ListPoliciesByProject", context.TODO(), mock.AnythingOfType("int64"), mock.AnythingOfType("*q.Query")).
				Return([]*po.Schema{pl}, nil)

			proCtl := &project.Controller{}
			proCtl.On("Get", context.TODO(), int64(1), mock.Anything, mock.Anything).Return(cosignTrustProject(), nil)

			artCtl := &artifact.Controller{}
			artCtl.On("Get", context.TODO(), int64(1), mock.Anything).Return(c.stored, nil)

			instMgr := &instance.Manager{}
			instMgr.On("Get", context.TODO(), mock.AnythingOfType("int64")).Return(&pr.Instance{
				ID:       1,
				Name:     "kraken",
				Vendor:   provider.DriverKraken,
				Endpoint: server.URL,
				Status:   provider.DriverStatusHealthy,
				AuthMode: auth.AuthModeNone,
			}, nil)

			execMgr := &task.ExecutionManager{}
			execMgr.On("Create", context.TODO(), mock.AnythingOfType("string"), mock.AnythingOfType("int64"),
				mock.AnythingOfType("string"), mock.AnythingOfType("map[string]interface {}")).Return(int64(1), nil)
			taskMgr := &task.Manager{}
			taskMgr.On("Create", context.TODO(), mock.AnythingOfType("int64"), mock.AnythingOfType("*task.Job"),
				mock.AnythingOfType("map[string]interface {}")).Return(int64(1), nil)

			enforcer := &defaultEnforcer{
				policyMgr:    policyMgr,
				executionMgr: execMgr,
				taskMgr:      taskMgr,
				artCtl:       artCtl,
				scanCtl:      scanCtlWithLowSeverity(),
				proCtl:       proCtl,
				instMgr:      instMgr,
				fullURLGetter: func(c *selector.Candidate) (string, error) {
					return "https://testing.harbor.com/v2/" + c.Namespace + "/" + c.Repository + "/manifests/" + c.Tags[0], nil
				},
				credMaker: func(context.Context, *selector.Candidate) (string, error) {
					return "fake-token", nil
				},
			}

			// the artifact of the event carries no accessories
			ids, err := enforcer.PreheatArtifact(context.TODO(), signedImage(1, "app", "v1"))
			require.NoError(t, err)
			assert.Len(t, ids, c.want)
		})
	}
}

// TestGetCandidatesSignatures checks the artifacts of a manual or scheduled run are listed with
// their accessories, but only when the project requires signed artifacts as it costs a query
// per artifact.
func TestGetCandidatesSignatures(t *testing.T) {
	cases := []struct {
		name           string
		project        *proModels.Project
		wantAccessory  bool
		wantSignedTags map[string]bool
	}{
		{
			name:           "cosign content trust",
			project:        cosignTrustProject(),
			wantAccessory:  true,
			wantSignedTags: map[string]bool{"signed": true, "unsigned": false},
		},
		{
			name:          "no content trust",
			project:       &proModels.Project{ProjectID: 1, Name: "library"},
			wantAccessory: false,
		},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			var option *car.Option
			artCtl := &artifact.Controller{}
			artCtl.On("List", context.TODO(), mock.AnythingOfType("*q.Query"), mock.AnythingOfType("*artifact.Option")).
				Return([]*car.Artifact{
					signedImage(1, "signed", "v1", accmodel.TypeCosignSignature),
					signedImage(2, "unsigned", "v1"),
				}, nil).
				Run(func(args mock.Arguments) { option = args.Get(2).(*car.Option) })
			enforcer := &defaultEnforcer{artCtl: artCtl, scanCtl: scanCtlWithLowSeverity()}

			candidates, err := enforcer.getCandidates(context.TODO(), &po.Schema{Name: "p", ProjectID: 1}, c.project)
			require.NoError(t, err)
			require.NotNil(t, option)
			assert.Equal(t, c.wantAccessory, option.WithAccessory)

			signed := map[string]bool{}
			for _, cand := range candidates {
				if c.wantSignedTags == nil {
					assert.Nil(t, cand.Signatures)
					continue
				}
				signed[cand.Repository] = cand.Signatures[cand.Tags[0]]
			}
			if c.wantSignedTags != nil {
				assert.Equal(t, map[string]bool{"signed": true, "unsigned": false}, signed)
			}
		})
	}
}
