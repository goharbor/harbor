package rbac

import (
	"context"
	"github.com/goharbor/harbor/src/pkg/rbac/model"
	"github.com/goharbor/harbor/src/testing/mock"
	"github.com/goharbor/harbor/src/testing/pkg/rbac/dao"
	"github.com/stretchr/testify/suite"
	"testing"
)

type managerTestSuite struct {
	suite.Suite
	mgr *manager
	dao *dao.DAO
}

func (m *managerTestSuite) SetupTest() {
	m.dao = &dao.DAO{}
	m.mgr = &manager{
		dao: m.dao,
	}
}

func (m *managerTestSuite) TestCreatePermission() {
	m.dao.On("CreatePermission", mock.Anything, mock.Anything).Return(int64(1), nil)
	_, err := m.mgr.CreatePermission(context.Background(), &model.RolePermission{})
	m.Require().Nil(err)
	m.dao.AssertExpectations(m.T())
}

func (m *managerTestSuite) TestDeletePermission() {
	m.dao.On("DeletePermission", mock.Anything, mock.Anything).Return(nil)
	err := m.mgr.DeletePermission(context.Background(), 1)
	m.Require().Nil(err)
	m.dao.AssertExpectations(m.T())
}

func (m *managerTestSuite) TestListPermissions() {
	m.dao.On("ListPermissions", mock.Anything, mock.Anything).Return([]*model.RolePermission{
		{
			ID:                 1,
			RoleType:           "robot",
			RoleID:             2,
			PermissionPolicyID: 3,
		},
	}, nil)
	rpers, err := m.mgr.ListPermissions(context.Background(), nil)
	m.Require().Nil(err)
	m.Equal(1, len(rpers))
	m.dao.AssertExpectations(m.T())
}

func (m *managerTestSuite) TestDeletePermissionsByRole() {
	m.dao.On("DeletePermissionsByRole", mock.Anything, mock.Anything, mock.Anything).Return(nil)
	err := m.mgr.DeletePermissionsByRole(context.Background(), "robot", 1)
	m.Require().Nil(err)
	m.dao.AssertExpectations(m.T())
}

func (m *managerTestSuite) TestCreateRbacPolicy() {
	m.dao.On("CreateRbacPolicy", mock.Anything, mock.Anything).Return(int64(1), nil)
	_, err := m.mgr.CreateRbacPolicy(context.Background(), &model.PermissionPolicy{})
	m.Require().Nil(err)
	m.dao.AssertExpectations(m.T())
}

func (m *managerTestSuite) TestDeleteRbacPolicy() {
	m.dao.On("DeleteRbacPolicy", mock.Anything, mock.Anything).Return(nil)
	err := m.mgr.DeleteRbacPolicy(context.Background(), 1)
	m.Require().Nil(err)
	m.dao.AssertExpectations(m.T())
}

func (m *managerTestSuite) TestListRbacPolicies() {
	m.dao.On("ListRbacPolicies", mock.Anything, mock.Anything).Return([]*model.PermissionPolicy{
		{
			ID:       1,
			Scope:    "/system",
			Resource: "repository",
			Action:   "create",
		},
	}, nil)
	rpers, err := m.mgr.ListRbacPolicies(context.Background(), nil)
	m.Require().Nil(err)
	m.Equal(1, len(rpers))
	m.dao.AssertExpectations(m.T())
}

func (m *managerTestSuite) TestGetPermissionsByRole() {
	m.dao.On("GetPermissionsByRole", mock.Anything, mock.Anything, mock.Anything).Return([]*model.UniversalRolePermission{
		{
			RoleType: "robot",
			RoleID:   1,
			Scope:    "/system",
			Resource: "repository",
			Action:   "create",
		},
	}, nil)
	rpers, err := m.mgr.GetPermissionsByRole(context.Background(), "robot", 1)
	m.Require().Nil(err)
	m.Equal(1, len(rpers))
	m.dao.AssertExpectations(m.T())
}

func TestManager(t *testing.T) {
	suite.Run(t, &managerTestSuite{})
}
