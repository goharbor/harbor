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

package dao

import (
	"testing"

	"github.com/goharbor/harbor/src/common/models"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAddRobot(t *testing.T) {
	robotName := "test1"
	robot := &models.Robot{
		Name:        robotName,
		Token:       "rKgjKEMpMEK23zqejkWn5GIVvgJps1vKACTa6tnGXXyOlOTsXFESccDvgaJx047q1",
		Description: "test1 description",
		ProjectID:   1,
	}

	// add
	id, err := AddRobot(robot)
	require.Nil(t, err)
	robot.ID = id

	require.Nil(t, err)
	assert.NotNil(t, id)

}

func TestGetRobot(t *testing.T) {
	robotName := "test2"
	robot := &models.Robot{
		Name:        robotName,
		Token:       "rKgjKEMpMEK23zqejkWn5GIVvgJps1vKACTa6tnGXXyOlOTsXFESccDvgaJx047q2",
		Description: "test2 description",
		ProjectID:   1,
	}

	// add
	id, err := AddRobot(robot)
	require.Nil(t, err)
	robot.ID = id

	robot, err = GetRobotByID(id)
	require.Nil(t, err)
	assert.Equal(t, robotName, robot.Name)

}

func TestListRobots(t *testing.T) {
	robotName := "test3"
	robot := &models.Robot{
		Name:        robotName,
		Token:       "rKgjKEMpMEK23zqejkWn5GIVvgJps1vKACTa6tnGXXyOlOTsXFESccDvgaJx047q3",
		Description: "test3 description",
		ProjectID:   1,
	}

	_, err := AddRobot(robot)
	require.Nil(t, err)

	robots, err := ListRobots(&models.RobotQuery{
		ProjectID: 1,
	})
	require.Nil(t, err)
	assert.Equal(t, 3, len(robots))

}

func TestDisableRobot(t *testing.T) {
	robotName := "test4"
	robot := &models.Robot{
		Name:        robotName,
		Token:       "rKgjKEMpMEK23zqejkWn5GIVvgJps1vKACTa6tnGXXyOlOTsXFESccDvgaJx047q4",
		Description: "test4 description",
		ProjectID:   1,
	}

	// add
	id, err := AddRobot(robot)
	require.Nil(t, err)

	// Disable
	err = DisableRobot(id)
	require.Nil(t, err)

	// Get
	robot, err = GetRobotByID(id)
	require.Nil(t, err)
	assert.Equal(t, true, robot.Disabled)

}
