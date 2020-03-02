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

package project

import (
	"fmt"

	"github.com/goharbor/harbor/src/common/dao"
	"github.com/goharbor/harbor/src/common/models"
)

var (
	// Mgr is the global project manager
	Mgr = New()
)

// Manager is used for project management
// currently, the interface only defines the methods needed for tag retention
// will expand it when doing refactor
type Manager interface {
	// List projects according to the query
	List(...*models.ProjectQueryParam) ([]*models.Project, error)
	// Get the project specified by the ID or name
	Get(interface{}) (*models.Project, error)
}

// New returns a default implementation of Manager
func New() Manager {
	return &manager{}
}

type manager struct{}

// List projects according to the query
func (m *manager) List(query ...*models.ProjectQueryParam) ([]*models.Project, error) {
	var q *models.ProjectQueryParam
	if len(query) > 0 {
		q = query[0]
	}
	return dao.GetProjects(q)
}

// Get the project specified by the ID
func (m *manager) Get(idOrName interface{}) (*models.Project, error) {
	id, ok := idOrName.(int64)
	if ok {
		return dao.GetProjectByID(id)
	}
	name, ok := idOrName.(string)
	if ok {
		return dao.GetProjectByName(name)
	}
	return nil, fmt.Errorf("invalid parameter: %v, should be ID(int64) or name(string)", idOrName)
}
