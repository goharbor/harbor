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
	"fmt"
	"sync"

	"github.com/vmware/harbor/src/common/models"
	"github.com/vmware/harbor/src/common/utils/log"
	"github.com/vmware/harbor/src/jobservice/config"
	"github.com/vmware/harbor/src/jobservice/replication"
	"github.com/vmware/harbor/src/jobservice/scan"
)

// SM is the state machine to handle job, it handles one job at a time.
type SM struct {
	CurrentJob    Job
	CurrentState  string
	PreviousState string
	//The states that don't have to exist in transition map, such as "Error", "Canceled"
	ForcedStates map[string]struct{}
	Transitions  map[string]map[string]struct{}
	Handlers     map[string]StateHandler
	desiredState string
	Logger       *log.Logger
	lock         *sync.Mutex
}

// EnterState transit the statemachine from the current state to the state in parameter.
// It returns the next state the statemachine should tranit to.
func (sm *SM) EnterState(s string) (string, error) {
	log.Debugf("Job: %v, transiting from State: %s, to State: %s", sm.CurrentJob, sm.CurrentState, s)
	targets, ok := sm.Transitions[sm.CurrentState]
	_, exist := targets[s]
	_, isForced := sm.ForcedStates[s]
	if !exist && !isForced {
		return "", fmt.Errorf("job: %v, transition from %s to %s does not exist", sm.CurrentJob, sm.CurrentState, s)
	}
	exitHandler, ok := sm.Handlers[sm.CurrentState]
	if ok {
		if err := exitHandler.Exit(); err != nil {
			return "", err
		}
	} else {
		log.Debugf("Job: %v, no exit handler found for state:%s, skip", sm.CurrentJob, sm.CurrentState)
	}
	enterHandler, ok := sm.Handlers[s]
	var next = models.JobContinue
	var err error
	if ok {
		if next, err = enterHandler.Enter(); err != nil {
			return "", err
		}
	} else {
		log.Debugf("Job: %v, no handler found for state:%s, skip", sm.CurrentJob, s)
	}
	sm.PreviousState = sm.CurrentState
	sm.CurrentState = s
	log.Debugf("Job: %v, transition succeeded, current state: %s", sm.CurrentJob, s)
	return next, nil
}

// Start kicks off the statemachine to transit from current state to s, and moves on
// It will search the transit map if the next state is "_continue", and
// will enter error state if there's more than one possible path when next state is "_continue"
func (sm *SM) Start(s string) {
	n, err := sm.EnterState(s)
	log.Debugf("Job: %v, next state from handler: %s", sm.CurrentJob, n)
	for len(n) > 0 && err == nil {
		if d := sm.getDesiredState(); len(d) > 0 {
			log.Debugf("Job: %v, Desired state: %s, will ignore the next state from handler", sm.CurrentJob, d)
			n = d
			sm.setDesiredState("")
			continue
		}
		if n == models.JobContinue && len(sm.Transitions[sm.CurrentState]) == 1 {
			for n = range sm.Transitions[sm.CurrentState] {
				break
			}
			log.Debugf("Job: %v, Continue to state: %s", sm.CurrentJob, n)
			continue
		}
		if n == models.JobContinue && len(sm.Transitions[sm.CurrentState]) != 1 {
			log.Errorf("Job: %v, next state is continue but there are %d possible next states in transition table", sm.CurrentJob, len(sm.Transitions[sm.CurrentState]))
			err = fmt.Errorf("Unable to continue")
			break
		}
		n, err = sm.EnterState(n)
		log.Debugf("Job: %v, next state from handler: %s", sm.CurrentJob, n)
	}
	if err != nil {
		log.Warningf("Job: %v, the statemachine will enter error state due to error: %v", sm.CurrentJob, err)
		sm.EnterState(models.JobError)
	}
}

// AddTransition add a transition to the transition table of state machine, the handler is the handler of target state "to"
func (sm *SM) AddTransition(from string, to string, h StateHandler) {
	_, ok := sm.Transitions[from]
	if !ok {
		sm.Transitions[from] = make(map[string]struct{})
	}
	sm.Transitions[from][to] = struct{}{}
	sm.Handlers[to] = h
}

// RemoveTransition removes a transition from transition table of the state machine
func (sm *SM) RemoveTransition(from string, to string) {
	_, ok := sm.Transitions[from]
	if !ok {
		return
	}
	delete(sm.Transitions[from], to)
}

// Stop will set the desired state as "stopped" such that when next tranisition happen the state machine will stop handling the current job
// and the worker can release itself to the workerpool.
func (sm *SM) Stop(job Job) {
	log.Debugf("Trying to stop the job: %v", job)
	sm.lock.Lock()
	defer sm.lock.Unlock()
	//need to check if the sm switched to other job
	if job.ID() == sm.CurrentJob.ID() && job.Type() == sm.CurrentJob.Type() {
		sm.desiredState = models.JobStopped
		log.Debugf("Desired state of job %v is set to stopped", job)
	} else {
		log.Debugf("State machine has switched to job %v, so the action to stop job %v will be ignored", sm.CurrentJob, job)
	}
}

func (sm *SM) getDesiredState() string {
	sm.lock.Lock()
	defer sm.lock.Unlock()
	return sm.desiredState
}

func (sm *SM) setDesiredState(s string) {
	sm.lock.Lock()
	defer sm.lock.Unlock()
	sm.desiredState = s
}

// Init initialzie the state machine, it will be called once in the lifecycle of state machine.
func (sm *SM) Init() {
	sm.lock = &sync.Mutex{}
	sm.Handlers = make(map[string]StateHandler)
	sm.Transitions = make(map[string]map[string]struct{})
	sm.ForcedStates = map[string]struct{}{
		models.JobError:    struct{}{},
		models.JobStopped:  struct{}{},
		models.JobCanceled: struct{}{},
		models.JobRetrying: struct{}{},
	}
}

// Reset resets the state machine and after prereq checking, it will start handling the job.
func (sm *SM) Reset(j Job) error {
	//To ensure the Job visible to the thread to stop the SM
	sm.lock.Lock()
	sm.CurrentJob = j
	sm.desiredState = ""
	sm.lock.Unlock()

	var err error
	sm.Logger, err = NewLogger(j)
	if err != nil {
		return err
	}
	//init states handlers
	sm.Handlers = make(map[string]StateHandler)
	sm.Transitions = make(map[string]map[string]struct{})
	sm.CurrentState = models.JobPending

	sm.AddTransition(models.JobPending, models.JobRunning, StatusUpdater{sm.CurrentJob, models.JobRunning})
	sm.AddTransition(models.JobRetrying, models.JobRunning, StatusUpdater{sm.CurrentJob, models.JobRunning})
	sm.Handlers[models.JobError] = StatusUpdater{sm.CurrentJob, models.JobError}
	sm.Handlers[models.JobStopped] = StatusUpdater{sm.CurrentJob, models.JobStopped}
	sm.Handlers[models.JobCanceled] = StatusUpdater{sm.CurrentJob, models.JobCanceled}
	sm.Handlers[models.JobRetrying] = Retry{sm.CurrentJob}
	if err := sm.CurrentJob.Init(); err != nil {
		return err
	}
	if err := sm.initTransitions(); err != nil {
		return err
	}
	return sm.kickOff()
}

func (sm *SM) kickOff() error {
	if repJob, ok := sm.CurrentJob.(*RepJob); ok {
		if repJob.parm.Enabled == 0 {
			log.Debugf("The policy of job:%v is disabled, will cancel the job", repJob)
			_, err := sm.EnterState(models.JobCanceled)
			if err != nil {
				log.Warningf("For job: %v, failed to update state to 'canceled', error: %v", repJob, err)
			}
			return err
		}
	}
	log.Debugf("In kickOff: will start job: %v", sm.CurrentJob)
	sm.Start(models.JobRunning)
	return nil
}

func (sm *SM) initTransitions() error {
	switch sm.CurrentJob.Type() {
	case ReplicationType:
		repJob, ok := sm.CurrentJob.(*RepJob)
		if !ok {
			//Shouldn't be here.
			return fmt.Errorf("The job: %v is not a type of RepJob", sm.CurrentJob)
		}
		jobParm := repJob.parm
		if jobParm.Operation == models.RepOpTransfer {
			addImgTransferTransition(sm, jobParm)
		} else if jobParm.Operation == models.RepOpDelete {
			addImgDeleteTransition(sm, jobParm)
		} else {
			return fmt.Errorf("unsupported operation: %s", jobParm.Operation)
		}
	case ScanType:
		scanJob, ok := sm.CurrentJob.(*ScanJob)
		if !ok {
			//Shouldn't be here.
			return fmt.Errorf("The job: %v is not a type of ScanJob", sm.CurrentJob)
		}
		addImgScanTransition(sm, scanJob.parm)
		return nil
	default:
		return fmt.Errorf("Unsupported job type: %v", sm.CurrentJob.Type())
	}
	return nil
}

//for testing onlly
/*
func addTestTransition(sm *SM) error {
	sm.AddTransition(models.JobRunning, "pull-img", ImgPuller{img: sm.Parms.Repository, logger: sm.Logger})
	return nil
}
*/

func addImgScanTransition(sm *SM, parm *ScanJobParm) {
	ctx := &scan.JobContext{
		Repository: parm.Repository,
		Tag:        parm.Tag,
		Digest:     parm.Digest,
		JobID:      sm.CurrentJob.ID(),
		Logger:     sm.Logger,
	}

	layerScanHandler := &scan.LayerScanHandler{Context: ctx}
	sm.AddTransition(models.JobRunning, scan.StateInitialize, &scan.Initializer{Context: ctx})
	sm.AddTransition(scan.StateInitialize, scan.StateScanLayer, layerScanHandler)
	sm.AddTransition(scan.StateScanLayer, scan.StateScanLayer, layerScanHandler)
	sm.AddTransition(scan.StateScanLayer, scan.StateSummarize, &scan.SummarizeHandler{Context: ctx})
	sm.AddTransition(scan.StateSummarize, models.JobFinished, &StatusUpdater{sm.CurrentJob, models.JobFinished})
}

func addImgTransferTransition(sm *SM, parm *RepJobParm) {
	base := replication.InitBaseHandler(parm.Repository, parm.LocalRegURL, config.JobserviceSecret(),
		parm.TargetURL, parm.TargetUsername, parm.TargetPassword,
		parm.Insecure, parm.Tags, sm.Logger)

	sm.AddTransition(models.JobRunning, replication.StateInitialize, &replication.Initializer{BaseHandler: base})
	sm.AddTransition(replication.StateInitialize, replication.StateCheck, &replication.Checker{BaseHandler: base})
	sm.AddTransition(replication.StateCheck, replication.StatePullManifest, &replication.ManifestPuller{BaseHandler: base})
	sm.AddTransition(replication.StatePullManifest, replication.StateTransferBlob, &replication.BlobTransfer{BaseHandler: base})
	sm.AddTransition(replication.StatePullManifest, models.JobFinished, &StatusUpdater{sm.CurrentJob, models.JobFinished})
	sm.AddTransition(replication.StateTransferBlob, replication.StatePushManifest, &replication.ManifestPusher{BaseHandler: base})
	sm.AddTransition(replication.StatePushManifest, replication.StatePullManifest, &replication.ManifestPuller{BaseHandler: base})
}

func addImgDeleteTransition(sm *SM, parm *RepJobParm) {
	deleter := replication.NewDeleter(parm.Repository, parm.Tags, parm.TargetURL,
		parm.TargetUsername, parm.TargetPassword, parm.Insecure, sm.Logger)

	sm.AddTransition(models.JobRunning, replication.StateDelete, deleter)
	sm.AddTransition(replication.StateDelete, models.JobFinished, &StatusUpdater{sm.CurrentJob, models.JobFinished})
}
