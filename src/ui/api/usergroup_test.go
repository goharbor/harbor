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

	"github.com/vmware/harbor/src/common/dao/group"

	"github.com/vmware/harbor/src/common/models"
)

const (
	URL = "/api/usergroups"
)

func TestUserGroupAPI_GetAndDelete(t *testing.T) {

	groupID, err := group.AddUserGroup(models.UserGroup{
		GroupName:     "harbor_users",
		GroupProperty: "cn=harbor_users,ou=groups,dc=example,dc=com",
		GroupType:     1,
	})

	if err != nil {
		t.Errorf("Error occurred when AddUserGroup: %v", err)
	}

	cases := []*codeCheckingCase{
		// 401
		&codeCheckingCase{
			request: &testingRequest{
				method: http.MethodGet,
				url:    URL,
			},
			code: http.StatusUnauthorized,
		},

		// 200
		&codeCheckingCase{
			request: &testingRequest{
				method:     http.MethodGet,
				url:        fmt.Sprintf("/api/usergroups/%d", groupID),
				credential: admin,
			},
			code: http.StatusOK,
		},
		// 200
		&codeCheckingCase{
			request: &testingRequest{
				method:     http.MethodGet,
				url:        fmt.Sprintf("/api/usergroups"),
				credential: admin,
			},
			code: http.StatusOK,
		},
		// 200
		&codeCheckingCase{
			request: &testingRequest{
				method:     http.MethodDelete,
				url:        fmt.Sprintf("/api/usergroups/%d", groupID),
				credential: admin,
			},
			code: http.StatusOK,
		},
	}

	runCodeCheckingCases(t, cases...)
}

func TestUserGroupAPI_Post(t *testing.T) {
	cases := []*codeCheckingCase{
		//200
		&codeCheckingCase{
			request: &testingRequest{
				method: http.MethodPost,
				url:    "/api/usergroups",
				bodyJSON: &models.UserGroup{
					GroupName:     "group_users",
					GroupProperty: "cn=group_users,ou=groups,dc=example,dc=com",
					GroupType:     1,
				},
				credential: admin,
			},
			code: http.StatusOK,
		},
	}
	runCodeCheckingCases(t, cases...)
}

func TestUserGroupAPI_Put(t *testing.T) {
	groupID, err := group.AddUserGroup(models.UserGroup{
		GroupName:     "harbor_group",
		GroupProperty: "cn=harbor_groups,ou=groups,dc=example,dc=com",
		GroupType:     1,
	})
	defer group.DeleteUserGroup(groupID)

	if err != nil {
		t.Errorf("Error occurred when AddUserGroup: %v", err)
	}
	cases := []*codeCheckingCase{
		//401
		&codeCheckingCase{
			request: &testingRequest{
				method: http.MethodPut,
				url:    fmt.Sprintf("/api/usergroups/%d", groupID),
				bodyJSON: &models.UserGroup{
					GroupName: "my_group",
				},
			},
			code: http.StatusUnauthorized,
		},
		//200
		&codeCheckingCase{
			request: &testingRequest{
				method: http.MethodPut,
				url:    fmt.Sprintf("/api/usergroups/%d", groupID),
				bodyJSON: &models.UserGroup{
					GroupName: "my_group",
				},
				credential: admin,
			},
			code: http.StatusOK,
		},
	}
	runCodeCheckingCases(t, cases...)
}
