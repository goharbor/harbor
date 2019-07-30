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

package retention

import (
	"testing"

	"github.com/goharbor/harbor/src/jobservice/job"
	"github.com/goharbor/harbor/src/pkg/retention/dep"
	"github.com/goharbor/harbor/src/pkg/retention/res"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
)

// DelRepoJobSuite tests the del repository job
type DelRepoJobSuite struct {
	suite.Suite

	oldClient dep.Client
}

// TestJob is entry of running JobTestSuite
func TestDelRepoJob(t *testing.T) {
	suite.Run(t, new(DelRepoJobSuite))
}

// SetupSuite ...
func (suite *DelRepoJobSuite) SetupSuite() {
	suite.oldClient = dep.DefaultClient
	dep.DefaultClient = &fakeRetentionClient{}
}

// TearDownSuite ...
func (suite *DelRepoJobSuite) TearDownSuite() {
	dep.DefaultClient = suite.oldClient
}

// TestRun ...
func (suite *DelRepoJobSuite) TestRun() {
	params := make(job.Parameters)
	params[ParamDryRun] = false
	repository := &res.Repository{
		Namespace: "library",
		Name:      "harbor",
		Kind:      res.Image,
	}
	repoJSON, err := repository.ToJSON()
	require.NoError(suite.T(), err)
	params[ParamRepo] = repoJSON

	j := &DelRepoJob{}
	err = j.Validate(params)
	require.NoError(suite.T(), err)

	err = j.Run(&fakeJobContext{}, params)
	require.NoError(suite.T(), err)
}
