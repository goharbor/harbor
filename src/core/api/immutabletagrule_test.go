package api

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/goharbor/harbor/src/common/models"
	"github.com/goharbor/harbor/src/pkg/immutabletag"
)

func TestImmutableTagRuleAPI_List(t *testing.T) {

	tagFilter := `{
    "id":0,
    "priority":0,
    "disabled":false,
    "action":"immutable",
    "template":"immutable_template",
    "tag_selectors":[{"kind":"doublestar","decoration":"matches","pattern":"**"}],
    "scope_selectors":{"repository":[{"kind":"doublestar","decoration":"repoMatches","pattern":"**"}]}
     }`

	mgr := immutabletag.NewDefaultRuleManager()
	id, err := mgr.CreateImmutableRule(&models.ImmutableRule{ProjectID: 1, TagFilter: tagFilter})
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
				var rules []models.ImmutableRule
				err := json.Unmarshal([]byte(responseRecorder.Body.String()), &rules)
				if err != nil {
					return err
				}
				if len(rules) <= 0 {
					return fmt.Errorf("no rules found")
				}
				if rules[0].TagFilter != tagFilter {
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

	tagFilter := `{
    "id":0,
    "priority":0,
    "disabled":false,
    "action":"immutable",
    "template":"immutable_template",
    "tag_selectors":[{"kind":"doublestar","decoration":"matches","pattern":"**"}],
    "scope_selectors":{"repository":[{"kind":"doublestar","decoration":"repoMatches","pattern":"**"}]}
}`
	body := &models.ImmutableRule{ProjectID: 1, TagFilter: tagFilter}
	cases := []*codeCheckingCase{
		// 401
		{
			request: &testingRequest{
				method:   http.MethodPost,
				url:      "/api/projects/1/immutabletagrules",
				bodyJSON: body,
			},
			code: http.StatusUnauthorized,
		},
		// 200
		{
			request: &testingRequest{
				method:     http.MethodPost,
				url:        "/api/projects/1/immutabletagrules",
				credential: admin,
				bodyJSON:   body,
			},
			code: http.StatusCreated,
		},
		// 200
		{
			request: &testingRequest{
				method:     http.MethodPost,
				url:        "/api/projects/1/immutabletagrules",
				credential: projAdmin,
				bodyJSON:   body,
			},
			code: http.StatusCreated,
		},
		// 403
		{
			request: &testingRequest{
				method:     http.MethodPost,
				url:        "/api/projects/1/immutabletagrules",
				credential: projGuest,
				bodyJSON:   body,
			},
			code: http.StatusForbidden,
		},
	}
	runCodeCheckingCases(t, cases...)

}

func TestImmutableTagRuleAPI_Put(t *testing.T) {
	tagFilter := `{
    "id":0,
    "priority":0,
    "disabled":false,
    "action":"immutable",
    "template":"immutable_template",
    "tag_selectors":[{"kind":"doublestar","decoration":"matches","pattern":"**"}],
    "scope_selectors":{"repository":[{"kind":"doublestar","decoration":"repoMatches","pattern":"**"}]}
}`
	tagFilter2 := `{
    "id":0,
    "priority":0,
    "disabled":false,
    "action":"immutable",
    "template":"immutable_template",
    "tag_selectors":[{"kind":"doublestar","decoration":"matches","pattern":"release-1.6.0"}],
    "scope_selectors":{"repository":[{"kind":"doublestar","decoration":"repoMatches","pattern":"regids"}]}
}`

	mgr := immutabletag.NewDefaultRuleManager()
	id, err := mgr.CreateImmutableRule(&models.ImmutableRule{ProjectID: 1, TagFilter: tagFilter})
	if err != nil {
		t.Error(err)
	}
	defer mgr.DeleteImmutableRule(id)
	url := fmt.Sprintf("/api/projects/1/immutabletagrules/%d", id)
	body := &models.ImmutableRule{ID: id, ProjectID: 1, TagFilter: tagFilter2}
	cases := []*codeCheckingCase{
		// 401
		{
			request: &testingRequest{
				method:   http.MethodPut,
				url:      url,
				bodyJSON: body,
			},
			code: http.StatusUnauthorized,
		},
		// 200
		{
			request: &testingRequest{
				method:     http.MethodPut,
				url:        url,
				credential: admin,
				bodyJSON:   body,
			},
			code: http.StatusOK,
		},
		// 200
		{
			request: &testingRequest{
				method:     http.MethodPut,
				url:        url,
				credential: projAdmin,
				bodyJSON:   body,
			},
			code: http.StatusOK,
		},
		// 403
		{
			request: &testingRequest{
				method:     http.MethodPut,
				url:        url,
				credential: projGuest,
				bodyJSON:   body,
			},
			code: http.StatusForbidden,
		},
	}
	runCodeCheckingCases(t, cases...)
}

func TestImmutableTagRuleAPI_Delete(t *testing.T) {
	tagFilter := `{
    "id":0,
    "priority":0,
    "disabled":false,
    "action":"immutable",
    "template":"immutable_template",
    "tag_selectors":[{"kind":"doublestar","decoration":"matches","pattern":"**"}],
    "scope_selectors":{"repository":[{"kind":"doublestar","decoration":"repoMatches","pattern":"**"}]}
}`
	mgr := immutabletag.NewDefaultRuleManager()
	id, err := mgr.CreateImmutableRule(&models.ImmutableRule{ProjectID: 1, TagFilter: tagFilter})
	if err != nil {
		t.Error(err)
	}
	defer mgr.DeleteImmutableRule(id)

	url := fmt.Sprintf("/api/projects/1/immutabletagrules/%d", id)

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
		// 200
		{
			request: &testingRequest{
				method:     http.MethodDelete,
				url:        url,
				credential: projAdmin,
			},
			code: http.StatusOK,
		},
	}
	runCodeCheckingCases(t, cases...)
}
