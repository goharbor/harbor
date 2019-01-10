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
	"fmt"
	"github.com/astaxie/beego/orm"
	"github.com/goharbor/harbor/src/common/models"
	"time"
)

// AddRobot ...
func AddRobot(robot *models.Robot) (int64, error) {
	now := time.Now()
	robot.CreationTime = now
	robot.UpdateTime = now
	return GetOrmer().Insert(robot)
}

// GetRobotByID ...
func GetRobotByID(id int64) (*models.Robot, error) {
	robot := &models.Robot{
		ID: id,
	}
	if err := GetOrmer().Read(robot); err != nil {
		if err == orm.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}

	return robot, nil
}

// ListRobots list robots according to the query conditions
func ListRobots(query *models.RobotQuery) ([]*models.Robot, error) {
	qs := getRobotQuerySetter(query)
	if query.Size > 0 {
		qs = qs.Limit(query.Size)
		if query.Page > 0 {
			qs = qs.Offset((query.Page - 1) * query.Size)
		}
	}
	qs = qs.OrderBy("Name")

	robots := []*models.Robot{}
	_, err := qs.All(&robots)
	return robots, err
}

func getRobotQuerySetter(query *models.RobotQuery) orm.QuerySeter {
	qs := GetOrmer().QueryTable(&models.Robot{})
	if len(query.Name) > 0 {
		qs = qs.Filter("Name", query.Name)
	}
	if query.ProjectID != 0 {
		qs = qs.Filter("ProjectID", query.ProjectID)
	}
	qs = qs.Filter("Disabled", false)
	return qs
}

// DisableRobot ...
func DisableRobot(id int64) error {
	robot, err := GetRobotByID(id)
	if err != nil {
		return err
	}
	robot.Name = fmt.Sprintf("%s#%d", robot.Name, robot.ID)
	robot.UpdateTime = time.Now()
	robot.Disabled = true
	_, err = GetOrmer().Update(robot, "Name", "UpdateTime", "Disabled")
	return err
}
