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
	"context"
	"regexp"
	"strings"

	"github.com/goharbor/harbor/src/common/utils"
	"github.com/goharbor/harbor/src/lib/errors"
	"github.com/goharbor/harbor/src/lib/q"
	"github.com/goharbor/harbor/src/pkg/project/dao"
	"github.com/goharbor/harbor/src/pkg/project/models"
)

// Manager is used for project management
type Manager interface {
	// Create create project instance
	Create(ctx context.Context, project *models.Project) (int64, error)

	// Count returns the total count of projects according to the query
	Count(ctx context.Context, query *q.Query) (total int64, err error)

	// Delete delete the project instance by id
	Delete(ctx context.Context, id int64) error

	// Get the project specified by the ID or name
	Get(ctx context.Context, idOrName interface{}) (*models.Project, error)

	// List projects according to the query
	List(ctx context.Context, query *q.Query) ([]*models.Project, error)

	// ListRoles returns the roles of user for the specific project
	ListRoles(ctx context.Context, projectID int64, userID int, groupIDs ...int) ([]int, error)
}

// New returns a default implementation of Manager
func New() Manager {
	return &manager{dao: dao.New()}
}

const projectNameMaxLen int = 255
const projectNameMinLen int = 1
const restrictedNameChars = `[a-z0-9]+(?:[._-][a-z0-9]+)*`

var (
	validProjectName = regexp.MustCompile(`^` + restrictedNameChars + `$`)
)

type manager struct {
	dao dao.DAO
}

// Create create project instance
func (m *manager) Create(ctx context.Context, project *models.Project) (int64, error) {
	if project.OwnerID <= 0 {
		return 0, errors.BadRequestError(nil).WithMessage("Owner is missing when creating project %s", project.Name)
	}

	if utils.IsIllegalLength(project.Name, projectNameMinLen, projectNameMaxLen) {
		format := "Project name %s is illegal in length. (greater than %d or less than %d)"
		return 0, errors.BadRequestError(nil).WithMessage(format, project.Name, projectNameMaxLen, projectNameMinLen)
	}

	legal := validProjectName.MatchString(project.Name)
	if !legal {
		return 0, errors.BadRequestError(nil).WithMessage("project name is not in lower case or contains illegal characters")
	}

	return m.dao.Create(ctx, project)
}

// Count returns the total count of projects according to the query
func (m *manager) Count(ctx context.Context, query *q.Query) (total int64, err error) {
	return m.dao.Count(ctx, query)
}

// Delete delete the project instance by id
func (m *manager) Delete(ctx context.Context, id int64) error {
	return m.dao.Delete(ctx, id)
}

// Get the project specified by the ID
func (m *manager) Get(ctx context.Context, idOrName interface{}) (*models.Project, error) {
	id, ok := idOrName.(int64)
	if ok {
		return m.dao.Get(ctx, id)
	}
	name, ok := idOrName.(string)
	if ok {
		// check white space in project name
		if strings.Contains(name, " ") {
			return nil, errors.BadRequestError(nil).WithMessage("invalid project name: '%s'", name)
		}
		return m.dao.GetByName(ctx, name)
	}
	return nil, errors.Errorf("invalid parameter: %v, should be ID(int64) or name(string)", idOrName)
}

// List projects according to the query
func (m *manager) List(ctx context.Context, query *q.Query) ([]*models.Project, error) {
	return m.dao.List(ctx, query)
}

// Lists the roles of user for the specific project
func (m *manager) ListRoles(ctx context.Context, projectID int64, userID int, groupIDs ...int) ([]int, error) {
	return m.dao.ListRoles(ctx, projectID, userID, groupIDs...)
}
