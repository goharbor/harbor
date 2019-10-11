// Copyright 2018 Project Harbor Authors
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
	"fmt"
	"net/http"
	"testing"

	"github.com/goharbor/harbor/src/common/models"
	"github.com/goharbor/harbor/src/common/rbac"
)

var (
	robotPath = "/api/projects/1/robots"
	robotID   int64
)

func TestRobotAPIPost(t *testing.T) {
	res := rbac.Resource("/project/1")

	rbacPolicy := &rbac.Policy{
		Resource: res.Subresource(rbac.ResourceRepository),
		Action:   "pull",
	}
	policies := []*rbac.Policy{}
	policies = append(policies, rbacPolicy)

	cases := []*codeCheckingCase{
		// 401
		{
			request: &testingRequest{
				method: http.MethodPost,
				url:    robotPath,
			},
			code: http.StatusUnauthorized,
		},

		// 403
		{
			request: &testingRequest{
				method:     http.MethodPost,
				url:        robotPath,
				bodyJSON:   &models.RobotReq{},
				credential: nonSysAdmin,
			},
			code: http.StatusForbidden,
		},
		// 201
		{
			request: &testingRequest{
				method: http.MethodPost,
				url:    robotPath,
				bodyJSON: &models.RobotReq{
					Name:        "test",
					Description: "test desc",
					Access:      policies,
				},
				credential: projAdmin4Robot,
			},
			code: http.StatusCreated,
		},
		// 400
		{
			request: &testingRequest{
				method: http.MethodPost,
				url:    robotPath,
				bodyJSON: &models.RobotReq{
					Name:        "testIllgel#",
					Description: "test desc",
				},
				credential: projAdmin4Robot,
			},
			code: http.StatusBadRequest,
		},
		{
			request: &testingRequest{
				method: http.MethodPost,
				url:    robotPath,
				bodyJSON: &models.RobotReq{
					Name:        "test",
					Description: "resource not exist",
					Access: []*rbac.Policy{
						{Resource: res.Subresource("foo"), Action: rbac.ActionCreate},
					},
				},
				credential: projAdmin4Robot,
			},
			code: http.StatusBadRequest,
		},
		{
			request: &testingRequest{
				method: http.MethodPost,
				url:    robotPath,
				bodyJSON: &models.RobotReq{
					Name:        "test",
					Description: "action not exist",
					Access: []*rbac.Policy{
						{Resource: res.Subresource(rbac.ResourceRepository), Action: "foo"},
					},
				},
				credential: projAdmin4Robot,
			},
			code: http.StatusBadRequest,
		},
		{
			request: &testingRequest{
				method: http.MethodPost,
				url:    robotPath,
				bodyJSON: &models.RobotReq{
					Name:        "test",
					Description: "policy not exit",
					Access: []*rbac.Policy{
						{Resource: res.Subresource(rbac.ResourceMember), Action: rbac.ActionPush},
					},
				},
				credential: projAdmin4Robot,
			},
			code: http.StatusBadRequest,
		},
		// 403 -- developer
		{
			request: &testingRequest{
				method: http.MethodPost,
				url:    robotPath,
				bodyJSON: &models.RobotReq{
					Name:        "test2",
					Description: "test2 desc",
				},
				credential: projDeveloper,
			},
			code: http.StatusForbidden,
		},

		// 409
		{
			request: &testingRequest{
				method: http.MethodPost,
				url:    robotPath,
				bodyJSON: &models.RobotReq{
					Name:        "test",
					Description: "test desc",
					Access:      policies,
				},
				credential: projAdmin4Robot,
			},
			code: http.StatusConflict,
		},
	}
	runCodeCheckingCases(t, cases...)
}

func TestRobotAPIGet(t *testing.T) {
	cases := []*codeCheckingCase{
		// 400
		{
			request: &testingRequest{
				method: http.MethodGet,
				url:    fmt.Sprintf("%s/%d", robotPath, 0),
			},
			code: http.StatusUnauthorized,
		},

		// 404
		{
			request: &testingRequest{
				method:     http.MethodGet,
				url:        fmt.Sprintf("%s/%d", robotPath, 1000),
				credential: projDeveloper,
			},
			code: http.StatusNotFound,
		},

		// 200
		{
			request: &testingRequest{
				method:     http.MethodGet,
				url:        fmt.Sprintf("%s/%d", robotPath, 1),
				credential: projDeveloper,
			},
			code: http.StatusOK,
		},

		// 200
		{
			request: &testingRequest{
				method:     http.MethodGet,
				url:        fmt.Sprintf("%s/%d", robotPath, 1),
				credential: projAdmin4Robot,
			},
			code: http.StatusOK,
		},
	}
	runCodeCheckingCases(t, cases...)
}

func TestRobotAPIList(t *testing.T) {
	cases := []*codeCheckingCase{
		// 401
		{
			request: &testingRequest{
				method: http.MethodGet,
				url:    robotPath,
			},
			code: http.StatusUnauthorized,
		},

		// 400
		{
			request: &testingRequest{
				method:     http.MethodGet,
				url:        "/api/projects/0/robots",
				credential: projAdmin4Robot,
			},
			code: http.StatusBadRequest,
		},

		// 200
		{
			request: &testingRequest{
				method:     http.MethodGet,
				url:        robotPath,
				credential: projDeveloper,
			},
			code: http.StatusOK,
		},

		// 200
		{
			request: &testingRequest{
				method:     http.MethodGet,
				url:        robotPath,
				credential: projAdmin4Robot,
			},
			code: http.StatusOK,
		},
	}
	runCodeCheckingCases(t, cases...)
}

func TestRobotAPIPut(t *testing.T) {
	cases := []*codeCheckingCase{
		// 401
		{
			request: &testingRequest{
				method: http.MethodPut,
				url:    fmt.Sprintf("%s/%d", robotPath, 1),
			},
			code: http.StatusUnauthorized,
		},

		// 400
		{
			request: &testingRequest{
				method:     http.MethodPut,
				url:        fmt.Sprintf("%s/%d", robotPath, 0),
				credential: projAdmin4Robot,
			},
			code: http.StatusBadRequest,
		},

		// 404
		{
			request: &testingRequest{
				method:     http.MethodPut,
				url:        fmt.Sprintf("%s/%d", robotPath, 10000),
				credential: projAdmin4Robot,
			},
			code: http.StatusNotFound,
		},

		// 403 non-member user
		{
			request: &testingRequest{
				method:     http.MethodPut,
				url:        fmt.Sprintf("%s/%d", robotPath, 1),
				credential: nonSysAdmin,
			},
			code: http.StatusForbidden,
		},

		// 403 developer
		{
			request: &testingRequest{
				method:     http.MethodPut,
				url:        fmt.Sprintf("%s/%d", robotPath, 1),
				credential: projDeveloper,
			},
			code: http.StatusForbidden,
		},

		// 200
		{
			request: &testingRequest{
				method: http.MethodPut,
				url:    fmt.Sprintf("%s/%d", robotPath, 1),
				bodyJSON: &models.Robot{
					Disabled: true,
				},
				credential: projAdmin4Robot,
			},
			code: http.StatusOK,
		},
	}

	runCodeCheckingCases(t, cases...)
}

func TestRobotAPIDelete(t *testing.T) {
	cases := []*codeCheckingCase{
		// 401
		{
			request: &testingRequest{
				method: http.MethodDelete,
				url:    fmt.Sprintf("%s/%d", robotPath, 1),
			},
			code: http.StatusUnauthorized,
		},

		// 400
		{
			request: &testingRequest{
				method:     http.MethodDelete,
				url:        fmt.Sprintf("%s/%d", robotPath, 0),
				credential: projAdmin4Robot,
			},
			code: http.StatusBadRequest,
		},

		// 404
		{
			request: &testingRequest{
				method:     http.MethodDelete,
				url:        fmt.Sprintf("%s/%d", robotPath, 10000),
				credential: projAdmin4Robot,
			},
			code: http.StatusNotFound,
		},

		// 403 non-member user
		{
			request: &testingRequest{
				method:     http.MethodDelete,
				url:        fmt.Sprintf("%s/%d", robotPath, 1),
				credential: nonSysAdmin,
			},
			code: http.StatusForbidden,
		},

		// 403 developer
		{
			request: &testingRequest{
				method:     http.MethodDelete,
				url:        fmt.Sprintf("%s/%d", robotPath, 1),
				credential: projDeveloper,
			},
			code: http.StatusForbidden,
		},

		// 200
		{
			request: &testingRequest{
				method:     http.MethodDelete,
				url:        fmt.Sprintf("%s/%d", robotPath, 1),
				credential: projAdmin4Robot,
			},
			code: http.StatusOK,
		},
	}

	runCodeCheckingCases(t, cases...)
}
