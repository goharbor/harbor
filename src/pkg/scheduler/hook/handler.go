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

package hook

import (
	"time"

	"github.com/goharbor/harbor/src/pkg/scheduler"
	"github.com/goharbor/harbor/src/pkg/scheduler/model"
)

// GlobalController is an instance of the default controller that can be used globally
var GlobalController = NewController()

// Controller updates the scheduler job status or runs the callback function
type Controller interface {
	UpdateStatus(scheduleID int64, status string) error
	Run(callbackFuncName string, params interface{}) error
}

// NewController returns an instance of the default controller
func NewController() Controller {
	return &controller{
		manager: scheduler.GlobalManager,
	}
}

type controller struct {
	manager scheduler.Manager
}

func (c *controller) UpdateStatus(scheduleID int64, status string) error {
	now := time.Now()
	return c.manager.Update(&model.Schedule{
		ID:         scheduleID,
		Status:     status,
		UpdateTime: &now,
	}, "Status", "UpdateTime")
}

func (c *controller) Run(callbackFuncName string, params interface{}) error {
	f, err := scheduler.GetCallbackFunc(callbackFuncName)
	if err != nil {
		return err
	}
	return f(params)
}
