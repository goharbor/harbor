//  Copyright Project Harbor Authors
//
//  Licensed under the Apache License, Version 2.0 (the "License");
//  you may not use this file except in compliance with the License.
//  You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
//  Unless required by applicable law or agreed to in writing, software
//  distributed under the License is distributed on an "AS IS" BASIS,
//  WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
//  See the License for the specific language governing permissions and
//  limitations under the License.

package jobservice

import (
	"context"
	"github.com/goharbor/harbor/src/lib/errors"
	"github.com/goharbor/harbor/src/lib/q"
	"github.com/goharbor/harbor/src/pkg/scheduler"
)

var (
	// SchedulerCtl ...
	SchedulerCtl = NewSchedulerCtrl()
)

// SchedulerController interface to manage schedule
type SchedulerController interface {
	// Get the schedule
	Get(ctx context.Context, vendorType string) (*scheduler.Schedule, error)
	// Create with cron type & string
	Create(ctx context.Context, vendorType, cronType, cron, callbackFuncName string, policy interface{}, extrasParam map[string]interface{}) (int64, error)
	// Delete the schedule
	Delete(ctx context.Context, vendorType string) error
}

type schedulerController struct {
	schedulerMgr scheduler.Scheduler
}

// NewSchedulerCtrl ...
func NewSchedulerCtrl() SchedulerController {
	return &schedulerController{schedulerMgr: scheduler.New()}
}
func (s schedulerController) Get(ctx context.Context, vendorType string) (*scheduler.Schedule, error) {
	sch, err := s.schedulerMgr.ListSchedules(ctx, q.New(q.KeyWords{"VendorType": vendorType}))
	if err != nil {
		return nil, err
	}
	if len(sch) == 0 {
		return nil, errors.New(nil).WithCode(errors.NotFoundCode).WithMessage("no schedule is found")
	}
	if sch[0] == nil {
		return nil, errors.New(nil).WithCode(errors.NotFoundCode).WithMessage("no schedule is found")
	}
	return sch[0], nil
}

func (s schedulerController) Create(ctx context.Context, vendorType, cronType, cron, callbackFuncName string,
	policy interface{}, extrasParam map[string]interface{}) (int64, error) {
	return s.schedulerMgr.Schedule(ctx, vendorType, -1, cronType, cron, callbackFuncName, policy, extrasParam)
}

func (s schedulerController) Delete(ctx context.Context, vendorType string) error {
	return s.schedulerMgr.UnScheduleByVendor(ctx, vendorType, -1)
}
