package robot

import (
	"context"
	"github.com/aliyun/alibaba-cloud-sdk-go/sdk/utils"
	"github.com/goharbor/harbor/src/common"
	"github.com/goharbor/harbor/src/common/utils/test"
	"github.com/goharbor/harbor/src/lib/config"
	"github.com/goharbor/harbor/src/lib/q"
	_ "github.com/goharbor/harbor/src/pkg/config/inmemory"
	"github.com/goharbor/harbor/src/pkg/permission/types"
	proModels "github.com/goharbor/harbor/src/pkg/project/models"
	rbac_model "github.com/goharbor/harbor/src/pkg/rbac/model"
	"github.com/goharbor/harbor/src/pkg/robot/model"
	htesting "github.com/goharbor/harbor/src/testing"
	"github.com/goharbor/harbor/src/testing/mock"
	"github.com/goharbor/harbor/src/testing/pkg/project"
	"github.com/goharbor/harbor/src/testing/pkg/rbac"
	"github.com/goharbor/harbor/src/testing/pkg/robot"
	"github.com/stretchr/testify/suite"
	"os"
	"testing"
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
		Secret:      utils.RandStringBytes(10),
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
	os.Setenv("KEY_PATH", secretKeyPath)

	conf := map[string]interface{}{
		common.RobotTokenDuration: "30",
	}
	config.InitWithSettings(conf)

	projectMgr := &project.Manager{}
	rbacMgr := &rbac.Manager{}
	robotMgr := &robot.Manager{}

	c := controller{robotMgr: robotMgr, rbacMgr: rbacMgr, proMgr: projectMgr}
	ctx := context.TODO()
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

	conf := map[string]interface{}{
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
			Secret:      utils.RandStringBytes(10),
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
		Keywords: map[string]interface{}{
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

func TestControllerTestSuite(t *testing.T) {
	suite.Run(t, &ControllerTestSuite{})
}
