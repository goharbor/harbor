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

package policy

import (
	"github.com/goharbor/harbor/src/common/dao"
	"testing"
	"time"

	index3 "github.com/goharbor/harbor/src/pkg/retention/policy/action/index"

	index2 "github.com/goharbor/harbor/src/pkg/retention/policy/alg/index"

	"github.com/goharbor/harbor/src/pkg/art/selectors/index"

	"github.com/goharbor/harbor/src/pkg/retention/dep"

	"github.com/pkg/errors"

	"github.com/goharbor/harbor/src/pkg/retention/policy/alg/or"

	"github.com/goharbor/harbor/src/pkg/art/selectors/label"

	"github.com/goharbor/harbor/src/pkg/art/selectors/doublestar"

	"github.com/goharbor/harbor/src/pkg/retention/policy/rule/latestps"

	"github.com/goharbor/harbor/src/pkg/retention/policy/action"

	"github.com/stretchr/testify/assert"

	"github.com/stretchr/testify/require"

	"github.com/goharbor/harbor/src/pkg/retention/policy/rule"

	"github.com/goharbor/harbor/src/pkg/retention/policy/lwp"

	"github.com/goharbor/harbor/src/pkg/art"

	"github.com/stretchr/testify/suite"
)

// TestBuilderSuite is the suite to test builder
type TestBuilderSuite struct {
	suite.Suite

	all       []*art.Candidate
	oldClient dep.Client
}

// TestBuilder is the entry of testing TestBuilderSuite
func TestBuilder(t *testing.T) {
	suite.Run(t, new(TestBuilderSuite))
}

// SetupSuite prepares the testing content if needed
func (suite *TestBuilderSuite) SetupSuite() {
	dao.PrepareTestForPostgresSQL()
	suite.all = []*art.Candidate{
		{
			NamespaceID: 1,
			Namespace:   "library",
			Repository:  "harbor",
			Kind:        "image",
			Tags:        []string{"latest"},
			Digest:      "latest",
			PushedTime:  time.Now().Unix(),
			Labels:      []string{"L1", "L2"},
		},
		{
			NamespaceID: 1,
			Namespace:   "library",
			Repository:  "harbor",
			Kind:        "image",
			Tags:        []string{"dev"},
			Digest:      "dev",
			PushedTime:  time.Now().Unix(),
			Labels:      []string{"L3"},
		},
	}

	index2.Register(index2.AlgorithmOR, or.New)
	index.Register(doublestar.Kind, []string{
		doublestar.Matches,
		doublestar.Excludes,
		doublestar.RepoMatches,
		doublestar.RepoExcludes,
		doublestar.NSMatches,
		doublestar.NSExcludes,
	}, doublestar.New)
	index.Register(label.Kind, []string{label.With, label.Without}, label.New)
	index3.Register(action.Retain, action.NewRetainAction)

	suite.oldClient = dep.DefaultClient
	dep.DefaultClient = &fakeRetentionClient{}
}

// TearDownSuite ...
func (suite *TestBuilderSuite) TearDownSuite() {
	dep.DefaultClient = suite.oldClient
}

// TestBuild tests the Build function
func (suite *TestBuilderSuite) TestBuild() {
	b := &basicBuilder{suite.all}

	params := make(rule.Parameters)
	params[latestps.ParameterK] = 10

	scopeSelectors := make(map[string][]*rule.Selector, 1)
	scopeSelectors["repository"] = []*rule.Selector{{
		Kind:       doublestar.Kind,
		Decoration: doublestar.RepoMatches,
		Pattern:    "**",
	}}

	lm := &lwp.Metadata{
		Algorithm: AlgorithmOR,
		Rules: []*rule.Metadata{{
			ID:             1,
			Priority:       999,
			Action:         action.Retain,
			Template:       latestps.TemplateID,
			Parameters:     params,
			ScopeSelectors: scopeSelectors,
			TagSelectors: []*rule.Selector{
				{
					Kind:       doublestar.Kind,
					Decoration: doublestar.Matches,
					Pattern:    "latest",
				},
			},
		}},
	}

	p, err := b.Build(lm, false)
	require.NoError(suite.T(), err)
	require.NotNil(suite.T(), p)

	results, err := p.Process(suite.all)
	require.NoError(suite.T(), err)
	assert.Equal(suite.T(), 1, len(results))
	assert.Condition(suite.T(), func() (success bool) {
		art := results[0]
		success = art.Error == nil &&
			art.Target != nil &&
			art.Target.Repository == "harbor" &&
			art.Target.Tags[0] == "dev"

		return
	})
}

type fakeRetentionClient struct{}

func (frc *fakeRetentionClient) DeleteRepository(repo *art.Repository) error {
	panic("implement me")
}

// GetCandidates ...
func (frc *fakeRetentionClient) GetCandidates(repo *art.Repository) ([]*art.Candidate, error) {
	return nil, errors.New("not implemented")
}

// Delete ...
func (frc *fakeRetentionClient) Delete(candidate *art.Candidate) error {
	return nil
}

// SubmitTask ...
func (frc *fakeRetentionClient) SubmitTask(taskID int64, repository *art.Repository, meta *lwp.Metadata) (string, error) {
	return "", errors.New("not implemented")
}
