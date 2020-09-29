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

package local

import (
	"os"
	"testing"

	"github.com/goharbor/harbor/src/common/models"
	errutil "github.com/goharbor/harbor/src/common/utils/error"
	"github.com/goharbor/harbor/src/common/utils/test"
	"github.com/stretchr/testify/assert"
)

func TestMain(m *testing.M) {
	test.InitDatabaseFromEnv()
	os.Exit(m.Run())
}

func TestGet(t *testing.T) {
	pm := &driver{}

	// project name
	project, err := pm.Get("library")
	assert.Nil(t, err)
	assert.NotNil(t, project)
	assert.Equal(t, "library", project.Name)

	// project ID
	project, err = pm.Get(int64(1))
	assert.Nil(t, err)
	assert.NotNil(t, project)
	assert.Equal(t, int64(1), project.ProjectID)

	// non-exist project
	project, err = pm.Get("non-exist-project")
	assert.Nil(t, err)
	assert.Nil(t, project)

	// invalid type
	project, err = pm.Get(true)
	assert.NotNil(t, err)
}

func TestCreateAndDelete(t *testing.T) {
	pm := &driver{}

	// nil project
	_, err := pm.Create(nil)
	assert.NotNil(t, err)

	// nil project name
	_, err = pm.Create(&models.Project{
		OwnerID: 1,
	})
	assert.NotNil(t, err)

	// nil owner id and nil owner name
	_, err = pm.Create(&models.Project{
		Name:      "test",
		OwnerName: "non_exist_user",
	})
	assert.NotNil(t, err)

	// valid project, owner id
	id, err := pm.Create(&models.Project{
		Name:    "test",
		OwnerID: 1,
	})
	assert.Nil(t, err)
	assert.Nil(t, pm.Delete(id))

	// valid project, owner name
	id, err = pm.Create(&models.Project{
		Name:      "test",
		OwnerName: "admin",
	})
	assert.Nil(t, err)
	assert.Nil(t, pm.Delete(id))

	// duplicate project name
	id, err = pm.Create(&models.Project{
		Name:      "test",
		OwnerName: "admin",
	})
	assert.Nil(t, err)
	defer pm.Delete(id)
	_, err = pm.Create(&models.Project{
		Name:      "test",
		OwnerName: "admin",
	})
	assert.Equal(t, errutil.ErrDupProject, err)
}

func TestUpdate(t *testing.T) {
	pm := &driver{}
	assert.Nil(t, pm.Update(1, nil))
}

func TestList(t *testing.T) {
	pm := &driver{}

	id, err := pm.Create(&models.Project{
		Name:    "get_all_test",
		OwnerID: 1,
		Metadata: map[string]string{
			models.ProMetaPublic: "true",
		},
	})
	assert.Nil(t, err)
	defer pm.Delete(id)

	// get by name
	result, err := pm.List(&models.ProjectQueryParam{
		Name: "get_all_test",
	})
	assert.Nil(t, err)
	assert.Equal(t, id, result.Projects[0].ProjectID)

	// get by owner
	result, err = pm.List(&models.ProjectQueryParam{
		Owner: "admin",
	})
	assert.Nil(t, err)
	exist := false
	for _, project := range result.Projects {
		if project.ProjectID == id {
			exist = true
			break
		}
	}
	assert.True(t, exist)

	// get by public
	value := true
	result, err = pm.List(&models.ProjectQueryParam{
		Public: &value,
	})
	assert.Nil(t, err)
	exist = false
	for _, project := range result.Projects {
		if project.ProjectID == id {
			exist = true
			break
		}
	}
	assert.True(t, exist)
}
