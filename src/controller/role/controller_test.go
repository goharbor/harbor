package role

import (
	"context"
	"testing"

	"github.com/stretchr/testify/suite"

	"github.com/goharbor/harbor/src/lib/errors"
	"github.com/goharbor/harbor/src/pkg/permission/types"
	"github.com/goharbor/harbor/src/pkg/role/model"
	"github.com/goharbor/harbor/src/testing/mock"
	testproject "github.com/goharbor/harbor/src/testing/pkg/project"
	testrbac "github.com/goharbor/harbor/src/testing/pkg/rbac"
	testrole "github.com/goharbor/harbor/src/testing/pkg/role"
)

type ControllerTestSuite struct {
	suite.Suite
	roleMgr *testrole.Manager
	rbacMgr *testrbac.Manager
	proMgr  *testproject.Manager
	c       controller
}

func (suite *ControllerTestSuite) SetupTest() {
	suite.roleMgr = &testrole.Manager{}
	suite.rbacMgr = &testrbac.Manager{}
	suite.proMgr = &testproject.Manager{}
	suite.c = controller{
		roleMgr: suite.roleMgr,
		rbacMgr: suite.rbacMgr,
		proMgr:  suite.proMgr,
	}
}

func (suite *ControllerTestSuite) TestDeleteBuiltinRole() {
	suite.roleMgr.On("Get", mock.Anything, int64(1)).Return(&model.Role{
		ID:        1,
		Name:      "projectAdmin",
		IsBuiltin: true,
	}, nil)

	err := suite.c.Delete(context.TODO(), int64(1))
	suite.Require().NotNil(err)
	suite.True(errors.IsErr(err, errors.ForbiddenCode))
	suite.roleMgr.AssertNotCalled(suite.T(), "Delete", mock.Anything, mock.Anything)
}

func (suite *ControllerTestSuite) TestDeleteCustomRole() {
	suite.roleMgr.On("Get", mock.Anything, int64(2)).Return(&model.Role{
		ID:        2,
		Name:      "myCustomRole",
		IsBuiltin: false,
	}, nil)
	suite.roleMgr.On("Delete", mock.Anything, int64(2)).Return(nil)
	suite.rbacMgr.On("DeletePermissionsByRole", mock.Anything, ROLETYPE, int64(2)).Return(nil)

	err := suite.c.Delete(context.TODO(), int64(2))
	suite.Nil(err)
}

func (suite *ControllerTestSuite) TestUpdateNilRole() {
	err := suite.c.Update(context.TODO(), nil, nil)
	suite.Require().NotNil(err)
	suite.True(errors.IsErr(err, errors.BadRequestCode))
}

func (suite *ControllerTestSuite) TestUpdateBuiltinRole() {
	suite.roleMgr.On("Get", mock.Anything, int64(1)).Return(&model.Role{
		ID:        1,
		Name:      "projectAdmin",
		IsBuiltin: true,
	}, nil)

	err := suite.c.Update(context.TODO(), &Role{
		Role: model.Role{ID: 1, Name: "projectAdmin"},
	}, &Option{WithPermission: true})
	suite.Require().NotNil(err)
	suite.True(errors.IsErr(err, errors.ForbiddenCode))
	suite.rbacMgr.AssertNotCalled(suite.T(), "DeletePermissionsByRole", mock.Anything, mock.Anything, mock.Anything)
}

func (suite *ControllerTestSuite) TestUpdateCustomRole() {
	suite.roleMgr.On("Get", mock.Anything, int64(2)).Return(&model.Role{
		ID:        2,
		Name:      "myCustomRole",
		IsBuiltin: false,
	}, nil)
	suite.roleMgr.On("Update", mock.Anything, mock.Anything, mock.Anything).Return(nil)
	suite.rbacMgr.On("DeletePermissionsByRole", mock.Anything, ROLETYPE, int64(2)).Return(nil)
	suite.rbacMgr.On("CreateRbacPolicy", mock.Anything, mock.Anything).Return(int64(1), nil)
	suite.rbacMgr.On("CreatePermission", mock.Anything, mock.Anything).Return(int64(1), nil)

	err := suite.c.Update(context.TODO(), &Role{
		Role: model.Role{ID: 2, Name: "myCustomRole"},
		Permissions: []*Permission{
			{
				Access: []*types.Policy{
					{Resource: "repository", Action: "pull"},
				},
			},
		},
	}, &Option{WithPermission: true})
	suite.Nil(err)
}

func (suite *ControllerTestSuite) TestCreateCustomRole() {
	suite.roleMgr.On("Create", mock.Anything, mock.Anything).Return(int64(7), nil)
	suite.rbacMgr.On("CreateRbacPolicy", mock.Anything, mock.Anything).Return(int64(1), nil)
	suite.rbacMgr.On("CreatePermission", mock.Anything, mock.Anything).Return(int64(1), nil)

	id, err := suite.c.Create(context.TODO(), &Role{
		Role: model.Role{Name: "myCustomRole"},
		Permissions: []*Permission{
			{
				Access: []*types.Policy{
					{Resource: "repository", Action: "pull"},
				},
			},
		},
	})
	suite.Nil(err)
	suite.Equal(int64(7), id)
}

func TestControllerTestSuite(t *testing.T) {
	suite.Run(t, &ControllerTestSuite{})
}
