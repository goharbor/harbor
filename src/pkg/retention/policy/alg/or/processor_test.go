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

package or

import (
	"errors"
	"github.com/goharbor/harbor/src/common/dao"
	"testing"
	"time"

	"github.com/goharbor/harbor/src/pkg/art"
	"github.com/goharbor/harbor/src/pkg/art/selectors/doublestar"
	"github.com/goharbor/harbor/src/pkg/art/selectors/label"
	"github.com/goharbor/harbor/src/pkg/retention/dep"
	"github.com/goharbor/harbor/src/pkg/retention/policy/action"
	"github.com/goharbor/harbor/src/pkg/retention/policy/alg"
	"github.com/goharbor/harbor/src/pkg/retention/policy/rule"
	"github.com/goharbor/harbor/src/pkg/retention/policy/rule/always"
	"github.com/goharbor/harbor/src/pkg/retention/policy/rule/lastx"
	"github.com/goharbor/harbor/src/pkg/retention/policy/rule/latestps"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
)

// ProcessorTestSuite is suite for testing processor
type ProcessorTestSuite struct {
	suite.Suite

	all []*art.Candidate

	oldClient dep.Client
}

// TestProcessor is entrance for ProcessorTestSuite
func TestProcessor(t *testing.T) {
	suite.Run(t, new(ProcessorTestSuite))
}

// SetupSuite ...
func (suite *ProcessorTestSuite) SetupSuite() {
	dao.PrepareTestForPostgresSQL()
	suite.all = []*art.Candidate{
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
}

// TearDownSuite ...
func (suite *ProcessorTestSuite) TearDownSuite() {
	dep.DefaultClient = suite.oldClient
}

// TestProcess tests process method
func (suite *ProcessorTestSuite) TestProcess() {

	perf := action.NewRetainAction(suite.all, false)

	params := make([]*alg.Parameter, 0)
	lastxParams := make(map[string]rule.Parameter)
	lastxParams[lastx.ParameterX] = 10
	params = append(params, &alg.Parameter{
		Evaluator: lastx.New(lastxParams),
		Selectors: []art.Selector{
			doublestar.New(doublestar.Matches, "*dev*"),
			label.New(label.With, "L1,L2"),
		},
		Performer: perf,
	})

	latestKParams := make(map[string]rule.Parameter)
	latestKParams[latestps.ParameterK] = 10
	params = append(params, &alg.Parameter{
		Evaluator: latestps.New(latestKParams),
		Selectors: []art.Selector{
			label.New(label.With, "L3"),
		},
		Performer: perf,
	})

	p := New(params)

	results, err := p.Process(suite.all)
	require.NoError(suite.T(), err)
	assert.Equal(suite.T(), 1, len(results))
	assert.Condition(suite.T(), func() bool {
		for _, r := range results {
			if r.Error != nil {
				return false
			}
		}

		return true
	}, "no errors in the returned result list")
}

// TestProcess2 ...
func (suite *ProcessorTestSuite) TestProcess2() {
	perf := action.NewRetainAction(suite.all, false)

	params := make([]*alg.Parameter, 0)
	alwaysParams := make(map[string]rule.Parameter)
	params = append(params, &alg.Parameter{
		Evaluator: always.New(alwaysParams),
		Selectors: []art.Selector{
			doublestar.New(doublestar.Matches, "latest"),
			label.New(label.With, ""),
		},
		Performer: perf,
	})

	p := New(params)

	results, err := p.Process(suite.all)
	require.NoError(suite.T(), err)
	assert.Equal(suite.T(), 1, len(results))
	assert.Condition(suite.T(), func() bool {
		found := false
		for _, r := range results {
			if r.Error != nil {
				return false
			}

			if r.Target.Tags[0] == "dev" {
				found = true
			}
		}

		return found
	}, "no errors in the returned result list")

}

type fakeRetentionClient struct{}

// GetCandidates ...
func (frc *fakeRetentionClient) GetCandidates(repo *art.Repository) ([]*art.Candidate, error) {
	return nil, errors.New("not implemented")
}

// Delete ...
func (frc *fakeRetentionClient) Delete(candidate *art.Candidate) error {
	return nil
}

// DeleteRepository ...
func (frc *fakeRetentionClient) DeleteRepository(repo *art.Repository) error {
	panic("implement me")
}
