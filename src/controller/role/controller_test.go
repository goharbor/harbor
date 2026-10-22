package role

import (
	"context"
	"testing"

	"github.com/jackc/pgx/v5/pgconn"
	"github.com/stretchr/testify/suite"

	"github.com/goharbor/harbor/src/lib/errors"
	"github.com/goharbor/harbor/src/lib/orm"
	"github.com/goharbor/harbor/src/pkg/permission/types"
	"github.com/goharbor/harbor/src/pkg/role/model"
	ormtesting "github.com/goharbor/harbor/src/testing/lib/orm"
	"github.com/goharbor/harbor/src/testing/mock"
	testmember "github.com/goharbor/harbor/src/testing/pkg/member"
	testproject "github.com/goharbor/harbor/src/testing/pkg/project"
	testrbac "github.com/goharbor/harbor/src/testing/pkg/rbac"
	testrole "github.com/goharbor/harbor/src/testing/pkg/role"
)

type ControllerTestSuite struct {
	suite.Suite
	roleMgr   *testrole.Manager
	rbacMgr   *testrbac.Manager
	proMgr    *testproject.Manager
	memberMgr *testmember.Manager
	c         controller
	// ctx carries a fake ormer so the controller's orm.WithTransaction wrappers
	// resolve an ormer from context; the actual DB calls are mocked at the manager level.
	ctx context.Context
}

func (suite *ControllerTestSuite) SetupTest() {
	suite.roleMgr = &testrole.Manager{}
	suite.rbacMgr = &testrbac.Manager{}
	suite.proMgr = &testproject.Manager{}
	suite.memberMgr = &testmember.Manager{}
	suite.c = controller{
		roleMgr:   suite.roleMgr,
		rbacMgr:   suite.rbacMgr,
		proMgr:    suite.proMgr,
		memberMgr: suite.memberMgr,
	}
	suite.ctx = orm.NewContext(context.TODO(), &ormtesting.FakeOrmer{})
}

func (suite *ControllerTestSuite) TestDeleteBuiltinRole() {
	suite.roleMgr.On("Get", mock.Anything, int64(1)).Return(&model.Role{
		ID:        1,
		Name:      "projectAdmin",
		IsBuiltin: true,
	}, nil)

	err := suite.c.Delete(suite.ctx, int64(1))
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
	suite.memberMgr.On("GetTotalOfProjectMembersByRole", mock.Anything, 2).Return(0, nil)
	suite.roleMgr.On("Delete", mock.Anything, int64(2)).Return(nil)
	suite.rbacMgr.On("DeletePermissionsByRole", mock.Anything, ROLETYPE, int64(2)).Return(nil)

	err := suite.c.Delete(suite.ctx, int64(2))
	suite.Nil(err)
}

// A custom role with no permission rows must still be deletable: the underlying
// DeletePermissionsByRole reports NotFound when it removes zero rows, and that
// must not roll back the (valid) role delete.
func (suite *ControllerTestSuite) TestDeleteCustomRoleWithoutPermissions() {
	suite.roleMgr.On("Get", mock.Anything, int64(4)).Return(&model.Role{
		ID:        4,
		Name:      "emptyRole",
		IsBuiltin: false,
	}, nil)
	suite.memberMgr.On("GetTotalOfProjectMembersByRole", mock.Anything, 4).Return(0, nil)
	suite.roleMgr.On("Delete", mock.Anything, int64(4)).Return(nil)
	suite.rbacMgr.On("DeletePermissionsByRole", mock.Anything, ROLETYPE, int64(4)).
		Return(errors.NotFoundError(nil).WithMessage("no permission rows"))

	err := suite.c.Delete(suite.ctx, int64(4))
	suite.Nil(err)
}

// A role still assigned to project members cannot be deleted (avoids orphaning
// project_member rows, which have no FK to role).
func (suite *ControllerTestSuite) TestDeleteAssignedRoleRejected() {
	suite.roleMgr.On("Get", mock.Anything, int64(3)).Return(&model.Role{
		ID:        3,
		Name:      "assignedRole",
		IsBuiltin: false,
	}, nil)
	suite.memberMgr.On("GetTotalOfProjectMembersByRole", mock.Anything, 3).Return(2, nil)

	err := suite.c.Delete(suite.ctx, int64(3))
	suite.Require().NotNil(err)
	suite.True(errors.IsErr(err, errors.PreconditionCode))
	suite.roleMgr.AssertNotCalled(suite.T(), "Delete", mock.Anything, mock.Anything)
}

func (suite *ControllerTestSuite) TestUpdateNilRole() {
	err := suite.c.Update(suite.ctx, nil, nil)
	suite.Require().NotNil(err)
	suite.True(errors.IsErr(err, errors.BadRequestCode))
}

func (suite *ControllerTestSuite) TestUpdateBuiltinRole() {
	suite.roleMgr.On("Get", mock.Anything, int64(1)).Return(&model.Role{
		ID:        1,
		Name:      "projectAdmin",
		IsBuiltin: true,
	}, nil)

	err := suite.c.Update(suite.ctx, &Role{
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
	// Update now persists name + description + the modification audit columns
	suite.roleMgr.On("Update", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil)
	suite.rbacMgr.On("DeletePermissionsByRole", mock.Anything, ROLETYPE, int64(2)).Return(nil)
	suite.rbacMgr.On("CreateRbacPolicy", mock.Anything, mock.Anything).Return(int64(1), nil)
	suite.rbacMgr.On("CreatePermission", mock.Anything, mock.Anything).Return(int64(1), nil)

	err := suite.c.Update(suite.ctx, &Role{
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

	id, err := suite.c.Create(suite.ctx, &Role{
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

// Failure injection: a permission insert failing must fail the whole Create
// (the transaction propagates the error) rather than leaving a named role with
// no permissions. Guards against silently dropping the transaction.
func (suite *ControllerTestSuite) TestCreateCustomRolePermissionFailure() {
	suite.roleMgr.On("Create", mock.Anything, mock.Anything).Return(int64(7), nil)
	suite.rbacMgr.On("CreateRbacPolicy", mock.Anything, mock.Anything).
		Return(int64(0), errors.New("insert failed"))

	_, err := suite.c.Create(suite.ctx, &Role{
		Role: model.Role{Name: "boom"},
		Permissions: []*Permission{
			{Access: []*types.Policy{{Resource: "repository", Action: "pull"}}},
		},
	})
	suite.Require().NotNil(err)
}

// A member assigned concurrently (after the count) makes the DB reject the role
// delete via the FK; the controller must surface that as a 412, not a 500.
func (suite *ControllerTestSuite) TestDeleteCustomRoleForeignKeyRace() {
	suite.roleMgr.On("Get", mock.Anything, int64(9)).Return(&model.Role{
		ID:        9,
		Name:      "racy",
		IsBuiltin: false,
	}, nil)
	suite.memberMgr.On("GetTotalOfProjectMembersByRole", mock.Anything, 9).Return(0, nil)
	suite.roleMgr.On("Delete", mock.Anything, int64(9)).
		Return(&pgconn.PgError{Code: "23503"}) // foreign_key_violation

	err := suite.c.Delete(suite.ctx, int64(9))
	suite.Require().NotNil(err)
	// The FK violation is surfaced as a foreign-key-constraint error, which maps
	// to HTTP 412 (same as the count guard's PreconditionFailed).
	suite.True(errors.IsErr(err, errors.ViolateForeignKeyConstraintCode))
}

func TestControllerTestSuite(t *testing.T) {
	suite.Run(t, &ControllerTestSuite{})
}
