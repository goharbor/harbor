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

package promgr

import (
	"testing"

	"github.com/goharbor/harbor/src/common/models"
	"github.com/goharbor/harbor/src/core/promgr/pmsdriver"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type fakePMSDriver struct {
	project *models.Project
}

func newFakePMSDriver() pmsdriver.PMSDriver {
	return &fakePMSDriver{
		project: &models.Project{
			ProjectID: 1,
			Name:      "library",
			Metadata: map[string]string{
				models.ProMetaPublic: "true",
			},
		},
	}
}

func (f *fakePMSDriver) Get(projectIDOrName interface{}) (*models.Project, error) {
	return f.project, nil
}

func (f *fakePMSDriver) Create(*models.Project) (int64, error) {
	return 1, nil
}

func (f *fakePMSDriver) Delete(projectIDOrName interface{}) error {
	return nil
}

func (f *fakePMSDriver) Update(projectIDOrName interface{}, project *models.Project) error {
	return nil
}

func (f *fakePMSDriver) List(query *models.ProjectQueryParam) (*models.ProjectQueryResult, error) {
	return &models.ProjectQueryResult{
		Total:    1,
		Projects: []*models.Project{f.project},
	}, nil
}

var (
	proMgr = NewDefaultProjectManager(newFakePMSDriver(), false)
)

func TestGet(t *testing.T) {
	project, err := proMgr.Get(1)
	require.Nil(t, err)
	assert.Equal(t, int64(1), project.ProjectID)
}

func TestCreate(t *testing.T) {
	id, err := proMgr.Create(&models.Project{
		Name:    "library",
		OwnerID: 1,
	})
	require.Nil(t, err)
	assert.Equal(t, int64(1), id)
}

func TestDelete(t *testing.T) {
	assert.Nil(t, proMgr.Delete(1))
}

func TestUpdate(t *testing.T) {
	assert.Nil(t, proMgr.Update(1,
		&models.Project{
			Metadata: map[string]string{
				models.ProMetaPublic: "true",
			},
		}))
}

func TestList(t *testing.T) {
	result, err := proMgr.List(nil)
	require.Nil(t, err)
	assert.Equal(t, int64(1), result.Total)
	assert.Equal(t, int64(1), result.Projects[0].ProjectID)
}

func TestIsPublic(t *testing.T) {
	public, err := proMgr.IsPublic(1)
	require.Nil(t, err)
	assert.True(t, public)
}

func TestExist(t *testing.T) {
	exist, err := proMgr.Exists(1)
	require.Nil(t, err)
	assert.True(t, exist)
}

func TestGetPublic(t *testing.T) {
	projects, err := proMgr.GetPublic()
	require.Nil(t, err)
	assert.Equal(t, 1, len(projects))
	assert.True(t, projects[0].IsPublic())
}

func TestGetAuthorized(t *testing.T) {
	projects, err := proMgr.GetAuthorized(nil)
	require.Nil(t, err)
	assert.Len(t, projects, 0)

	projects, err = proMgr.GetAuthorized(&models.User{UserID: 1})
	require.Nil(t, err)
	assert.Len(t, projects, 1)
}

func TestGetMetadataManager(t *testing.T) {
	metaMgr := proMgr.GetMetadataManager()
	assert.Nil(t, metaMgr)
}
