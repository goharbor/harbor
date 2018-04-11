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
	"net/http/httptest"
	"testing"

	"github.com/vmware/harbor/src/common/dao/project"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/vmware/harbor/src/common"
	"github.com/vmware/harbor/src/common/dao"
	"github.com/vmware/harbor/src/common/models"
	"github.com/vmware/harbor/src/replication"
	rep_models "github.com/vmware/harbor/src/replication/models"
	api_models "github.com/vmware/harbor/src/ui/api/models"
)

var (
	repPolicyAPIBasePath       = "/api/policies/replication"
	policyName                 = "testPolicy"
	projectID            int64 = 1
	targetID             int64
	policyID             int64
)

func TestRepPolicyAPIPost(t *testing.T) {
	postFunc := func(resp *httptest.ResponseRecorder) error {
		id, err := parseResourceID(resp)
		if err != nil {
			return err
		}
		policyID = id
		return nil
	}

	CommonAddTarget()
	targetID = int64(CommonGetTarget())

	cases := []*codeCheckingCase{
		// 401
		&codeCheckingCase{
			request: &testingRequest{
				method: http.MethodPost,
				url:    repPolicyAPIBasePath,
			},
			code: http.StatusUnauthorized,
		},
		// 403
		&codeCheckingCase{
			request: &testingRequest{
				method:     http.MethodPost,
				url:        repPolicyAPIBasePath,
				credential: nonSysAdmin,
			},
			code: http.StatusForbidden,
		},

		// 400, invalid name
		&codeCheckingCase{
			request: &testingRequest{
				method:     http.MethodPost,
				url:        repPolicyAPIBasePath,
				bodyJSON:   &api_models.ReplicationPolicy{},
				credential: sysAdmin,
			},
			code: http.StatusBadRequest,
		},
		// 400, invalid projects
		&codeCheckingCase{
			request: &testingRequest{
				method: http.MethodPost,
				url:    repPolicyAPIBasePath,
				bodyJSON: &api_models.ReplicationPolicy{
					Name: policyName,
				},
				credential: sysAdmin,
			},
			code: http.StatusBadRequest,
		},
		// 400, invalid targets
		&codeCheckingCase{
			request: &testingRequest{
				method: http.MethodPost,
				url:    repPolicyAPIBasePath,
				bodyJSON: &api_models.ReplicationPolicy{
					Name: policyName,
					Projects: []*models.Project{
						&models.Project{
							ProjectID: projectID,
						},
					},
				},
				credential: sysAdmin,
			},
			code: http.StatusBadRequest,
		},
		// 400, invalid filters
		&codeCheckingCase{
			request: &testingRequest{
				method: http.MethodPost,
				url:    repPolicyAPIBasePath,
				bodyJSON: &api_models.ReplicationPolicy{
					Name: policyName,
					Projects: []*models.Project{
						&models.Project{
							ProjectID: projectID,
						},
					},
					Targets: []*models.RepTarget{
						&models.RepTarget{
							ID: targetID,
						},
					},
					Filters: []rep_models.Filter{
						rep_models.Filter{
							Kind:    "invalid_filter_kind",
							Pattern: "",
						},
					},
				},
				credential: sysAdmin,
			},
			code: http.StatusBadRequest,
		},
		// 400, invalid trigger
		&codeCheckingCase{
			request: &testingRequest{
				method: http.MethodPost,
				url:    repPolicyAPIBasePath,
				bodyJSON: &api_models.ReplicationPolicy{
					Name: policyName,
					Projects: []*models.Project{
						&models.Project{
							ProjectID: projectID,
						},
					},
					Targets: []*models.RepTarget{
						&models.RepTarget{
							ID: targetID,
						},
					},
					Filters: []rep_models.Filter{
						rep_models.Filter{
							Kind:    replication.FilterItemKindRepository,
							Pattern: "*",
						},
					},
					Trigger: &rep_models.Trigger{
						Kind: "invalid_trigger_kind",
					},
				},
				credential: sysAdmin,
			},
			code: http.StatusBadRequest,
		},
		// 404, project not found
		&codeCheckingCase{
			request: &testingRequest{
				method: http.MethodPost,
				url:    repPolicyAPIBasePath,
				bodyJSON: &api_models.ReplicationPolicy{
					Name: policyName,
					Projects: []*models.Project{
						&models.Project{
							ProjectID: 10000,
						},
					},
					Targets: []*models.RepTarget{
						&models.RepTarget{
							ID: targetID,
						},
					},
					Filters: []rep_models.Filter{
						rep_models.Filter{
							Kind:    replication.FilterItemKindRepository,
							Pattern: "*",
						},
					},
					Trigger: &rep_models.Trigger{
						Kind: replication.TriggerKindManual,
					},
				},
				credential: sysAdmin,
			},
			code: http.StatusNotFound,
		},
		// 404, target not found
		&codeCheckingCase{
			request: &testingRequest{
				method: http.MethodPost,
				url:    repPolicyAPIBasePath,
				bodyJSON: &api_models.ReplicationPolicy{
					Name: policyName,
					Projects: []*models.Project{
						&models.Project{
							ProjectID: projectID,
						},
					},
					Targets: []*models.RepTarget{
						&models.RepTarget{
							ID: 10000,
						},
					},
					Filters: []rep_models.Filter{
						rep_models.Filter{
							Kind:    replication.FilterItemKindRepository,
							Pattern: "*",
						},
					},
					Trigger: &rep_models.Trigger{
						Kind: replication.TriggerKindManual,
					},
				},
				credential: sysAdmin,
			},
			code: http.StatusNotFound,
		},
		// 201
		&codeCheckingCase{
			request: &testingRequest{
				method: http.MethodPost,
				url:    repPolicyAPIBasePath,
				bodyJSON: &api_models.ReplicationPolicy{
					Name: policyName,
					Projects: []*models.Project{
						&models.Project{
							ProjectID: projectID,
						},
					},
					Targets: []*models.RepTarget{
						&models.RepTarget{
							ID: targetID,
						},
					},
					Filters: []rep_models.Filter{
						rep_models.Filter{
							Kind:    replication.FilterItemKindRepository,
							Pattern: "*",
						},
					},
					Trigger: &rep_models.Trigger{
						Kind: replication.TriggerKindManual,
					},
				},
				credential: sysAdmin,
			},
			code:     http.StatusCreated,
			postFunc: postFunc,
		},
	}

	runCodeCheckingCases(t, cases...)
}

func TestRepPolicyAPIGet(t *testing.T) {

	cases := []*codeCheckingCase{
		// 404
		&codeCheckingCase{
			request: &testingRequest{
				method:     http.MethodGet,
				url:        fmt.Sprintf("%s/%d", repPolicyAPIBasePath, 10000),
				credential: sysAdmin,
			},
			code: http.StatusNotFound,
		},
		// 401
		&codeCheckingCase{
			request: &testingRequest{
				method: http.MethodGet,
				url:    fmt.Sprintf("%s/%d", repPolicyAPIBasePath, policyID),
			},
			code: http.StatusUnauthorized,
		},
	}

	runCodeCheckingCases(t, cases...)

	// 200
	policy := &api_models.ReplicationPolicy{}
	err := handleAndParse(
		&testingRequest{
			method:     http.MethodGet,
			url:        fmt.Sprintf("%s/%d", repPolicyAPIBasePath, policyID),
			credential: sysAdmin,
		}, policy)
	require.Nil(t, err)
	assert.Equal(t, policyID, policy.ID)
	assert.Equal(t, policyName, policy.Name)
}

func TestRepPolicyAPIList(t *testing.T) {
	projectAdmin := models.User{
		Username: "project_admin",
		Password: "ProjectAdmin",
		Email:    "project_admin@test.com",
	}
	projectDev := models.User{
		Username: "project_dev",
		Password: "ProjectDev",
		Email:    "project_dev@test.com",
	}
	var proAdminPMID, proDevPMID int
	proAdminID, err := dao.Register(projectAdmin)
	if err != nil {
		panic(err)
	}
	defer dao.DeleteUser(int(proAdminID))
	if proAdminPMID, err = project.AddProjectMember(models.Member{
		ProjectID:  1,
		Role:       models.PROJECTADMIN,
		EntityID:   int(proAdminID),
		EntityType: common.UserMember,
	}); err != nil {
		panic(err)
	}
	defer project.DeleteProjectMemberByID(proAdminPMID)

	proDevID, err := dao.Register(projectDev)
	if err != nil {
		panic(err)
	}
	defer dao.DeleteUser(int(proDevID))

	if proDevPMID, err = project.AddProjectMember(models.Member{
		ProjectID:  1,
		Role:       models.DEVELOPER,
		EntityID:   int(proDevID),
		EntityType: common.UserMember,
	}); err != nil {
		panic(err)
	}
	defer project.DeleteProjectMemberByID(proDevPMID)

	// 400: invalid project ID
	runCodeCheckingCases(t, &codeCheckingCase{
		request: &testingRequest{
			method: http.MethodGet,
			url:    repPolicyAPIBasePath,
			queryStruct: struct {
				ProjectID int64 `url:"project_id"`
			}{
				ProjectID: -1,
			},
			credential: sysAdmin,
		},
		code: http.StatusBadRequest,
	})

	// 200 system admin
	policies := []*api_models.ReplicationPolicy{}
	err = handleAndParse(
		&testingRequest{
			method: http.MethodGet,
			url:    repPolicyAPIBasePath,
			queryStruct: struct {
				ProjectID int64  `url:"project_id"`
				Name      string `url:"name"`
			}{
				ProjectID: projectID,
				Name:      policyName,
			},
			credential: sysAdmin,
		}, &policies)
	require.Nil(t, err)
	require.Equal(t, 1, len(policies))
	assert.Equal(t, policyID, policies[0].ID)
	assert.Equal(t, policyName, policies[0].Name)

	// 200 project admin
	policies = []*api_models.ReplicationPolicy{}
	err = handleAndParse(
		&testingRequest{
			method: http.MethodGet,
			url:    repPolicyAPIBasePath,
			queryStruct: struct {
				ProjectID int64  `url:"project_id"`
				Name      string `url:"name"`
			}{
				ProjectID: projectID,
				Name:      policyName,
			},
			credential: &usrInfo{
				Name:   projectAdmin.Username,
				Passwd: projectAdmin.Password,
			},
		}, &policies)
	require.Nil(t, err)
	require.Equal(t, 1, len(policies))
	assert.Equal(t, policyID, policies[0].ID)
	assert.Equal(t, policyName, policies[0].Name)

	// 200 project developer
	policies = []*api_models.ReplicationPolicy{}
	err = handleAndParse(
		&testingRequest{
			method: http.MethodGet,
			url:    repPolicyAPIBasePath,
			queryStruct: struct {
				ProjectID int64  `url:"project_id"`
				Name      string `url:"name"`
			}{
				ProjectID: projectID,
				Name:      policyName,
			},
			credential: &usrInfo{
				Name:   projectDev.Username,
				Passwd: projectDev.Password,
			},
		}, &policies)
	require.Nil(t, err)
	require.Equal(t, 0, len(policies))

	// 200
	policies = []*api_models.ReplicationPolicy{}
	err = handleAndParse(
		&testingRequest{
			method: http.MethodGet,
			url:    repPolicyAPIBasePath,
			queryStruct: struct {
				ProjectID int64  `url:"project_id"`
				Name      string `url:"name"`
			}{
				ProjectID: projectID,
				Name:      "non_exist_policy",
			},
			credential: sysAdmin,
		}, &policies)
	require.Nil(t, err)
	require.Equal(t, 0, len(policies))
}

func TestRepPolicyAPIPut(t *testing.T) {
	cases := []*codeCheckingCase{
		// 404
		&codeCheckingCase{
			request: &testingRequest{
				method:     http.MethodPut,
				url:        fmt.Sprintf("%s/%d", repPolicyAPIBasePath, 10000),
				credential: sysAdmin,
			},
			code: http.StatusNotFound,
		},
		// 400, invalid trigger
		&codeCheckingCase{
			request: &testingRequest{
				method: http.MethodPut,
				url:    fmt.Sprintf("%s/%d", repPolicyAPIBasePath, policyID),
				bodyJSON: &api_models.ReplicationPolicy{
					Name: policyName,
					Projects: []*models.Project{
						&models.Project{
							ProjectID: projectID,
						},
					},
					Targets: []*models.RepTarget{
						&models.RepTarget{
							ID: targetID,
						},
					},
					Filters: []rep_models.Filter{
						rep_models.Filter{
							Kind:    replication.FilterItemKindRepository,
							Pattern: "*",
						},
					},
					Trigger: &rep_models.Trigger{
						Kind: "invalid_trigger_kind",
					},
				},
				credential: sysAdmin,
			},
			code: http.StatusBadRequest,
		},
		// 200
		&codeCheckingCase{
			request: &testingRequest{
				method: http.MethodPut,
				url:    fmt.Sprintf("%s/%d", repPolicyAPIBasePath, policyID),
				bodyJSON: &api_models.ReplicationPolicy{
					Name: policyName,
					Projects: []*models.Project{
						&models.Project{
							ProjectID: projectID,
						},
					},
					Targets: []*models.RepTarget{
						&models.RepTarget{
							ID: targetID,
						},
					},
					Filters: []rep_models.Filter{
						rep_models.Filter{
							Kind:    replication.FilterItemKindRepository,
							Pattern: "*",
						},
					},
					Trigger: &rep_models.Trigger{
						Kind: replication.TriggerKindImmediate,
					},
				},
				credential: sysAdmin,
			},
			code: http.StatusOK,
		},
	}

	runCodeCheckingCases(t, cases...)
}

func TestRepPolicyAPIDelete(t *testing.T) {
	cases := []*codeCheckingCase{
		// 404
		&codeCheckingCase{
			request: &testingRequest{
				method:     http.MethodDelete,
				url:        fmt.Sprintf("%s/%d", repPolicyAPIBasePath, 10000),
				credential: sysAdmin,
			},
			code: http.StatusNotFound,
		},
		// 200
		&codeCheckingCase{
			request: &testingRequest{
				method:     http.MethodDelete,
				url:        fmt.Sprintf("%s/%d", repPolicyAPIBasePath, policyID),
				credential: sysAdmin,
			},
			code: http.StatusOK,
		},
	}

	runCodeCheckingCases(t, cases...)
}

func TestConvertToRepPolicy(t *testing.T) {
	cases := []struct {
		input    *api_models.ReplicationPolicy
		expected rep_models.ReplicationPolicy
	}{
		{
			input:    nil,
			expected: rep_models.ReplicationPolicy{},
		},
		{
			input: &api_models.ReplicationPolicy{
				ID:          1,
				Name:        "policy",
				Description: "description",
				Filters: []rep_models.Filter{
					rep_models.Filter{
						Kind:    "filter_kind_01",
						Pattern: "*",
					},
				},
				ReplicateDeletion: true,
				Trigger: &rep_models.Trigger{
					Kind: "trigger_kind_01",
				},
				Projects: []*models.Project{
					&models.Project{
						ProjectID: 1,
						Name:      "library",
					},
				},
				Targets: []*models.RepTarget{
					&models.RepTarget{
						ID: 1,
					},
				},
			},
			expected: rep_models.ReplicationPolicy{
				ID:          1,
				Name:        "policy",
				Description: "description",
				Filters: []rep_models.Filter{
					rep_models.Filter{
						Kind:    "filter_kind_01",
						Pattern: "*",
					},
				},
				ReplicateDeletion: true,
				Trigger: &rep_models.Trigger{
					Kind: "trigger_kind_01",
				},
				ProjectIDs: []int64{1},
				Namespaces: []string{"library"},
				TargetIDs:  []int64{1},
			},
		},
	}

	for _, c := range cases {
		assert.EqualValues(t, c.expected, convertToRepPolicy(c.input))
	}
}
