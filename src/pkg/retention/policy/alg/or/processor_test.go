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
	"testing"
	"time"

	"github.com/goharbor/harbor/src/pkg/retention/dep"

	"github.com/goharbor/harbor/src/pkg/retention/policy/lwp"

	"github.com/goharbor/harbor/src/pkg/retention/policy/action"
	"github.com/goharbor/harbor/src/pkg/retention/policy/alg"
	"github.com/goharbor/harbor/src/pkg/retention/policy/rule"
	"github.com/goharbor/harbor/src/pkg/retention/policy/rule/lastx"
	"github.com/goharbor/harbor/src/pkg/retention/policy/rule/latestk"
	"github.com/goharbor/harbor/src/pkg/retention/res"
	"github.com/goharbor/harbor/src/pkg/retention/res/selectors/doublestar"
	"github.com/goharbor/harbor/src/pkg/retention/res/selectors/label"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
)

// ProcessorTestSuite is suite for testing processor
type ProcessorTestSuite struct {
	suite.Suite

	p   alg.Processor
	all []*res.Candidate

	oldClient dep.Client
}

// TestProcessor is entrance for ProcessorTestSuite
func TestProcessor(t *testing.T) {
	suite.Run(t, new(ProcessorTestSuite))
}

// SetupSuite ...
func (suite *ProcessorTestSuite) SetupSuite() {
	suite.all = []*res.Candidate{
		{
			Namespace:  "library",
			Repository: "harbor",
			Kind:       "image",
			Tag:        "latest",
			PushedTime: time.Now().Unix(),
			Labels:     []string{"L1", "L2"},
		},
		{
			Namespace:  "library",
			Repository: "harbor",
			Kind:       "image",
			Tag:        "dev",
			PushedTime: time.Now().Unix(),
			Labels:     []string{"L3"},
		},
	}

	params := make([]*alg.Parameter, 0)

	perf := action.NewRetainAction(suite.all)

	lastxParams := make(map[string]rule.Parameter)
	lastxParams[lastx.ParameterX] = 10
	params = append(params, &alg.Parameter{
		Evaluator: lastx.New(lastxParams),
		Selectors: []res.Selector{
			doublestar.New(doublestar.Matches, "*dev*"),
			label.New(label.With, "L1,L2"),
		},
		Performer: perf,
	})

	latestKParams := make(map[string]rule.Parameter)
	latestKParams[latestk.ParameterK] = 10
	params = append(params, &alg.Parameter{
		Evaluator: latestk.New(latestKParams),
		Selectors: []res.Selector{
			label.New(label.With, "L3"),
		},
		Performer: perf,
	})

	p, err := alg.Get(alg.AlgorithmOR, params)
	require.NoError(suite.T(), err)

	suite.p = p

	suite.oldClient = dep.DefaultClient
	dep.DefaultClient = &fakeRetentionClient{}
}

// TearDownSuite ...
func (suite *ProcessorTestSuite) TearDownSuite() {
	dep.DefaultClient = suite.oldClient
}

// TestProcess tests process method
func (suite *ProcessorTestSuite) TestProcess() {
	results, err := suite.p.Process(suite.all)
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

type fakeRetentionClient struct{}

// GetCandidates ...
func (frc *fakeRetentionClient) GetCandidates(repo *res.Repository) ([]*res.Candidate, error) {
	return nil, errors.New("not implemented")
}

// Delete ...
func (frc *fakeRetentionClient) Delete(candidate *res.Candidate) error {
	return nil
}

// SubmitTask ...
func (frc *fakeRetentionClient) SubmitTask(taskID int64, repository *res.Repository, meta *lwp.Metadata) (string, error) {
	return "", errors.New("not implemented")
}
