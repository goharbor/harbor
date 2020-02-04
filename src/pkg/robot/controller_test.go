package robot

import (
	"github.com/goharbor/harbor/src/common"
	"github.com/goharbor/harbor/src/common/rbac"
	"github.com/goharbor/harbor/src/common/utils/test"
	core_cfg "github.com/goharbor/harbor/src/core/config"
	"github.com/goharbor/harbor/src/pkg/q"
	"github.com/goharbor/harbor/src/pkg/robot/model"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"testing"
)

type ControllerTestSuite struct {
	suite.Suite
	ctr     Controller
	t       *testing.T
	assert  *assert.Assertions
	require *require.Assertions

	robotID int64
}

// SetupSuite ...
func (s *ControllerTestSuite) SetupSuite() {
	test.InitDatabaseFromEnv()
	conf := map[string]interface{}{
		common.RobotTokenDuration: "30",
	}
	core_cfg.InitWithSettings(conf)
	s.t = s.T()
	s.assert = assert.New(s.t)
	s.require = require.New(s.t)
	s.ctr = RobotCtr
}

func (s *ControllerTestSuite) TestRobotAccount() {

	res := rbac.Resource("/project/1")

	rbacPolicy := &rbac.Policy{
		Resource: res.Subresource(rbac.ResourceRepository),
		Action:   "pull",
	}
	policies := []*rbac.Policy{}
	policies = append(policies, rbacPolicy)

	robot1 := &model.RobotCreate{
		Name:        "robot1",
		Description: "TestCreateRobotAccount",
		ProjectID:   int64(1),
		Access:      policies,
	}

	robot, err := s.ctr.CreateRobotAccount(robot1)
	s.require.Nil(err)
	s.require.Equal(robot.ProjectID, int64(1))
	s.require.Equal(robot.Description, "TestCreateRobotAccount")
	s.require.NotEmpty(robot.Token)
	s.require.Equal(robot.Name, common.RobotPrefix+"robot1")

	robotGet, err := s.ctr.GetRobotAccount(robot.ID)
	s.require.Nil(err)
	s.require.Equal(robotGet.ProjectID, int64(1))
	s.require.Equal(robotGet.Description, "TestCreateRobotAccount")

	robot.Disabled = true
	err = s.ctr.UpdateRobotAccount(robot)
	s.require.Nil(err)
	s.require.Equal(robot.Disabled, true)

	robot2 := &model.RobotCreate{
		Name:        "robot2",
		Description: "TestCreateRobotAccount",
		ProjectID:   int64(1),
		Access:      policies,
	}
	r2, _ := s.ctr.CreateRobotAccount(robot2)
	s.robotID = r2.ID

	robot3 := &model.RobotCreate{
		Name:        "robot3",
		Description: "TestCreateRobotAccount",
		ExpiresAt:   expiresAt,
		ProjectID:   int64(11),
		Access:      policies,
	}
	r3, _ := s.ctr.CreateRobotAccount(robot3)

	keywords := make(map[string]interface{})
	keywords["ProjectID"] = int64(1)
	query := &q.Query{
		Keywords: keywords,
	}
	robots, err := s.ctr.ListRobotAccount(query)
	s.require.Nil(err)
	s.require.Equal(len(robots), 2)
	s.require.Equal(robots[1].Name, common.RobotPrefix+"robot2")

	err = s.ctr.DeleteRobotAccount(robot.ID)
	s.require.Nil(err)
	err = s.ctr.DeleteRobotAccount(r3.ID)
	s.require.Nil(err)

	robots, err = s.ctr.ListRobotAccount(query)
	s.require.Equal(len(robots), 1)
}

// TearDownSuite clears env for test suite
func (s *ControllerTestSuite) TearDownSuite() {
	err := s.ctr.DeleteRobotAccount(s.robotID)
	require.NoError(s.T(), err, "delete robot")
}

// TestController ...
func TestController(t *testing.T) {
	suite.Run(t, new(ControllerTestSuite))
}
