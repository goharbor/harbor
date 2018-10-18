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

package admiral

import (
	"net/http"
	"sort"
	"testing"

	"github.com/goharbor/harbor/src/common/models"
	errutil "github.com/goharbor/harbor/src/common/utils/error"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var (
	client      = http.DefaultClient
	endpoint    = "http://127.0.0.1:8282"
	tokenReader = &RawTokenReader{
		Token: "token",
	}
)

func TestConvert(t *testing.T) {
	// nil project
	pro, err := convert(nil)
	assert.Nil(t, err)
	assert.Nil(t, pro)

	// project without property __projectIndex
	p := &project{}
	pro, err = convert(p)
	assert.NotNil(t, err)
	assert.Nil(t, pro)

	// project with invalid __projectIndex
	p = &project{
		CustomProperties: map[string]string{
			"__projectIndex": "invalid_value",
		},
	}
	pro, err = convert(p)
	assert.NotNil(t, err)
	assert.Nil(t, pro)

	// project with invalid __enableContentTrust
	p = &project{
		CustomProperties: map[string]string{
			"__enableContentTrust": "invalid_value",
		},
	}
	pro, err = convert(p)
	assert.NotNil(t, err)
	assert.Nil(t, pro)

	// project with invalid __preventVulnerableImagesFromRunning
	p = &project{
		CustomProperties: map[string]string{
			"__preventVulnerableImagesFromRunning": "invalid_value",
		},
	}
	pro, err = convert(p)
	assert.NotNil(t, err)
	assert.Nil(t, pro)

	// project with invalid __automaticallyScanImagesOnPush
	p = &project{
		CustomProperties: map[string]string{
			"__automaticallyScanImagesOnPush": "invalid_value",
		},
	}
	pro, err = convert(p)
	assert.NotNil(t, err)
	assert.Nil(t, pro)

	// valid project
	p = &project{
		Name:   "test",
		Public: true,
		CustomProperties: map[string]string{
			"__projectIndex":                               "1",
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
	assert.True(t, pro.IsPublic())
	assert.Equal(t, int64(1), pro.ProjectID)
	assert.True(t, pro.ContentTrustEnabled())
	assert.True(t, pro.VulPrevented())
	assert.Equal(t, "medium", pro.Severity())
	assert.True(t, pro.AutoScan())
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
        "__projectIndex": "2",
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
        "__projectIndex": "2",
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
	d := NewDriver(client, endpoint, tokenReader)
	name := "project_for_test_get"
	id, err := d.Create(&models.Project{
		Name: name,
	})
	require.Nil(t, err)
	defer delete(t, id)

	// get by invalid input type
	_, err = d.Get([]string{})
	assert.NotNil(t, err)

	// get by invalid ID
	project, err := d.Get(int64(0))
	assert.Nil(t, err)
	assert.Nil(t, project)

	// get by invalid name
	project, err = d.Get("invalid_name")
	assert.Nil(t, err)
	assert.Nil(t, project)

	// get by valid ID
	project, err = d.Get(id)
	assert.Nil(t, err)
	assert.Equal(t, id, project.ProjectID)

	// get by valid name
	project, err = d.Get(name)
	assert.Nil(t, err)
	assert.Equal(t, id, project.ProjectID)
}

func TestCreate(t *testing.T) {
	d := NewDriver(client, endpoint, tokenReader)

	name := "project_for_test_create"
	id, err := d.Create(&models.Project{
		Name: name,
		Metadata: map[string]string{
			models.ProMetaPublic:             "true",
			models.ProMetaEnableContentTrust: "true",
			models.ProMetaPreventVul:         "true",
			models.ProMetaSeverity:           "medium",
			models.ProMetaAutoScan:           "true",
		},
	})
	require.Nil(t, err)
	defer delete(t, id)

	project, err := d.Get(id)
	assert.Nil(t, err)
	assert.Equal(t, name, project.Name)
	assert.True(t, project.IsPublic())
	assert.True(t, project.ContentTrustEnabled())
	assert.True(t, project.VulPrevented())
	assert.Equal(t, "medium", project.Severity())
	assert.True(t, project.AutoScan())

	// duplicate project name
	_, err = d.Create(&models.Project{
		Name: name,
	})
	assert.Equal(t, errutil.ErrDupProject, err)
}

func TestDelete(t *testing.T) {
	d := NewDriver(client, endpoint, tokenReader)

	// non-exist project
	err := d.Delete(int64(0))
	assert.NotNil(t, err)

	// delete by ID
	name := "project_for_pm_based_on_pms_id"
	id, err := d.Create(&models.Project{
		Name: name,
	})
	require.Nil(t, err)
	err = d.Delete(id)
	assert.Nil(t, err)

	// delete by name
	name = "project_for_pm_based_on_pms_name"
	id, err = d.Create(&models.Project{
		Name: name,
	})
	require.Nil(t, err)
	err = d.Delete(name)
	assert.Nil(t, err)
}

func TestUpdate(t *testing.T) {
	d := NewDriver(client, endpoint, tokenReader)
	err := d.Update(nil, nil)
	assert.NotNil(t, err)
}

func TestList(t *testing.T) {
	d := NewDriver(client, endpoint, tokenReader)

	name1 := "project_for_test_get_all_01"
	id1, err := d.Create(&models.Project{
		Name: name1,
	})
	require.Nil(t, err)
	defer delete(t, id1)

	name2 := "project_for_test_get_all_02"
	id2, err := d.Create(&models.Project{
		Name: name2,
		Metadata: map[string]string{
			models.ProMetaPublic: "true",
		},
	})
	require.Nil(t, err)
	defer delete(t, id2)

	// no filter
	result, err := d.List(nil)
	require.Nil(t, err)
	found1 := false
	found2 := false
	for _, project := range result.Projects {
		if project.ProjectID == id1 {
			found1 = true
		}
		if project.ProjectID == id2 {
			found2 = true
		}
	}
	assert.True(t, found1)
	assert.True(t, found2)

	// filter by name
	result, err = d.List(&models.ProjectQueryParam{
		Name: name1,
	})
	require.Nil(t, err)
	found1 = false
	for _, project := range result.Projects {
		if project.ProjectID == id1 {
			found1 = true
			break
		}
	}
	assert.True(t, found1)

	// filter by public
	value := true
	result, err = d.List(&models.ProjectQueryParam{
		Public: &value,
	})
	require.Nil(t, err)
	found2 = false
	for _, project := range result.Projects {
		if project.ProjectID == id2 {
			found2 = true
			break
		}
	}
	assert.True(t, found2)
}

func delete(t *testing.T, id int64) {
	d := NewDriver(client, endpoint, tokenReader)
	if err := d.Delete(id); err != nil {
		t.Logf("failed to delete project %d: %v", id, err)
	}
}
