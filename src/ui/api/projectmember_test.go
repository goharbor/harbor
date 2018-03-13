// Copyright (c) 2017 VMware, Inc. All Rights Reserved.
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

	"github.com/vmware/harbor/src/common/dao"
	"github.com/vmware/harbor/src/common/dao/project"
	"github.com/vmware/harbor/src/common/models"
)

func TestProjectMemberAPI_Get(t *testing.T) {
	cases := []*codeCheckingCase{
		// 401
		&codeCheckingCase{
			request: &testingRequest{
				method: http.MethodGet,
				url:    "/api/projects/1/projectmembers",
			},
			code: http.StatusUnauthorized,
		},
		//200
		&codeCheckingCase{
			request: &testingRequest{
				method:     http.MethodGet,
				url:        "/api/projects/1/projectmembers",
				credential: admin,
			},
			code: http.StatusOK,
		},
		// 401
		&codeCheckingCase{
			request: &testingRequest{
				method:     http.MethodGet,
				url:        "/api/projects/1/projectmembers/1",
				credential: admin,
			},
			code: http.StatusOK,
		},
	}
	runCodeCheckingCases(t, cases...)
}

func TestProjectMemberAPI_Post(t *testing.T) {
	userID, err := dao.Register(models.User{
		Username: "restuser",
		Password: "Harbor12345",
		Email:    "restuser@example.com",
	})
	defer dao.DeleteUser(int(userID))
	if err != nil {
		t.Errorf("Error occurred when create user: %v", err)
	}

	cases := []*codeCheckingCase{
		// 401
		&codeCheckingCase{
			request: &testingRequest{
				method: http.MethodPost,
				url:    "/api/projects/1/projectmembers",
				bodyJSON: &models.MemberReq{
					Role:       1,
					EntityType: "u",
					EntityID:   int(userID),
				},
			},
			code: http.StatusUnauthorized,
		},
		&codeCheckingCase{
			request: &testingRequest{
				method: http.MethodPost,
				url:    "/api/projects/1/projectmembers",
				bodyJSON: &models.MemberReq{
					Role:       1,
					EntityType: "u",
					EntityID:   int(userID),
				},
				credential: admin,
			},
			code: http.StatusOK,
		},
		&codeCheckingCase{
			request: &testingRequest{
				method: http.MethodPost,
				url:    "/api/projects/1/projectmembers",
				bodyJSON: &models.MemberReq{
					Role:       1,
					EntityType: "u",
					EntityID:   0,
				},
				credential: admin,
			},
			code: http.StatusInternalServerError,
		},
		// &codeCheckingCase{
		// 	request: &testingRequest{
		// 		method: http.MethodPost,
		// 		url:    "/api/projects/1/projectmembers",
		// 		bodyJSON: &models.MemberReq{
		// 			Role:         1,
		// 			EntityType:   "u",
		// 			LdapUserName: "mike",
		// 		},
		// 		credential: admin,
		// 	},
		// 	code: http.StatusOK,
		// },
		// &codeCheckingCase{
		// 	request: &testingRequest{
		// 		method: http.MethodPost,
		// 		url:    "/api/projects/1/projectmembers",
		// 		bodyJSON: &models.MemberReq{
		// 			Role:        1,
		// 			EntityType:  "g",
		// 			LdapGroupDN: "cn=harbor_users,ou=groups,dc=example,dc=com",
		// 		},
		// 		credential: admin,
		// 	},
		// 	code: http.StatusOK,
		// },
	}
	runCodeCheckingCases(t, cases...)
}

func TestProjectMemberAPI_PutAndDelete(t *testing.T) {

	userID, err := dao.Register(models.User{
		Username: "restuser",
		Password: "Harbor12345",
		Email:    "restuser@example.com",
	})
	defer dao.DeleteUser(int(userID))
	if err != nil {
		t.Errorf("Error occurred when create user: %v", err)
	}

	ID, err := project.AddProjectMember(models.MemberReq{
		ProjectID:  1,
		Role:       1,
		EntityID:   int(userID),
		EntityType: "u",
	})
	if err != nil {
		t.Errorf("Error occurred when add project member: %v", err)
	}
	URL := fmt.Sprintf("/api/projects/1/projectmembers/%v", ID)
	badURL := fmt.Sprintf("/api/projects/1/projectmembers/%v", 0)
	cases := []*codeCheckingCase{
		// 401
		&codeCheckingCase{
			request: &testingRequest{
				method: http.MethodPut,
				url:    URL,
				bodyJSON: &models.Member{
					Role: 2,
				},
			},
			code: http.StatusUnauthorized,
		},
		// 200
		&codeCheckingCase{
			request: &testingRequest{
				method: http.MethodPut,
				url:    URL,
				bodyJSON: &models.Member{
					Role: 2,
				},
				credential: admin,
			},
			code: http.StatusOK,
		},
		// 500
		&codeCheckingCase{
			request: &testingRequest{
				method: http.MethodPut,
				url:    badURL,
				bodyJSON: &models.Member{
					Role: 2,
				},
				credential: admin,
			},
			code: http.StatusInternalServerError,
		},
		// 500
		&codeCheckingCase{
			request: &testingRequest{
				method: http.MethodPut,
				url:    URL,
				bodyJSON: &models.Member{
					Role: -2,
				},
				credential: admin,
			},
			code: http.StatusInternalServerError,
		},
		// 200
		&codeCheckingCase{
			request: &testingRequest{
				method:     http.MethodDelete,
				url:        URL,
				credential: admin,
			},
			code: http.StatusOK,
		},
	}

	runCodeCheckingCases(t, cases...)

}
