// Copyright Project Harbor Authors
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

package dao

import (
	"testing"

	"github.com/goharbor/harbor/src/common/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var (
	project    = "library"
	name       = "library/repository-test"
	repository = &models.RepoRecord{
		Name:      name,
		ProjectID: 1,
	}
)

func TestGetTotalOfRepositories(t *testing.T) {
	total, err := GetTotalOfRepositories()
	require.Nil(t, err)

	err = addRepository(repository)
	require.Nil(t, err)
	defer deleteRepository(name)

	n, err := GetTotalOfRepositories()
	require.Nil(t, err)
	assert.Equal(t, total+1, n)
}

func addRepository(repository *models.RepoRecord) error {
	return AddRepository(*repository)
}

func deleteRepository(name string) error {
	return DeleteRepository(name)
}

func clearRepositoryData() error {
	return ClearTable(models.RepoTable)
}
