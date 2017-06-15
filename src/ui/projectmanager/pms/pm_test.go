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

package pms

import (
	"sort"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/vmware/harbor/src/common/models"
)

var (
	endpoint = "http://127.0.0.1:8282"
	token    = ""
)

func TestConvert(t *testing.T) {
	//nil project
	pro, err := convert(nil)
	assert.Nil(t, err)
	assert.Nil(t, pro)

	//project without property __harborId
	p := &project{}
	pro, err = convert(p)
	assert.NotNil(t, err)
	assert.Nil(t, pro)

	//project with invalid __harborId
	p = &project{
		CustomProperties: map[string]string{
			"__harborId": "invalid_value",
		},
	}
	pro, err = convert(p)
	assert.NotNil(t, err)
	assert.Nil(t, pro)

	//project with invalid __enableContentTrust
	p = &project{
		CustomProperties: map[string]string{
			"__enableContentTrust": "invalid_value",
		},
	}
	pro, err = convert(p)
	assert.NotNil(t, err)
	assert.Nil(t, pro)

	//project with invalid __preventVulnerableImagesFromRunning
	p = &project{
		CustomProperties: map[string]string{
			"__preventVulnerableImagesFromRunning": "invalid_value",
		},
	}
	pro, err = convert(p)
	assert.NotNil(t, err)
	assert.Nil(t, pro)

	//project with invalid __automaticallyScanImagesOnPush
	p = &project{
		CustomProperties: map[string]string{
			"__automaticallyScanImagesOnPush": "invalid_value",
		},
	}
	pro, err = convert(p)
	assert.NotNil(t, err)
	assert.Nil(t, pro)

	//valid project
	p = &project{
		Name:   "test",
		Public: true,
		CustomProperties: map[string]string{
			"__harborId":                                   "1",
			"__enableContentTrust":                         "true",
			"__preventVulnerableImagesFromRunning":         "true",
			"__preventVulnerableImagesFromRunningSeverity": "medium",
			"__automaticallyScanImagesOnPush":              "true",
		},
	}
	pro, err = convert(p)
	assert.Nil(t, err)
	assert.NotNil(t, pro)
	assert.Equal(t, "test", pro.Name)
	assert.Equal(t, 1, pro.Public)
	assert.Equal(t, int64(1), pro.ProjectID)
	assert.True(t, pro.EnableContentTrust)
	assert.True(t, pro.PreventVulnerableImagesFromRunning)
	assert.Equal(t, "medium", pro.PreventVulnerableImagesFromRunningSeverity)
	assert.True(t, pro.AutomaticallyScanImagesOnPush)
}

func TestParse(t *testing.T) {
	data := `{
  "totalCount": 2,
  "documentLinks": [
    "/projects/default-project",
    "/projects/fc6c6c7ddd430875551449a65e7c8"
  ],
  "documents": {
    "/projects/fc6c6c7ddd430875551449a65e7c8": {
      "isPublic": false,
      "description": "This is a test project.",
      "id": "41427587-70e9-4671-9a9e-b9def0a07bb7",
      "name": "project02",
      "customProperties": {
        "__harborId": "2",
        "__enableContentTrust": "true",
        "__preventVulnerableImagesFromRunning": "true",
        "__preventVulnerableImagesFromRunningSeverity": "medium",
        "__automaticallyScanImagesOnPush": "false"
      },
      "documentVersion": 0,
      "documentEpoch": 0,
      "documentKind": "com:vmware:admiral:auth:project:ProjectService:ProjectState",
      "documentSelfLink": "/projects/fc6c6c7ddd430875551449a65e7c8",
      "documentUpdateTimeMicros": 1496729973549001,
      "documentUpdateAction": "POST",
      "documentExpirationTimeMicros": 0,
      "documentOwner": "f65900c4-2b6a-4671-8cf7-c17340dd3d39"
    },
	"/projects/default-project": {
      "isPublic": false,
      "administratorsUserGroupLink": "/core/authz/user-groups/fc6c6c7ddd43087555143835bcaf8",
      "membersUserGroupLink": "/core/authz/user-groups/fc6c6c7ddd43087555143835bde80",
      "id": "default-project",
      "name": "default-project",
      "customProperties": {
        "__harborId": "2",
        "__enableContentTrust": "true",
        "__preventVulnerableImagesFromRunning": "true",
        "__preventVulnerableImagesFromRunningSeverity": "medium",
        "__automaticallyScanImagesOnPush": "false"
      },
      "documentVersion": 0,
      "documentEpoch": 0,
      "documentKind": "com:vmware:admiral:auth:project:ProjectService:ProjectState",
      "documentSelfLink": "/projects/default-project",
      "documentUpdateTimeMicros": 1496725292012001,
      "documentUpdateAction": "POST",
      "documentExpirationTimeMicros": 0,
      "documentOwner": "f65900c4-2b6a-4671-8cf7-c17340dd3d39",
      "documentAuthPrincipalLink": "/core/authz/system-user"
    }
	},
  "documentCount": 2,
  "queryTimeMicros": 1,
  "documentVersion": 0,
  "documentUpdateTimeMicros": 0,
  "documentExpirationTimeMicros": 0,
  "documentOwner": "f65900c4-2b6a-4671-8cf7-c17340dd3d39"
}`

	projects, err := parse([]byte(data))
	assert.Nil(t, err)
	assert.Equal(t, 2, len(projects))

	ids := []string{projects[0].ID, projects[1].ID}
	sort.Strings(ids)

	assert.Equal(t, "default-project", ids[0])
	assert.Equal(t, "fc6c6c7ddd430875551449a65e7c8", ids[1])
}

func TestGet(t *testing.T) {
	pm := NewProjectManager(endpoint, token)
	name := "project_for_test_get"
	id, err := pm.Create(&models.Project{
		Name: name,
	})
	require.Nil(t, err)
	defer func(id int64) {
		if err := pm.Delete(id); err != nil {
			require.Nil(t, err)
		}
	}(id)

	// get by invalid input type
	_, err = pm.Get([]string{})
	assert.NotNil(t, err)

	// get by invalid ID
	project, err := pm.Get(int64(0))
	assert.Nil(t, err)
	assert.Nil(t, project)

	// get by invalid name
	project, err = pm.Get("invalid_name")
	assert.Nil(t, err)
	assert.Nil(t, project)

	// get by valid ID
	project, err = pm.Get(id)
	assert.Nil(t, err)
	assert.Equal(t, id, project.ProjectID)

	// get by valid name
	project, err = pm.Get(name)
	assert.Nil(t, err)
	assert.Equal(t, id, project.ProjectID)
}

func TestIsPublic(t *testing.T) {
	pm := NewProjectManager(endpoint, token)

	// invalid input type
	public, err := pm.IsPublic([]string{})
	assert.NotNil(t, err)
	assert.False(t, public)

	// non-exist project
	public, err = pm.IsPublic(int64(0))
	assert.Nil(t, err)
	assert.False(t, public)

	// public project
	name := "project_for_pm_based_on_pms_public"
	id, err := pm.Create(&models.Project{
		Name:   name,
		Public: 1,
	})
	require.Nil(t, err)
	defer func(id int64) {
		if err := pm.Delete(id); err != nil {
			require.Nil(t, err)
		}
	}(id)

	public, err = pm.IsPublic(id)
	assert.Nil(t, err)
	assert.True(t, public)

	public, err = pm.IsPublic(name)
	assert.Nil(t, err)
	assert.True(t, public)

	// private project
	name = "project_for_pm_based_on_pms_private"
	id, err = pm.Create(&models.Project{
		Name:   name,
		Public: 0,
	})
	require.Nil(t, err)
	defer func(id int64) {
		if err := pm.Delete(id); err != nil {
			require.Nil(t, err)
		}
	}(id)

	public, err = pm.IsPublic(id)
	assert.Nil(t, err)
	assert.False(t, public)

	public, err = pm.IsPublic(name)
	assert.Nil(t, err)
	assert.False(t, public)
}

func TestExist(t *testing.T) {
	pm := NewProjectManager(endpoint, token)

	// invalid input type
	exist, err := pm.Exist([]string{})
	assert.NotNil(t, err)
	assert.False(t, exist)

	// non-exist project
	exist, err = pm.Exist(int64(0))
	assert.Nil(t, err)
	assert.False(t, exist)

	// exist project
	name := "project_for_test_exist"
	id, err := pm.Create(&models.Project{
		Name: name,
	})
	require.Nil(t, err)
	defer func(id int64) {
		if err := pm.Delete(id); err != nil {
			require.Nil(t, err)
		}
	}(id)

	exist, err = pm.Exist(id)
	assert.Nil(t, err)
	assert.True(t, exist)

	exist, err = pm.Exist(name)
	assert.Nil(t, err)
	assert.True(t, exist)
}

func TestGetRoles(t *testing.T) {
	pm := NewProjectManager(endpoint, token)

	// nil username, nil project
	roles, err := pm.GetRoles("", nil)
	assert.Nil(t, err)
	assert.Zero(t, len(roles))

	// non-exist project
	_, err = pm.GetRoles("user01", "non_exist_project")
	assert.NotNil(t, err)

	// exist project
	name := "project_for_test_get_roles"
	id, err := pm.Create(&models.Project{
		Name: name,
	})
	require.Nil(t, err)
	defer func(id int64) {
		if err := pm.Delete(id); err != nil {
			require.Nil(t, err)
		}
	}(id)

	roles, err = pm.GetRoles("user01", id)
	assert.Nil(t, err)
	assert.Zero(t, len(roles))

	// TODO add test cases for real role of user
}

func TestGetPublic(t *testing.T) {
	pm := NewProjectManager(endpoint, token)

	projects, err := pm.GetPublic()
	assert.Nil(t, nil)
	size := len(projects)

	name := "project_for_test_get_public"
	id, err := pm.Create(&models.Project{
		Name:   name,
		Public: 1,
	})
	require.Nil(t, err)
	defer func(id int64) {
		if err := pm.Delete(id); err != nil {
			require.Nil(t, err)
		}
	}(id)

	projects, err = pm.GetPublic()
	assert.Nil(t, nil)
	assert.Equal(t, size+1, len(projects))

	found := false
	for _, project := range projects {
		if project.ProjectID == id {
			found = true
			break
		}
	}
	assert.True(t, found)
}

// TODO add test case
func TestGetByMember(t *testing.T) {

}

func TestCreate(t *testing.T) {
	pm := NewProjectManager(endpoint, token)

	name := "project_for_test_create"
	id, err := pm.Create(&models.Project{
		Name:                                       name,
		Public:                                     1,
		EnableContentTrust:                         true,
		PreventVulnerableImagesFromRunning:         true,
		PreventVulnerableImagesFromRunningSeverity: "medium",
		AutomaticallyScanImagesOnPush:              true,
	})
	require.Nil(t, err)
	defer func(id int64) {
		if err := pm.Delete(id); err != nil {
			require.Nil(t, err)
		}
	}(id)

	project, err := pm.Get(id)
	assert.Nil(t, err)
	assert.Equal(t, name, project.Name)
	assert.Equal(t, 1, project.Public)
	assert.True(t, project.EnableContentTrust)
	assert.True(t, project.PreventVulnerableImagesFromRunning)
	assert.Equal(t, "medium", project.PreventVulnerableImagesFromRunningSeverity)
	assert.True(t, project.AutomaticallyScanImagesOnPush)
}

func TestDelete(t *testing.T) {
	pm := NewProjectManager(endpoint, token)

	// non-exist project
	err := pm.Delete(int64(0))
	assert.NotNil(t, err)

	// delete by ID
	name := "project_for_pm_based_on_pms_id"
	id, err := pm.Create(&models.Project{
		Name: name,
	})
	require.Nil(t, err)
	err = pm.Delete(id)
	assert.Nil(t, err)

	// delete by name
	name = "project_for_pm_based_on_pms_name"
	id, err = pm.Create(&models.Project{
		Name: name,
	})
	require.Nil(t, err)
	err = pm.Delete(name)
	assert.Nil(t, err)
}

func TestUpdate(t *testing.T) {
	pm := NewProjectManager(endpoint, token)
	err := pm.Update(nil, nil)
	assert.NotNil(t, err)
}

func TestGetAll(t *testing.T) {
	pm := NewProjectManager(endpoint, token)
	_, err := pm.GetAll(nil)
	assert.NotNil(t, err)
}

func TestGetTotal(t *testing.T) {
	pm := NewProjectManager(endpoint, token)
	_, err := pm.GetTotal(nil)
	assert.NotNil(t, err)
}

// TODO add test case
func TestGetHasReadPerm(t *testing.T) {

}
