package api

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/goharbor/harbor/src/pkg/immutabletag"
	"github.com/goharbor/harbor/src/pkg/immutabletag/model"
)

func TestImmutableTagRuleAPI_List(t *testing.T) {

	metadata := &model.Metadata{
		ProjectID: 1,
		Disabled:  false,
		TagSelectors: []*model.Selector{
			{
				Kind:       "doublestar",
				Decoration: "matches",
				Pattern:    "release-[\\d\\.]+",
			},
		},
		ScopeSelectors: map[string][]*model.Selector{
			"repository": {
				{
					Kind:       "doublestar",
					Decoration: "matches",
					Pattern:    ".+",
				},
			},
		},
	}
	mgr := immutabletag.NewDefaultRuleManager()
	id, err := mgr.CreateImmutableRule(metadata)
	if err != nil {
		t.Error(err)
	}
	defer mgr.DeleteImmutableRule(id)
	cases := []*codeCheckingCase{
		// 401
		{
			request: &testingRequest{
				method: http.MethodGet,
				url:    "/api/projects/1/immutabletagrules",
			},
			code: http.StatusUnauthorized,
		},
		// 200
		{
			request: &testingRequest{
				method:     http.MethodGet,
				url:        "/api/projects/1/immutabletagrules",
				credential: admin,
			},
			postFunc: func(responseRecorder *httptest.ResponseRecorder) error {
				var rules []model.Metadata
				err := json.Unmarshal([]byte(responseRecorder.Body.String()), &rules)
				if err != nil {
					return err
				}
				if len(rules) <= 0 {
					return fmt.Errorf("no rules found")
				}
				if rules[0].TagSelectors[0].Kind != "doublestar" {
					return fmt.Errorf("rule is not expected. actual: %v", responseRecorder.Body.String())
				}
				return nil
			},
			code: http.StatusOK,
		},
		// 200
		{
			request: &testingRequest{
				method:     http.MethodGet,
				url:        "/api/projects/1/immutabletagrules",
				credential: projAdmin,
			},
			code: http.StatusOK,
		},
		// 403
		{
			request: &testingRequest{
				method:     http.MethodGet,
				url:        "/api/projects/1/immutabletagrules",
				credential: projGuest,
			},
			code: http.StatusForbidden,
		},
	}
	runCodeCheckingCases(t, cases...)

}

func TestImmutableTagRuleAPI_Post(t *testing.T) {

	//	body := `{
	//	  "projectID":1,
	//    "priority":0,
	//    "template": "immutable_template",
	//    "action": "immutable",
	//    "disabled":false,
	//    "action":"immutable",
	//    "template":"immutable_template",
	//    "tag_selectors":[{"kind":"doublestar","decoration":"matches","pattern":"**"}],
	//    "scope_selectors":{"repository":[{"kind":"doublestar","decoration":"repoMatches","pattern":"**"}]}
	//  }`

	metadata := &model.Metadata{
		ProjectID: 1,
		Disabled:  false,
		Priority:  0,
		Template:  "immutable_template",
		Action:    "immutable",
		TagSelectors: []*model.Selector{
			{
				Kind:       "doublestar",
				Decoration: "matches",
				Pattern:    "release-[\\d\\.]+",
			},
		},
		ScopeSelectors: map[string][]*model.Selector{
			"repository": {
				{
					Kind:       "doublestar",
					Decoration: "matches",
					Pattern:    ".+",
				},
			},
		},
	}

	cases := []*codeCheckingCase{
		// 401
		{
			request: &testingRequest{
				method:   http.MethodPost,
				url:      "/api/projects/1/immutabletagrules",
				bodyJSON: metadata,
			},
			code: http.StatusUnauthorized,
		},
		// 201
		{
			request: &testingRequest{
				method:     http.MethodPost,
				url:        "/api/projects/1/immutabletagrules",
				credential: admin,
				bodyJSON:   metadata,
			},
			code: http.StatusCreated,
		},
		// 201
		{
			request: &testingRequest{
				method:     http.MethodPost,
				url:        "/api/projects/1/immutabletagrules",
				credential: projAdmin,
				bodyJSON:   metadata,
			},
			code: http.StatusCreated,
		},
		// 403
		{
			request: &testingRequest{
				method:     http.MethodPost,
				url:        "/api/projects/1/immutabletagrules",
				credential: projGuest,
				bodyJSON:   metadata,
			},
			code: http.StatusForbidden,
		},
	}
	runCodeCheckingCases(t, cases...)

}

func TestImmutableTagRuleAPI_Put(t *testing.T) {

	metadata := &model.Metadata{
		ProjectID: 1,
		Disabled:  false,
		TagSelectors: []*model.Selector{
			{
				Kind:       "doublestar",
				Decoration: "matches",
				Pattern:    "release-[\\d\\.]+",
			},
		},
		ScopeSelectors: map[string][]*model.Selector{
			"repository": {
				{
					Kind:       "doublestar",
					Decoration: "matches",
					Pattern:    ".+",
				},
			},
		},
	}

	metadata2 := &model.Metadata{
		ProjectID: 1,
		Disabled:  false,
		TagSelectors: []*model.Selector{
			{
				Kind:       "doublestar",
				Decoration: "matches",
				Pattern:    "latest",
			},
		},
		ScopeSelectors: map[string][]*model.Selector{
			"repository": {
				{
					Kind:       "doublestar",
					Decoration: "matches",
					Pattern:    ".+",
				},
			},
		},
	}
	mgr := immutabletag.NewDefaultRuleManager()
	id, err := mgr.CreateImmutableRule(metadata)
	if err != nil {
		t.Error(err)
	}
	defer mgr.DeleteImmutableRule(id)

	url := fmt.Sprintf("/api/projects/1/immutabletagrules/%d", id)
	url2 := fmt.Sprintf("/api/projects/3/immutabletagrules/%d", id)
	cases := []*codeCheckingCase{
		// 401
		{
			request: &testingRequest{
				method:   http.MethodPut,
				url:      url,
				bodyJSON: metadata2,
			},
			code: http.StatusUnauthorized,
		},
		// 200
		{
			request: &testingRequest{
				method:     http.MethodPut,
				url:        url,
				credential: admin,
				bodyJSON:   metadata2,
			},
			code: http.StatusOK,
		},
		// 200
		{
			request: &testingRequest{
				method:     http.MethodPut,
				url:        url,
				credential: projAdmin,
				bodyJSON:   metadata2,
			},
			code: http.StatusOK,
		},
		// 403
		{
			request: &testingRequest{
				method:     http.MethodPut,
				url:        url,
				credential: projGuest,
				bodyJSON:   metadata2,
			},
			code: http.StatusForbidden,
		},
		// 404
		{
			request: &testingRequest{
				method:     http.MethodPut,
				url:        url2,
				credential: projAdmin,
				bodyJSON:   metadata2,
			},
			code: http.StatusNotFound,
		},
	}
	runCodeCheckingCases(t, cases...)
}

func TestImmutableTagRuleAPI_Delete(t *testing.T) {
	metadata := &model.Metadata{
		ProjectID: 1,
		Disabled:  false,
		TagSelectors: []*model.Selector{
			{
				Kind:       "doublestar",
				Decoration: "matches",
				Pattern:    "latest",
			},
		},
		ScopeSelectors: map[string][]*model.Selector{
			"repository": {
				{
					Kind:       "doublestar",
					Decoration: "matches",
					Pattern:    ".+",
				},
			},
		},
	}

	mgr := immutabletag.NewDefaultRuleManager()
	id, err := mgr.CreateImmutableRule(metadata)
	if err != nil {
		t.Error(err)
	}
	defer mgr.DeleteImmutableRule(id)

	url := fmt.Sprintf("/api/projects/1/immutabletagrules/%d", id)
	wrongURL := fmt.Sprintf("/api/projects/3/immutabletagrules/%d", id)

	cases := []*codeCheckingCase{
		// 401
		{
			request: &testingRequest{
				method: http.MethodDelete,
				url:    url,
			},
			code: http.StatusUnauthorized,
		},
		// 403
		{
			request: &testingRequest{
				method:     http.MethodDelete,
				url:        url,
				credential: projGuest,
			},
			code: http.StatusForbidden,
		},
		// 404
		{
			request: &testingRequest{
				method:     http.MethodDelete,
				url:        wrongURL,
				credential: projAdmin,
			},
			code: http.StatusNotFound,
		},
		// 200
		{
			request: &testingRequest{
				method:     http.MethodDelete,
				url:        url,
				credential: projAdmin,
			},
			code: http.StatusOK,
		},
		// 404
		{
			request: &testingRequest{
				method:     http.MethodDelete,
				url:        url,
				credential: projAdmin,
			},
			code: http.StatusNotFound,
		},
	}
	runCodeCheckingCases(t, cases...)
}
