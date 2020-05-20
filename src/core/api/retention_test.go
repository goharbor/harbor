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

package api

import (
	"encoding/json"
	"fmt"
	"net/http"
	"testing"
	"time"

	"github.com/goharbor/harbor/src/pkg/retention/dao"
	"github.com/goharbor/harbor/src/pkg/retention/dao/models"
	"github.com/goharbor/harbor/src/pkg/retention/mocks"
	"github.com/goharbor/harbor/src/pkg/retention/policy"
	"github.com/goharbor/harbor/src/pkg/retention/policy/rule"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestGetMetadatas(t *testing.T) {
	cases := []*codeCheckingCase{
		// 401
		{
			request: &testingRequest{
				method:     http.MethodGet,
				url:        "/api/retentions/metadatas",
				credential: sysAdmin,
			},
			code: http.StatusOK,
		},
	}

	runCodeCheckingCases(t, cases...)
}

func TestCreatePolicy(t *testing.T) {
	// mock retention api controller
	mockController := &mocks.APIController{}
	mockController.On("CreateRetention", mock.AnythingOfType("*policy.Metadata")).Return(int64(1), nil)

	controller := retentionController
	retentionController = mockController
	defer func() {
		retentionController = controller
	}()

	p1 := &policy.Metadata{
		Algorithm: "or",
		Rules: []rule.Metadata{
			{
				ID:       1,
				Priority: 1,
				Template: "latestPushedK",
				Action:   "retain",
				Parameters: rule.Parameters{
					"latestPushedK": 10,
				},
				TagSelectors: []*rule.Selector{
					{
						Kind:       "doublestar",
						Decoration: "matches",
						Pattern:    "release-[\\d\\.]+",
					},
				},
				ScopeSelectors: map[string][]*rule.Selector{
					"repository": {
						{
							Kind:       "doublestar",
							Decoration: "matches",
							Pattern:    ".+",
						},
					},
				},
			},
		},
		Trigger: &policy.Trigger{
			Kind: "Schedule",
			Settings: map[string]interface{}{
				"cron": "* 22 11 * * *",
			},
		},
		Scope: &policy.Scope{
			Level:     "project",
			Reference: 1,
		},
	}

	cases := []*codeCheckingCase{
		// 401
		{
			request: &testingRequest{
				method: http.MethodPost,
				url:    "/api/retentions",
			},
			code: http.StatusUnauthorized,
		},
		{
			request: &testingRequest{
				method:     http.MethodPost,
				url:        "/api/retentions",
				bodyJSON:   p1,
				credential: sysAdmin,
			},
			code: http.StatusCreated,
		},
		{
			request: &testingRequest{
				method: http.MethodPost,
				url:    "/api/retentions",
				bodyJSON: &policy.Metadata{
					Algorithm: "NODEF",
					Rules: []rule.Metadata{
						{
							ID:       1,
							Priority: 1,
							Template: "latestPushedK",
							Action:   "retain",
							Parameters: rule.Parameters{
								"latestPushedK": 10,
							},
							TagSelectors: []*rule.Selector{
								{
									Kind:       "doublestar",
									Decoration: "matches",
									Pattern:    "release-[\\d\\.]+",
								},
							},
							ScopeSelectors: map[string][]*rule.Selector{
								"repository": {
									{
										Kind:       "doublestar",
										Decoration: "matches",
										Pattern:    ".+",
									},
								},
							},
						},
					},
					Trigger: &policy.Trigger{
						Kind:     "Schedule",
						Settings: map[string]interface{}{},
					},
					Scope: &policy.Scope{
						Level:     "project",
						Reference: 1,
					},
				},
				credential: sysAdmin,
			},
			code: http.StatusBadRequest,
		},
		{
			request: &testingRequest{
				method: http.MethodPost,
				url:    "/api/retentions",
				bodyJSON: &policy.Metadata{
					Algorithm: "or",
					Rules: []rule.Metadata{
						{
							ID:       1,
							Priority: 1,
							Template: "latestPushedK",
							Action:   "retain",
							Parameters: rule.Parameters{
								"latestPushedK": 10,
							},
							TagSelectors: []*rule.Selector{
								{
									Kind:       "doublestar",
									Decoration: "matches",
									Pattern:    "release-[\\d\\.]+",
								},
							},
							ScopeSelectors: map[string][]*rule.Selector{
								"repository": {
									{
										Kind:       "doublestar",
										Decoration: "matches",
										Pattern:    ".+",
									},
								},
							},
						},
						{
							ID:       2,
							Priority: 1,
							Template: "latestPushedK",
							Action:   "retain",
							Parameters: rule.Parameters{
								"latestPushedK": 10,
							},
							TagSelectors: []*rule.Selector{
								{
									Kind:       "doublestar",
									Decoration: "matches",
									Pattern:    "release-[\\d\\.]+",
								},
							},
							ScopeSelectors: map[string][]*rule.Selector{
								"repository": {
									{
										Kind:       "doublestar",
										Decoration: "matches",
										Pattern:    ".+",
									},
								},
							},
						},
					},
					Trigger: &policy.Trigger{
						Kind: "Schedule",
						Settings: map[string]interface{}{
							"cron": "* 22 11 * * *",
						},
					},
					Scope: &policy.Scope{
						Level:     "project",
						Reference: 1,
					},
				},
				credential: sysAdmin,
			},
			code: http.StatusConflict,
		},
	}

	runCodeCheckingCases(t, cases...)
}

func TestPolicy(t *testing.T) {
	p := &policy.Metadata{
		Algorithm: "or",
		Rules: []rule.Metadata{
			{
				ID:       1,
				Priority: 1,
				Template: "latestPushedK",
				Action:   "retain",
				Parameters: rule.Parameters{
					"latestPushedK": 10,
				},
				TagSelectors: []*rule.Selector{
					{
						Kind:       "doublestar",
						Decoration: "matches",
						Pattern:    "release-[\\d\\.]+",
					},
				},
				ScopeSelectors: map[string][]*rule.Selector{
					"repository": {
						{
							Kind:       "doublestar",
							Decoration: "matches",
							Pattern:    ".+",
						},
					},
				},
			},
		},
		Trigger: &policy.Trigger{
			Kind: "Schedule",
			Settings: map[string]interface{}{
				"cron": "* 22 11 * * *",
			},
		},
		Scope: &policy.Scope{
			Level:     "project",
			Reference: 1,
		},
	}
	p1 := &models.RetentionPolicy{
		ScopeLevel:  p.Scope.Level,
		TriggerKind: p.Trigger.Kind,
		CreateTime:  time.Now(),
		UpdateTime:  time.Now(),
	}
	data, _ := json.Marshal(p)
	p1.Data = string(data)

	id, err := dao.CreatePolicy(p1)
	require.Nil(t, err)
	require.True(t, id > 0)

	// mock retention api controller
	mockController := &mocks.APIController{}
	mockController.On("GetRetention", mock.AnythingOfType("int64")).Return(p, nil)
	mockController.On("UpdateRetention", mock.AnythingOfType("*policy.Metadata")).Return(nil)
	mockController.On("TriggerRetentionExec",
		mock.AnythingOfType("int64"),
		mock.AnythingOfType("string"),
		mock.AnythingOfType("bool")).Return(int64(1), nil)
	mockController.On("ListRetentionExecs", mock.AnythingOfType("int64"), mock.AnythingOfType("*q.Query")).Return(nil, nil)
	mockController.On("GetTotalOfRetentionExecs", mock.AnythingOfType("int64")).Return(int64(0), nil)

	controller := retentionController
	retentionController = mockController
	defer func() {
		retentionController = controller
	}()

	cases := []*codeCheckingCase{
		{
			request: &testingRequest{
				method:     http.MethodGet,
				url:        fmt.Sprintf("/api/retentions/%d", id),
				credential: sysAdmin,
			},
			code: http.StatusOK,
		},
		{
			request: &testingRequest{
				method: http.MethodPut,
				url:    fmt.Sprintf("/api/retentions/%d", id),
				bodyJSON: &policy.Metadata{
					Algorithm: "or",
					Rules: []rule.Metadata{
						{
							ID:       1,
							Priority: 1,
							Template: "latestPushedK",
							Action:   "retain",
							Parameters: rule.Parameters{
								"latestPushedK": 10,
							},
							TagSelectors: []*rule.Selector{
								{
									Kind:       "doublestar",
									Decoration: "matches",
									Pattern:    "release-[\\d\\.]+",
								},
							},
							ScopeSelectors: map[string][]*rule.Selector{
								"repository": {
									{
										Kind:       "doublestar",
										Decoration: "matches",
										Pattern:    "b.+",
									},
								},
							},
						},
					},
					Trigger: &policy.Trigger{
						Kind: "Schedule",
						Settings: map[string]interface{}{
							"cron": "* 22 11 * * *",
						},
					},
					Scope: &policy.Scope{
						Level:     "project",
						Reference: 1,
					},
				},
				credential: sysAdmin,
			},
			code: http.StatusOK,
		},
		{
			request: &testingRequest{
				method: http.MethodPut,
				url:    fmt.Sprintf("/api/retentions/%d", id),
				bodyJSON: &policy.Metadata{
					Algorithm: "or",
					Rules: []rule.Metadata{
						{
							ID:       1,
							Priority: 1,
							Template: "latestPushedK",
							Action:   "retain",
							Parameters: rule.Parameters{
								"latestPushedK": 10,
							},
							TagSelectors: []*rule.Selector{
								{
									Kind:       "doublestar",
									Decoration: "matches",
									Pattern:    "release-[\\d\\.]+",
								},
							},
							ScopeSelectors: map[string][]*rule.Selector{
								"repository": {
									{
										Kind:       "doublestar",
										Decoration: "matches",
										Pattern:    "b.+",
									},
								},
							},
						},
						{
							ID:       2,
							Priority: 1,
							Template: "latestPushedK",
							Action:   "retain",
							Parameters: rule.Parameters{
								"latestPushedK": 10,
							},
							TagSelectors: []*rule.Selector{
								{
									Kind:       "doublestar",
									Decoration: "matches",
									Pattern:    "release-[\\d\\.]+",
								},
							},
							ScopeSelectors: map[string][]*rule.Selector{
								"repository": {
									{
										Kind:       "doublestar",
										Decoration: "matches",
										Pattern:    "b.+",
									},
								},
							},
						},
					},
					Trigger: &policy.Trigger{
						Kind: "Schedule",
						Settings: map[string]interface{}{
							"cron": "* 22 11 * * *",
						},
					},
					Scope: &policy.Scope{
						Level:     "project",
						Reference: 1,
					},
				},
				credential: sysAdmin,
			},
			code: http.StatusConflict,
		},
		{
			request: &testingRequest{
				method: http.MethodPost,
				url:    fmt.Sprintf("/api/retentions/%d/executions", id),
				bodyJSON: &struct {
					DryRun bool `json:"dry_run"`
				}{
					DryRun: false,
				},
				credential: sysAdmin,
			},
			code: http.StatusCreated,
		},
		{
			request: &testingRequest{
				method:     http.MethodGet,
				url:        fmt.Sprintf("/api/retentions/%d/executions", id),
				credential: sysAdmin,
			},
			code: http.StatusOK,
		},
	}

	runCodeCheckingCases(t, cases...)
}
