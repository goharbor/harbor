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
	"github.com/goharbor/harbor/src/lib/selector"
	"testing"
	"time"

	"github.com/goharbor/harbor/src/pkg/retention/policy/action"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
)

// IndexTestSuite tests the rule index
type IndexTestSuite struct {
	suite.Suite

	candidates []*selector.Candidate
}

// TestIndexEntry is entry of IndexTestSuite
func TestIndexEntry(t *testing.T) {
	suite.Run(t, new(IndexTestSuite))
}

// SetupSuite ...
func (suite *IndexTestSuite) SetupSuite() {
	Register("fakeAction", newFakePerformer)

	suite.candidates = []*selector.Candidate{{
		Namespace:  "library",
		Repository: "harbor",
		Kind:       "image",
		Tags:       []string{"latest"},
		PushedTime: time.Now().Unix(),
		Labels:     []string{"L1", "L2"},
	}}
}

// TestRegister tests register
func (suite *IndexTestSuite) TestGet() {
	p, err := Get("fakeAction", nil, false)
	require.NoError(suite.T(), err)
	require.NotNil(suite.T(), p)

	results, err := p.Perform(suite.candidates)
	require.NoError(suite.T(), err)
	assert.Equal(suite.T(), 1, len(results))
	assert.Condition(suite.T(), func() (success bool) {
		r := results[0]
		success = r.Target != nil &&
			r.Error == nil &&
			r.Target.Repository == "harbor" &&
			r.Target.Tags[0] == "latest"

		return
	})
}

type fakePerformer struct {
	parameters interface{}
	isDryRun   bool
}

// Perform the artifacts
func (p *fakePerformer) Perform(candidates []*selector.Candidate) (results []*selector.Result, err error) {
	for _, c := range candidates {
		results = append(results, &selector.Result{
			Target: c,
		})
	}

	return
}

func newFakePerformer(params interface{}, isDryRun bool) action.Performer {
	return &fakePerformer{
		parameters: params,
		isDryRun:   isDryRun,
	}
}
