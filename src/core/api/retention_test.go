package api

import (
	"encoding/json"
	"fmt"
	"github.com/goharbor/harbor/src/pkg/retention/dao"
	"github.com/goharbor/harbor/src/pkg/retention/dao/models"
	"github.com/goharbor/harbor/src/pkg/retention/policy"
	"github.com/goharbor/harbor/src/pkg/retention/policy/rule"
	"github.com/stretchr/testify/require"
	"net/http"
	"testing"
	"time"
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
	p1 := &policy.Metadata{
		Algorithm: "or",
		Rules: []rule.Metadata{
			{
				ID:       1,
				Priority: 1,
				Template: "recentXdays",
				Action:   "retain",
				Parameters: rule.Parameters{
					"num": 10,
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
			code: http.StatusOK,
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
							Template: "recentXdays",
							Action:   "retain",
							Parameters: rule.Parameters{
								"num": 10,
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
							Template: "recentXdays",
							Action:   "retain",
							Parameters: rule.Parameters{
								"num": 10,
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
							Template: "recentXdays",
							Action:   "retain",
							Parameters: rule.Parameters{
								"num": 10,
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
				Template: "recentXdays",
				Action:   "retain",
				Parameters: rule.Parameters{
					"num": 10,
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
							Template: "recentXdays",
							Action:   "retain",
							Parameters: rule.Parameters{
								"num": 10,
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
							Template: "recentXdays",
							Action:   "retain",
							Parameters: rule.Parameters{
								"num": 10,
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
							Template: "recentXdays",
							Action:   "retain",
							Parameters: rule.Parameters{
								"num": 10,
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
			code: http.StatusOK,
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
