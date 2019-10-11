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

package index

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/stretchr/testify/require"

	"github.com/goharbor/harbor/src/pkg/art"
	"github.com/goharbor/harbor/src/pkg/retention/policy/rule"

	"github.com/stretchr/testify/suite"
)

// IndexTestSuite tests the rule index
type IndexTestSuite struct {
	suite.Suite
}

// TestIndexEntry is entry of IndexTestSuite
func TestIndexEntry(t *testing.T) {
	suite.Run(t, new(IndexTestSuite))
}

// SetupSuite ...
func (suite *IndexTestSuite) SetupSuite() {
	Register(&Metadata{
		TemplateID: "fakeEvaluator",
		Action:     "retain",
		Parameters: []*IndexedParam{
			{
				Name:     "fakeParam",
				Type:     "int",
				Unit:     "count",
				Required: true,
			},
		},
	}, newFakeEvaluator)
}

// TestRegister tests register
func (suite *IndexTestSuite) TestGet() {

	params := make(rule.Parameters)
	params["fakeParam"] = 99
	evaluator, err := Get("fakeEvaluator", params)
	require.NoError(suite.T(), err)
	require.NotNil(suite.T(), evaluator)

	candidates := []*art.Candidate{{
		Namespace:  "library",
		Repository: "harbor",
		Kind:       "image",
		Tag:        "latest",
		PushedTime: time.Now().Unix(),
		Labels:     []string{"L1", "L2"},
	}}

	results, err := evaluator.Process(candidates)
	require.NoError(suite.T(), err)
	assert.Equal(suite.T(), 1, len(results))
	assert.Condition(suite.T(), func() bool {
		c := results[0]
		return c.Repository == "harbor" && c.Tag == "latest"
	})
}

// TestIndex tests Index
func (suite *IndexTestSuite) TestIndex() {
	metas := Index()
	require.Equal(suite.T(), 8, len(metas))
	assert.Condition(suite.T(), func() bool {
		for _, m := range metas {
			if m.TemplateID == "fakeEvaluator" &&
				m.Action == "retain" &&
				len(m.Parameters) > 0 {
				return true
			}
		}
		return false
	}, "check fake evaluator in index")
}

type fakeEvaluator struct {
	i int
}

// Process rule
func (e *fakeEvaluator) Process(artifacts []*art.Candidate) ([]*art.Candidate, error) {
	return artifacts, nil
}

// Action of the rule
func (e *fakeEvaluator) Action() string {
	return "retain"
}

// newFakeEvaluator is the factory of fakeEvaluator
func newFakeEvaluator(parameters rule.Parameters) rule.Evaluator {
	i := 10
	if v, ok := parameters["fakeParam"]; ok {
		i = v.(int)
	}

	return &fakeEvaluator{i}
}
