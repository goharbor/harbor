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

package scheduler

import (
	"fmt"
	"net/http"
	"sync"
	"time"

	chttp "github.com/goharbor/harbor/src/common/http"
	"github.com/goharbor/harbor/src/common/job"
	"github.com/goharbor/harbor/src/common/job/models"
	"github.com/goharbor/harbor/src/common/utils/log"
	"github.com/goharbor/harbor/src/core/config"
	"github.com/goharbor/harbor/src/pkg/scheduler/model"
	"github.com/pkg/errors"
)

const (
	jobParamCallbackFunc       = "callback_func"
	jobParamCallbackFuncParams = "params"
)

var (
	// GlobalScheduler is an instance of the default scheduler that
	// can be used globally. Call Init() to initialize it first
	GlobalScheduler Scheduler
	registry        = make(map[string]CallbackFunc)
)

// CallbackFunc defines the function that the scheduler calls when triggered
type CallbackFunc func(interface{}) error

// Scheduler provides the capability to run a periodic task, a callback function
// needs to be registered before using the scheduler
type Scheduler interface {
	Schedule(cron string, callbackFuncName string, params interface{}) (int64, error)
	UnSchedule(id int64) error
}

// Register the callback function with name, and the function will be called
// by the scheduler when the scheduler is triggered
func Register(name string, callbackFunc CallbackFunc) error {
	if len(name) == 0 {
		return errors.New("empty name")
	}
	if callbackFunc == nil {
		return errors.New("callback function is nil")
	}

	_, exist := registry[name]
	if exist {
		return fmt.Errorf("callback function %s already exists", name)
	}
	registry[name] = callbackFunc

	return nil
}

// GetCallbackFunc returns the registered callback function specified by the name
func GetCallbackFunc(name string) (CallbackFunc, error) {
	f, exist := registry[name]
	if !exist {
		return nil, fmt.Errorf("callback function %s not found", name)
	}
	return f, nil
}

func callbackFuncExist(name string) bool {
	_, exist := registry[name]
	return exist
}

// Init the GlobalScheduler
func Init() {
	GlobalScheduler = New(config.InternalCoreURL())
}

// New returns an instance of the default scheduler
func New(internalCoreURL string) Scheduler {
	return &scheduler{
		internalCoreURL:  internalCoreURL,
		jobserviceClient: job.GlobalClient,
		manager:          GlobalManager,
	}
}

type scheduler struct {
	sync.RWMutex
	internalCoreURL  string
	manager          Manager
	jobserviceClient job.Client
}

func (s *scheduler) Schedule(cron string, callbackFuncName string, params interface{}) (int64, error) {
	if !callbackFuncExist(callbackFuncName) {
		return 0, fmt.Errorf("callback function %s not found", callbackFuncName)
	}

	// create schedule record
	now := time.Now()
	scheduleID, err := s.manager.Create(&model.Schedule{
		CreationTime: &now,
		UpdateTime:   &now,
	})
	if err != nil {
		return 0, err
	}
	log.Debugf("the schedule record %d created", scheduleID)

	// submit scheduler job to Jobservice
	statusHookURL := fmt.Sprintf("%s/service/notifications/schedules/%d", s.internalCoreURL, scheduleID)
	jd := &models.JobData{
		Name: JobNameScheduler,
		Parameters: map[string]interface{}{
			jobParamCallbackFunc:       callbackFuncName,
			jobParamCallbackFuncParams: params,
		},
		Metadata: &models.JobMetadata{
			JobKind: job.JobKindPeriodic,
			Cron:    cron,
		},
		StatusHook: statusHookURL,
	}
	jobID, err := s.jobserviceClient.SubmitJob(jd)
	if err != nil {
		// if failed to submit to Jobservice, delete the schedule record in database
		e := s.manager.Delete(scheduleID)
		if e != nil {
			log.Errorf("failed to delete the schedule %d: %v", scheduleID, e)
		}
		return 0, err
	}
	log.Debugf("the scheduler job submitted to Jobservice, job ID: %s", jobID)

	// populate the job ID for the schedule
	err = s.manager.Update(&model.Schedule{
		ID:    scheduleID,
		JobID: jobID,
	}, "JobID")
	if err != nil {
		// stop the scheduler job
		if e := s.jobserviceClient.PostAction(jobID, job.JobActionStop); e != nil {
			log.Errorf("failed to stop the scheduler job %s: %v", jobID, e)
		}
		// delete the schedule record
		if e := s.manager.Delete(scheduleID); e != nil {
			log.Errorf("failed to delete the schedule record %d: %v", scheduleID, e)
		}
		return 0, err
	}

	return scheduleID, nil
}

func (s *scheduler) UnSchedule(id int64) error {
	schedule, err := s.manager.Get(id)
	if err != nil {
		return err
	}
	if schedule == nil {
		return fmt.Errorf("the schedule record %d not found", id)
	}
	if err = s.jobserviceClient.PostAction(schedule.JobID, job.JobActionStop); err != nil {
		herr, ok := err.(*chttp.Error)
		// if the job specified by jobID is not found in Jobservice, just delete
		// the schedule record
		if !ok || herr.Code != http.StatusNotFound {
			return err
		}
	}
	log.Debugf("the stop action for job %s submitted to the Jobservice", schedule.JobID)
	if err = s.manager.Delete(schedule.ID); err != nil {
		return err
	}
	log.Debugf("the schedule record %d deleted", schedule.ID)

	return nil
}
