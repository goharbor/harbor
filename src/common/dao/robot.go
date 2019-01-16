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
	"github.com/astaxie/beego/orm"
	"github.com/goharbor/harbor/src/common/models"
	"strings"
	"time"
)

// AddRobot ...
func AddRobot(robot *models.Robot) (int64, error) {
	now := time.Now()
	robot.CreationTime = now
	robot.UpdateTime = now
	id, err := GetOrmer().Insert(robot)
	if err != nil {
		if strings.Contains(err.Error(), "duplicate key value violates unique constraint") {
			return 0, ErrDupRows
		}
		return 0, err
	}
	return id, nil
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
	qs := getRobotQuerySetter(query).OrderBy("Name")
	if query != nil {
		if query.Size > 0 {
			qs = qs.Limit(query.Size)
			if query.Page > 0 {
				qs = qs.Offset((query.Page - 1) * query.Size)
			}
		}
	}
	robots := []*models.Robot{}
	_, err := qs.All(&robots)
	return robots, err
}

func getRobotQuerySetter(query *models.RobotQuery) orm.QuerySeter {
	qs := GetOrmer().QueryTable(&models.Robot{})

	if query == nil {
		return qs
	}

	if len(query.Name) > 0 {
		if query.FuzzyMatchName {
			qs = qs.Filter("Name__icontains", query.Name)
		} else {
			qs = qs.Filter("Name", query.Name)
		}
	}
	if query.ProjectID != 0 {
		qs = qs.Filter("ProjectID", query.ProjectID)
	}
	return qs
}

// CountRobot ...
func CountRobot(query *models.RobotQuery) (int64, error) {
	return getRobotQuerySetter(query).Count()
}

// UpdateRobot ...
func UpdateRobot(robot *models.Robot) error {
	robot.UpdateTime = time.Now()
	_, err := GetOrmer().Update(robot)
	return err
}

// DeleteRobot ...
func DeleteRobot(id int64) error {
	_, err := GetOrmer().QueryTable(&models.Robot{}).Filter("ID", id).Delete()
	return err
}
