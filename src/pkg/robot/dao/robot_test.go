package dao

import (
	"github.com/goharbor/harbor/src/common/dao"
	"github.com/goharbor/harbor/src/lib/q"
	"github.com/goharbor/harbor/src/pkg/robot/model"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"testing"
)

type robotAccountDaoTestSuite struct {
	suite.Suite
	require *require.Assertions
	assert  *assert.Assertions
	dao     RobotAccountDao
	id1     int64
	id2     int64
	id3     int64
	id4     int64
}

func (t *robotAccountDaoTestSuite) SetupSuite() {
	t.require = require.New(t.T())
	t.assert = assert.New(t.T())
	dao.PrepareTestForPostgresSQL()
	t.dao = New()
}

func (t *robotAccountDaoTestSuite) TestCreateRobotAccount() {
	robotName := "test1"
	robot := &model.Robot{
		Name:        robotName,
		Description: "test1 description",
		ProjectID:   1,
	}
	id, err := t.dao.CreateRobotAccount(robot)
	t.require.Nil(err)
	t.id1 = id
	t.require.Nil(err)
	t.require.NotNil(id)
}

func (t *robotAccountDaoTestSuite) TestGetRobotAccount() {
	robotName := "test2"
	robot := &model.Robot{
		Name:        robotName,
		Description: "test2 description",
		ProjectID:   1,
	}

	// add
	id, err := t.dao.CreateRobotAccount(robot)
	t.require.Nil(err)
	t.id2 = id

	robot, err = t.dao.GetRobotAccount(id)
	t.require.Nil(err)
	t.require.Equal(robotName, robot.Name)
}

func (t *robotAccountDaoTestSuite) TestListRobotAccounts() {
	robotName := "test3"
	robot := &model.Robot{
		Name:        robotName,
		Description: "test3 description",
		ProjectID:   1,
	}

	id, err := t.dao.CreateRobotAccount(robot)
	t.require.Nil(err)
	t.id3 = id

	keywords := make(map[string]interface{})
	keywords["ProjectID"] = 1
	robots, err := t.dao.ListRobotAccounts(&q.Query{
		Keywords: keywords,
	})
	t.require.Nil(err)
	t.require.Equal(3, len(robots))
}

func (t *robotAccountDaoTestSuite) TestUpdateRobotAccount() {
	robotName := "test4"
	robot := &model.Robot{
		Name:        robotName,
		Description: "test4 description",
		ProjectID:   1,
	}
	// add
	id, err := t.dao.CreateRobotAccount(robot)
	t.require.Nil(err)
	t.id4 = id
	// Disable
	robot.Disabled = true
	err = t.dao.UpdateRobotAccount(robot)
	t.require.Nil(err)
	// Get
	robot, err = t.dao.GetRobotAccount(id)
	t.require.Nil(err)
	t.require.Equal(true, robot.Disabled)
}

func (t *robotAccountDaoTestSuite) TestDeleteRobotAccount() {
	robotName := "test5"
	robot := &model.Robot{
		Name:        robotName,
		Description: "test5 description",
		ProjectID:   1,
	}
	// add
	id, err := t.dao.CreateRobotAccount(robot)
	t.require.Nil(err)
	// Disable
	err = t.dao.DeleteRobotAccount(id)
	t.require.Nil(err)
	// Get
	robot, err = t.dao.GetRobotAccount(id)
	t.require.Nil(err)
}

// TearDownSuite clears env for test suite
func (t *robotAccountDaoTestSuite) TearDownSuite() {
	err := t.dao.DeleteRobotAccount(t.id1)
	require.NoError(t.T(), err, "delete robot 1")

	err = t.dao.DeleteRobotAccount(t.id2)
	require.NoError(t.T(), err, "delete robot 2")

	err = t.dao.DeleteRobotAccount(t.id3)
	require.NoError(t.T(), err, "delete robot 3")

	err = t.dao.DeleteRobotAccount(t.id4)
	require.NoError(t.T(), err, "delete robot 4")
}

func TestRobotAccountDaoTestSuite(t *testing.T) {
	suite.Run(t, &robotAccountDaoTestSuite{})
}
