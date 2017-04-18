// Copyright (c) 2017 VMware, Inc. All Rights Reserved.
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

package job

import (
	"time"

	"github.com/vmware/harbor/src/common/dao"
	"github.com/vmware/harbor/src/common/models"
	"github.com/vmware/harbor/src/common/utils/log"
)

// StateHandler handles transition, it associates with each state, will be called when
// SM enters and exits a state during a transition.
type StateHandler interface {
	// Enter returns the next state, if it returns empty string the SM will hold the current state or
	// or decide the next state.
	Enter() (string, error)
	//Exit should be idempotent
	Exit() error
}

// StatusUpdater implements the StateHandler interface which updates the status of a job in DB when the job enters
// a status.
type StatusUpdater struct {
	JobID int64
	State string
}

// Enter updates the status of a job and returns "_continue" status to tell state machine to move on.
// If the status is a final status it returns empty string and the state machine will be stopped.
func (su StatusUpdater) Enter() (string, error) {
	err := dao.UpdateRepJobStatus(su.JobID, su.State)
	if err != nil {
		log.Warningf("Failed to update state of job: %d, state: %s, error: %v", su.JobID, su.State, err)
	}
	var next = models.JobContinue
	if su.State == models.JobStopped || su.State == models.JobError || su.State == models.JobFinished {
		next = ""
	}
	return next, err
}

// Exit ...
func (su StatusUpdater) Exit() error {
	return nil
}

// Retry handles a special "retrying" in which case it will update the status in DB and reschedule the job
// via scheduler
type Retry struct {
	JobID int64
}

// Enter ...
func (jr Retry) Enter() (string, error) {
	err := dao.UpdateRepJobStatus(jr.JobID, models.JobRetrying)
	if err != nil {
		log.Errorf("Failed to update state of job :%d to Retrying, error: %v", jr.JobID, err)
	}
	go Reschedule(jr.JobID)
	return "", err
}

// Exit ...
func (jr Retry) Exit() error {
	return nil
}

// ImgPuller was for testing
type ImgPuller struct {
	img    string
	logger *log.Logger
}

// Enter ...
func (ip ImgPuller) Enter() (string, error) {
	ip.logger.Infof("I'm pretending to pull img:%s, then sleep 30s", ip.img)
	time.Sleep(30 * time.Second)
	ip.logger.Infof("wake up from sleep.... testing retry")
	return models.JobRetrying, nil
}

// Exit ...
func (ip ImgPuller) Exit() error {
	return nil
}

// ImgPusher is a statehandler for testing
type ImgPusher struct {
	targetURL string
	logger    *log.Logger
}

// Enter ...
func (ip ImgPusher) Enter() (string, error) {
	ip.logger.Infof("I'm pretending to push img to:%s, then sleep 30s", ip.targetURL)
	time.Sleep(30 * time.Second)
	ip.logger.Infof("wake up from sleep.... testing retry")
	return models.JobRetrying, nil
}

// Exit ...
func (ip ImgPusher) Exit() error {
	return nil
}
