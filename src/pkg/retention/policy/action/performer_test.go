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
	"testing"
	"time"

	"github.com/goharbor/harbor/src/pkg/retention/dep"
	"github.com/goharbor/harbor/src/pkg/retention/res"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
)

// TestPerformerSuite tests the performer related function
type TestPerformerSuite struct {
	suite.Suite

	oldClient dep.Client
	all       []*res.Candidate
}

// TestPerformer is the entry of the TestPerformerSuite
func TestPerformer(t *testing.T) {
	suite.Run(t, new(TestPerformerSuite))
}

// SetupSuite ...
func (suite *TestPerformerSuite) SetupSuite() {
	suite.all = []*res.Candidate{
		{
			Namespace:  "library",
			Repository: "harbor",
			Kind:       "image",
			Tag:        "latest",
			Digest:     "latest",
			PushedTime: time.Now().Unix(),
			Labels:     []string{"L1", "L2"},
		},
		{
			Namespace:  "library",
			Repository: "harbor",
			Kind:       "image",
			Tag:        "dev",
			Digest:     "dev",
			PushedTime: time.Now().Unix(),
			Labels:     []string{"L3"},
		},
	}

	suite.oldClient = dep.DefaultClient
	dep.DefaultClient = &fakeRetentionClient{}
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

	candidates := []*res.Candidate{
		{
			Namespace:  "library",
			Repository: "harbor",
			Kind:       "image",
			Tag:        "latest",
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
	assert.Equal(suite.T(), "dev", results[0].Target.Tag)
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

// DeleteRepository ...
func (frc *fakeRetentionClient) DeleteRepository(repo *res.Repository) error {
	panic("implement me")
}
