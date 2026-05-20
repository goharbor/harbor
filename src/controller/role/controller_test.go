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

func TestControllerTestSuite(t *testing.T) {
	suite.Run(t, &ControllerTestSuite{})
}

// legacy TODO tests preserved below
/*
import (
	"context"
	"os"
	"testing"

	"github.com/aliyun/alibaba-cloud-sdk-go/sdk/utils"
	"github.com/stretchr/testify/suite"

	"github.com/goharbor/harbor/src/common"
	"github.com/goharbor/harbor/src/common/security"
	"github.com/goharbor/harbor/src/common/utils/test"
	"github.com/goharbor/harbor/src/lib/config"
	"github.com/goharbor/harbor/src/lib/q"
	_ "github.com/goharbor/harbor/src/pkg/config/inmemory"
	"github.com/goharbor/harbor/src/pkg/permission/types"
	proModels "github.com/goharbor/harbor/src/pkg/project/models"
	rbac_model "github.com/goharbor/harbor/src/pkg/rbac/model"
	"github.com/goharbor/harbor/src/pkg/robot/model"
	htesting "github.com/goharbor/harbor/src/testing"
	testsec "github.com/goharbor/harbor/src/testing/common/security"
	"github.com/goharbor/harbor/src/testing/mock"
	"github.com/goharbor/harbor/src/testing/pkg/project"
	"github.com/goharbor/harbor/src/testing/pkg/rbac"
	"github.com/goharbor/harbor/src/testing/pkg/robot"
)

type ControllerTestSuite struct {
	htesting.Suite
}

func (suite *ControllerTestSuite) TestGet() {
	projectMgr := &project.Manager{}
	rbacMgr := &rbac.Manager{}
	robotMgr := &robot.Manager{}

	c := controller{robotMgr: robotMgr, rbacMgr: rbacMgr, proMgr: projectMgr}
	ctx := context.TODO()
	projectMgr.On("Get", mock.Anything, mock.Anything).Return(&proModels.Project{ProjectID: 1, Name: "library"}, nil)
	robotMgr.On("Get", mock.Anything, mock.Anything).Return(&model.Robot{
		Name:        "library+test",
		Description: "test get method",
		ProjectID:   1,
		Secret:      utils.GetNonce(),
	}, nil)
	rbacMgr.On("GetPermissionsByRole", mock.Anything, mock.Anything, mock.Anything).Return([]*rbac_model.UniversalRolePermission{
		{
			RoleType: ROBOTTYPE,
			RoleID:   1,
			Scope:    "/project/1",
			Resource: "repository",
			Action:   "pull",
		},
		{
			RoleType: ROBOTTYPE,
			RoleID:   1,
			Scope:    "/project/1",
			Resource: "repository",
			Action:   "push",
		},
	}, nil)
	robot, err := c.Get(ctx, int64(1), &Option{
		WithPermission: true,
	})
	suite.Nil(err)

	suite.Equal("project", robot.Permissions[0].Kind)
	suite.Equal("library", robot.Permissions[0].Namespace)
	suite.Equal("pull", robot.Permissions[0].Access[0].Action.String())
	suite.Equal("project", robot.Level)

}

func (suite *ControllerTestSuite) TestCount() {
	projectMgr := &project.Manager{}
	rbacMgr := &rbac.Manager{}
	robotMgr := &robot.Manager{}

	c := controller{robotMgr: robotMgr, rbacMgr: rbacMgr, proMgr: projectMgr}
	ctx := context.TODO()

	robotMgr.On("Count", mock.Anything, mock.Anything).Return(int64(1), nil)

	ct, err := c.Count(ctx, nil)
	suite.Nil(err)
	suite.Equal(int64(1), ct)
}

func (suite *ControllerTestSuite) TestCreate() {
	secretKeyPath := "/tmp/secretkey"
	_, err := test.GenerateKey(secretKeyPath)
	suite.Nil(err)
	defer os.Remove(secretKeyPath)
	suite.T().Setenv("KEY_PATH", secretKeyPath)

	conf := map[string]any{
		common.RobotTokenDuration: "30",
	}
	config.InitWithSettings(conf)

	projectMgr := &project.Manager{}
	rbacMgr := &rbac.Manager{}
	robotMgr := &robot.Manager{}

	c := controller{robotMgr: robotMgr, rbacMgr: rbacMgr, proMgr: projectMgr}
	secCtx := &testsec.Context{}
	secCtx.On("GetUsername").Return("security-context-user")
	ctx := security.NewContext(context.Background(), secCtx)
	projectMgr.On("Get", mock.Anything, mock.Anything).Return(&proModels.Project{ProjectID: 1, Name: "library"}, nil)
	robotMgr.On("Create", mock.Anything, mock.Anything).Return(int64(1), nil)
	rbacMgr.On("CreateRbacPolicy", mock.Anything, mock.Anything, mock.Anything).Return(int64(1), nil)
	rbacMgr.On("CreatePermission", mock.Anything, mock.Anything, mock.Anything).Return(int64(1), nil)

	id, _, err := c.Create(ctx, &Robot{
		Robot: model.Robot{
			Name:        "testcreate",
			Description: "testcreate",
			Duration:    0,
		},
		ProjectName: "library",
		Level:       LEVELPROJECT,
		Permissions: []*Permission{
			{
				Kind:      "project",
				Namespace: "library",
				Access: []*types.Policy{
					{
						Resource: "repository",
						Action:   "push",
					},
					{
						Resource: "repository",
						Action:   "pull",
					},
				},
			},
		},
	})
	suite.Nil(err)
	suite.Equal(int64(1), id)
}

func (suite *ControllerTestSuite) TestDelete() {
	projectMgr := &project.Manager{}
	rbacMgr := &rbac.Manager{}
	robotMgr := &robot.Manager{}

	c := controller{robotMgr: robotMgr, rbacMgr: rbacMgr, proMgr: projectMgr}
	ctx := context.TODO()

	robotMgr.On("Get", mock.Anything, mock.Anything).Return(&model.Robot{
		Name:        "library+test",
		Description: "test get method",
		ProjectID:   1,
		Secret:      utils.GetNonce(),
	}, nil)
	robotMgr.On("Delete", mock.Anything, mock.Anything).Return(nil)
	rbacMgr.On("DeletePermissionsByRole", mock.Anything, mock.Anything, mock.Anything).Return(nil)

	err := c.Delete(ctx, int64(1))
	suite.Nil(err)
}

func (suite *ControllerTestSuite) TestUpdate() {
	projectMgr := &project.Manager{}
	rbacMgr := &rbac.Manager{}
	robotMgr := &robot.Manager{}

	c := controller{robotMgr: robotMgr, rbacMgr: rbacMgr, proMgr: projectMgr}
	ctx := context.TODO()

	conf := map[string]any{
		common.RobotPrefix: "robot$",
	}
	config.InitWithSettings(conf)

	robotMgr.On("Update", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil)
	projectMgr.On("Get", mock.Anything, mock.Anything).Return(&proModels.Project{ProjectID: 1, Name: "library"}, nil)
	rbacMgr.On("DeletePermissionsByRole", mock.Anything, mock.Anything, mock.Anything).Return(nil)

	rbacMgr.On("CreateRbacPolicy", mock.Anything, mock.Anything, mock.Anything).Return(int64(1), nil)
	rbacMgr.On("CreatePermission", mock.Anything, mock.Anything, mock.Anything).Return(int64(1), nil)

	err := c.Update(ctx, &Robot{
		Robot: model.Robot{
			Name:        "testcreate",
			Description: "testcreate",
			Duration:    0,
		},
		ProjectName: "library",
		Level:       LEVELPROJECT,
		Permissions: []*Permission{
			{
				Kind:      "project",
				Namespace: "library",
				Access: []*types.Policy{
					{
						Resource: "repository",
						Action:   "push",
					},
					{
						Resource: "repository",
						Action:   "pull",
					},
				},
			},
		},
	}, &Option{
		WithPermission: true,
	})
	suite.Nil(err)
}

func (suite *ControllerTestSuite) TestList() {
	projectMgr := &project.Manager{}
	rbacMgr := &rbac.Manager{}
	robotMgr := &robot.Manager{}

	c := controller{robotMgr: robotMgr, rbacMgr: rbacMgr, proMgr: projectMgr}
	ctx := context.TODO()

	projectMgr.On("Get", mock.Anything, mock.Anything).Return(&proModels.Project{ProjectID: 1, Name: "library"}, nil)
	robotMgr.On("List", mock.Anything, mock.Anything).Return([]*model.Robot{
		{
			Name:        "test",
			Description: "test list method",
			ProjectID:   1,
			Secret:      utils.GetNonce(),
		},
	}, nil)
	rbacMgr.On("GetPermissionsByRole", mock.Anything, mock.Anything, mock.Anything).Return([]*rbac_model.UniversalRolePermission{
		{
			RoleType: ROBOTTYPE,
			RoleID:   1,
			Scope:    "/project/1",
			Resource: "repository",
			Action:   "pull",
		},
		{
			RoleType: ROBOTTYPE,
			RoleID:   1,
			Scope:    "/project/1",
			Resource: "repository",
			Action:   "push",
		},
	}, nil)
	projectMgr.On("Get", mock.Anything, mock.Anything).Return(&proModels.Project{ProjectID: 1, Name: "library"}, nil)
	rs, err := c.List(ctx, &q.Query{
		Keywords: map[string]any{
			"name": "test3",
		},
	}, &Option{
		WithPermission: true,
	})
	suite.Nil(err)
	suite.Equal("project", rs[0].Permissions[0].Kind)
	suite.Equal("library", rs[0].Permissions[0].Namespace)
	suite.Equal("pull", rs[0].Permissions[0].Access[0].Action.String())
	suite.Equal("project", rs[0].Level)

}

func (suite *ControllerTestSuite) TestToScope() {
	projectMgr := &project.Manager{}
	rbacMgr := &rbac.Manager{}
	robotMgr := &robot.Manager{}

	c := controller{robotMgr: robotMgr, rbacMgr: rbacMgr, proMgr: projectMgr}
	ctx := context.TODO()

	projectMgr.On("Get", mock.Anything, mock.Anything).Return(&proModels.Project{ProjectID: 1, Name: "library"}, nil)

	p := &Permission{
		Kind:      "system",
		Namespace: "/",
	}
	scope, err := c.toScope(ctx, p)
	suite.Nil(err)
	suite.Equal("/system", scope)

	p = &Permission{
		Kind:      "system",
		Namespace: "&",
	}
	_, err = c.toScope(ctx, p)
	suite.NotNil(err)

	p = &Permission{
		Kind:      "project",
		Namespace: "library",
	}
	scope, err = c.toScope(ctx, p)
	suite.Nil(err)
	suite.Equal("/project/1", scope)

	p = &Permission{
		Kind:      "project",
		Namespace: "*",
	}
	scope, err = c.toScope(ctx, p)
	suite.Nil(err)
	suite.Equal("/project/*", scope)

}

func (suite *ControllerTestSuite) TestIsValidSec() {
	sec := "1234abcdABCD"
	suite.True(IsValidSec(sec))
	sec = "1234abcd"
	suite.False(IsValidSec(sec))
	sec = "123abc"
	suite.False(IsValidSec(sec))
	// secret of length 128 characters long should be ok
	sec = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789abcd"
	suite.True(IsValidSec(sec))
	// secret of length larger than 128 characters long, such as 129 characters long, should return false
	sec = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789abcde"
	suite.False(IsValidSec(sec))
}

func (suite *ControllerTestSuite) TestCreateSec() {
	_, pwd, _, err := CreateSec()
	suite.Nil(err)
	suite.True(IsValidSec(pwd))
}
func TestControllerTestSuite(t *testing.T) {
	suite.Run(t, &ControllerTestSuite{})
}

*/
