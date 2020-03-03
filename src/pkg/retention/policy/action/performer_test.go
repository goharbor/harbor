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

package action

import (
	"github.com/goharbor/harbor/src/common/dao"
	"github.com/goharbor/harbor/src/pkg/artifactselector"
	"github.com/goharbor/harbor/src/pkg/immutabletag"
	"testing"
	"time"

	immumodel "github.com/goharbor/harbor/src/pkg/immutabletag/model"
	"github.com/goharbor/harbor/src/pkg/retention/dep"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
)

// TestPerformerSuite tests the performer related function
type TestPerformerSuite struct {
	suite.Suite

	oldClient dep.Client
	all       []*artifactselector.Candidate
}

// TestPerformer is the entry of the TestPerformerSuite
func TestPerformer(t *testing.T) {
	suite.Run(t, new(TestPerformerSuite))
}

// SetupSuite ...
func (suite *TestPerformerSuite) SetupSuite() {
	suite.all = []*artifactselector.Candidate{
		{
			Namespace:  "library",
			Repository: "harbor",
			Kind:       "image",
			Tags:       []string{"latest"},
			Digest:     "latest",
			PushedTime: time.Now().Unix(),
			Labels:     []string{"L1", "L2"},
		},
		{
			Namespace:  "library",
			Repository: "harbor",
			Kind:       "image",
			Tags:       []string{"dev"},
			Digest:     "dev",
			PushedTime: time.Now().Unix(),
			Labels:     []string{"L3"},
		},
	}

	suite.oldClient = dep.DefaultClient
	dep.DefaultClient = &fakeRetentionClient{}
	dao.PrepareTestForPostgresSQL()
}

// TearDownSuite ...
func (suite *TestPerformerSuite) TearDownSuite() {
	dep.DefaultClient = suite.oldClient
}

// TestPerform tests Perform action
func (suite *TestPerformerSuite) TestPerform() {
	p := &retainAction{
		all: suite.all,
	}

	candidates := []*artifactselector.Candidate{
		{
			Namespace:  "library",
			Repository: "harbor",
			Kind:       "image",
			Tags:       []string{"latest"},
			Digest:     "latest",
			PushedTime: time.Now().Unix(),
			Labels:     []string{"L1", "L2"},
		},
	}

	results, err := p.Perform(candidates)
	require.NoError(suite.T(), err)
	require.Equal(suite.T(), 1, len(results))
	require.NotNil(suite.T(), results[0].Target)
	assert.NoError(suite.T(), results[0].Error)
	assert.Equal(suite.T(), "dev", results[0].Target.Tags[0])
}

// TestPerform tests Perform action
func (suite *TestPerformerSuite) TestPerformImmutable() {
	all := []*artifactselector.Candidate{
		{
			NamespaceID: 1,
			Namespace:   "library",
			Repository:  "harbor",
			Kind:        "image",
			Tags:        []string{"latest"},
			Digest:      "d0",
			PushedTime:  time.Now().Unix(),
			Labels:      []string{"L1", "L2"},
		},
		{
			NamespaceID: 1,
			Namespace:   "library",
			Repository:  "harbor",
			Kind:        "image",
			Tags:        []string{"dev"},
			Digest:      "d1",
			PushedTime:  time.Now().Unix(),
			Labels:      []string{"L3"},
		},
		{
			NamespaceID: 1,
			Namespace:   "library",
			Repository:  "test",
			Kind:        "image",
			Tags:        []string{"immute"},
			Digest:      "d2",
			PushedTime:  time.Now().Unix(),
			Labels:      []string{"L1", "L2"},
		},
		{
			NamespaceID: 1,
			Namespace:   "library",
			Repository:  "test",
			Kind:        "image",
			Tags:        []string{"samedig"},
			Digest:      "d2",
			PushedTime:  time.Now().Unix(),
			Labels:      []string{"L1", "L2"},
		},
	}
	p := &retainAction{
		all: all,
	}

	rule := &immumodel.Metadata{
		ProjectID: 1,
		Priority:  1,
		Action:    "immutable",
		Template:  "immutable_template",
		TagSelectors: []*immumodel.Selector{
			{
				Kind:       "doublestar",
				Decoration: "matches",
				Pattern:    "immute",
			},
		},
		ScopeSelectors: map[string][]*immumodel.Selector{
			"repository": {
				{
					Kind:       "doublestar",
					Decoration: "repoMatches",
					Pattern:    "**",
				},
			},
		},
	}
	imid, e := immutabletag.ImmuCtr.CreateImmutableRule(rule)
	assert.NoError(suite.T(), e)
	defer func() {
		assert.NoError(suite.T(), immutabletag.ImmuCtr.DeleteImmutableRule(imid))
	}()

	candidates := []*artifactselector.Candidate{
		{
			NamespaceID: 1,
			Namespace:   "library",
			Repository:  "harbor",
			Kind:        "image",
			Tags:        []string{"latest"},
			Digest:      "d0",
			PushedTime:  time.Now().Unix(),
			Labels:      []string{"L1", "L2"},
		},
	}

	results, err := p.Perform(candidates)
	require.NoError(suite.T(), err)
	require.Equal(suite.T(), 3, len(results))
	for _, r := range results {
		require.NotNil(suite.T(), r.Target)
		if r.Target.Digest == "d1" {
			require.NoError(suite.T(), r.Error)
			require.Equal(suite.T(), "dev", r.Target.Tags[0])
		} else if r.Target.Digest == "d2" {
			require.Error(suite.T(), r.Error)
			require.IsType(suite.T(), (*artifactselector.ImmutableError)(nil), r.Error)
		} else {
			require.Fail(suite.T(), "should not delete "+r.Target.Hash())
		}
	}
	require.NotNil(suite.T(), results[0].Target)
	assert.NoError(suite.T(), results[0].Error)
	assert.Equal(suite.T(), "dev", results[0].Target.Tags[0])
}

type fakeRetentionClient struct{}

// GetCandidates ...
func (frc *fakeRetentionClient) GetCandidates(repo *artifactselector.Repository) ([]*artifactselector.Candidate, error) {
	return nil, errors.New("not implemented")
}

// Delete ...
func (frc *fakeRetentionClient) Delete(candidate *artifactselector.Candidate) error {
	return nil
}

// DeleteRepository ...
func (frc *fakeRetentionClient) DeleteRepository(repo *artifactselector.Repository) error {
	panic("implement me")
}
