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
	"context"
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/goharbor/harbor/src/jobservice/job"
	"github.com/goharbor/harbor/src/jobservice/logger"
	"github.com/goharbor/harbor/src/pkg/retention/dep"
	"github.com/goharbor/harbor/src/pkg/retention/policy"
	"github.com/goharbor/harbor/src/pkg/retention/policy/action"
	"github.com/goharbor/harbor/src/pkg/retention/policy/alg"
	"github.com/goharbor/harbor/src/pkg/retention/policy/alg/or"
	"github.com/goharbor/harbor/src/pkg/retention/policy/lwp"
	"github.com/goharbor/harbor/src/pkg/retention/policy/rule"
	"github.com/goharbor/harbor/src/pkg/retention/policy/rule/latestk"
	"github.com/goharbor/harbor/src/pkg/retention/res"
	"github.com/goharbor/harbor/src/pkg/retention/res/selectors"
	"github.com/goharbor/harbor/src/pkg/retention/res/selectors/doublestar"
	"github.com/goharbor/harbor/src/pkg/retention/res/selectors/label"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
)

// JobTestSuite is test suite for testing job
type JobTestSuite struct {
	suite.Suite

	oldClient dep.Client
}

// TestJob is entry of running JobTestSuite
func TestJob(t *testing.T) {
	suite.Run(t, new(JobTestSuite))
}

// SetupSuite ...
func (suite *JobTestSuite) SetupSuite() {
	alg.Register(alg.AlgorithmOR, or.New)
	selectors.Register(doublestar.Kind, []string{
		doublestar.Matches,
		doublestar.Excludes,
		doublestar.RepoMatches,
		doublestar.RepoExcludes,
		doublestar.NSMatches,
		doublestar.NSExcludes,
	}, doublestar.New)
	selectors.Register(label.Kind, []string{label.With, label.Without}, label.New)
	action.Register(action.Retain, action.NewRetainAction)

	suite.oldClient = dep.DefaultClient
	dep.DefaultClient = &fakeRetentionClient{}
}

// TearDownSuite ...
func (suite *JobTestSuite) TearDownSuite() {
	dep.DefaultClient = suite.oldClient
}

func (suite *JobTestSuite) TestRunSuccess() {
	params := make(job.Parameters)
	params[ParamDryRun] = false
	params[ParamRepo] = &res.Repository{
		Namespace: "library",
		Name:      "harbor",
		Kind:      res.Image,
	}

	scopeSelectors := make(map[string][]*rule.Selector)
	scopeSelectors["project"] = []*rule.Selector{{
		Kind:       doublestar.Kind,
		Decoration: doublestar.RepoMatches,
		Pattern:    "{harbor}",
	}}

	ruleParams := make(rule.Parameters)
	ruleParams[latestk.ParameterK] = 10

	params[ParamMeta] = &lwp.Metadata{
		Algorithm: policy.AlgorithmOR,
		Rules: []*rule.Metadata{
			{
				ID:         1,
				Priority:   999,
				Action:     action.Retain,
				Template:   latestk.TemplateID,
				Parameters: ruleParams,
				TagSelectors: []*rule.Selector{{
					Kind:       label.Kind,
					Decoration: label.With,
					Pattern:    "L3",
				}, {
					Kind:       doublestar.Kind,
					Decoration: doublestar.Matches,
					Pattern:    "**",
				}},
				ScopeSelectors: scopeSelectors,
			},
		},
	}

	j := &Job{}
	err := j.Validate(params)
	require.NoError(suite.T(), err)

	err = j.Run(&fakeJobContext{}, params)
	require.NoError(suite.T(), err)
}

type fakeRetentionClient struct{}

// GetCandidates ...
func (frc *fakeRetentionClient) GetCandidates(repo *res.Repository) ([]*res.Candidate, error) {
	return []*res.Candidate{
		{
			Namespace:    "library",
			Repository:   "harbor",
			Kind:         "image",
			Tag:          "latest",
			PushedTime:   time.Now().Unix() - 11,
			PulledTime:   time.Now().Unix() - 2,
			CreationTime: time.Now().Unix() - 10,
			Labels:       []string{"L1", "L2"},
		},
		{
			Namespace:    "library",
			Repository:   "harbor",
			Kind:         "image",
			Tag:          "dev",
			PushedTime:   time.Now().Unix() - 10,
			PulledTime:   time.Now().Unix() - 3,
			CreationTime: time.Now().Unix() - 20,
			Labels:       []string{"L3"},
		},
	}, nil
}

// Delete ...
func (frc *fakeRetentionClient) Delete(candidate *res.Candidate) error {
	return nil
}

// SubmitTask ...
func (frc *fakeRetentionClient) SubmitTask(taskID int64, repository *res.Repository, meta *lwp.Metadata) (string, error) {
	return "", errors.New("not implemented")
}

type fakeLogger struct{}

// For debuging
func (l *fakeLogger) Debug(v ...interface{}) {}

// For debuging with format
func (l *fakeLogger) Debugf(format string, v ...interface{}) {}

// For logging info
func (l *fakeLogger) Info(v ...interface{}) {
	fmt.Println(v...)
}

// For logging info with format
func (l *fakeLogger) Infof(format string, v ...interface{}) {
	fmt.Printf(format+"\n", v...)
}

// For warning
func (l *fakeLogger) Warning(v ...interface{}) {}

// For warning with format
func (l *fakeLogger) Warningf(format string, v ...interface{}) {}

// For logging error
func (l *fakeLogger) Error(v ...interface{}) {
	fmt.Println(v...)
}

// For logging error with format
func (l *fakeLogger) Errorf(format string, v ...interface{}) {
}

// For fatal error
func (l *fakeLogger) Fatal(v ...interface{}) {}

// For fatal error with error
func (l *fakeLogger) Fatalf(format string, v ...interface{}) {}

type fakeJobContext struct{}

func (c *fakeJobContext) Build(tracker job.Tracker) (job.Context, error) {
	return nil, nil
}

func (c *fakeJobContext) Get(prop string) (interface{}, bool) {
	return nil, false
}

func (c *fakeJobContext) SystemContext() context.Context {
	return context.TODO()
}

func (c *fakeJobContext) Checkin(status string) error {
	fmt.Printf("Check in: %s\n", status)

	return nil
}

func (c *fakeJobContext) OPCommand() (job.OPCommand, bool) {
	return "", false
}

func (c *fakeJobContext) GetLogger() logger.Interface {
	return &fakeLogger{}
}

func (c *fakeJobContext) Tracker() job.Tracker {
	return nil
}
