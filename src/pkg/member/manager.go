//  Copyright Project Harbor Authors
//
//  Licensed under the Apache License, Version 2.0 (the "License");
//  you may not use this file except in compliance with the License.
//  You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
//  Unless required by applicable law or agreed to in writing, software
//  distributed under the License is distributed on an "AS IS" BASIS,
//  WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
//  See the License for the specific language governing permissions and
//  limitations under the License.

package member

import (
	"context"
	"github.com/goharbor/harbor/src/lib/errors"
	"github.com/goharbor/harbor/src/lib/q"
	"github.com/goharbor/harbor/src/pkg/member/dao"
	"github.com/goharbor/harbor/src/pkg/member/models"
)

var (
	// Mgr default project member manager
	Mgr = NewManager()
)

// Manager is used to manage the project member
type Manager interface {
	// AddProjectMember add project member
	AddProjectMember(ctx context.Context, member models.Member) (int, error)
	// Delete delete project member
	Delete(ctx context.Context, projectID int64, memberID int) error
	// Get get project member by ID
	Get(ctx context.Context, projectID int64, memberID int) (*models.Member, error)
	// List list the project member by conditions
	List(ctx context.Context, queryMember models.Member, query *q.Query) ([]*models.Member, error)
	// UpdateRole update project member's role
	UpdateRole(ctx context.Context, projectID int64, pmID int, role int) error
	// SearchMemberByName search project member by name
	SearchMemberByName(ctx context.Context, projectID int64, entityName string) ([]*models.Member, error)
	// DeleteMemberByUserID delete project member by user id
	DeleteMemberByUserID(ctx context.Context, uid int) error
	// GetTotalOfProjectMembers get the total amount of project members
	GetTotalOfProjectMembers(ctx context.Context, projectID int64, query *q.Query, roles ...int) (int, error)
	// ListRoles list project roles
	ListRoles(ctx context.Context, user *models.User, projectID int64) ([]int, error)
}

type manager struct {
	dao dao.DAO
}

func (m *manager) Get(ctx context.Context, projectID int64, memberID int) (*models.Member, error) {
	query := models.Member{
		ID:        memberID,
		ProjectID: projectID,
	}
	pm, err := m.dao.GetProjectMember(ctx, query, nil)
	if err != nil {
		return nil, err
	}
	if len(pm) == 0 {
		return nil, errors.NotFoundError(nil).
			WithMessage("the project member is not found, project id %v, member id %v", projectID, memberID)
	}
	return pm[0], nil
}

func (m *manager) AddProjectMember(ctx context.Context, member models.Member) (int, error) {
	return m.dao.AddProjectMember(ctx, member)
}

func (m *manager) UpdateRole(ctx context.Context, projectID int64, pmID int, role int) error {
	return m.dao.UpdateProjectMemberRole(ctx, projectID, pmID, role)
}

func (m *manager) SearchMemberByName(ctx context.Context, projectID int64, entityName string) ([]*models.Member, error) {
	return m.dao.SearchMemberByName(ctx, projectID, entityName)
}

func (m *manager) GetTotalOfProjectMembers(ctx context.Context, projectID int64, query *q.Query, roles ...int) (int, error) {
	return m.dao.GetTotalOfProjectMembers(ctx, projectID, query, roles...)
}

func (m *manager) ListRoles(ctx context.Context, user *models.User, projectID int64) ([]int, error) {
	return m.dao.ListRoles(ctx, user, projectID)
}

func (m *manager) List(ctx context.Context, queryMember models.Member, query *q.Query) ([]*models.Member, error) {
	return m.dao.GetProjectMember(ctx, queryMember, query)
}

func (m *manager) Delete(ctx context.Context, projectID int64, memberID int) error {
	return m.dao.DeleteProjectMemberByID(ctx, projectID, memberID)
}

func (m *manager) DeleteMemberByUserID(ctx context.Context, uid int) error {
	return m.dao.DeleteProjectMemberByUserID(ctx, uid)
}

// NewManager ...
func NewManager() Manager {
	return &manager{dao: dao.New()}
}
