package job

import (
	"fmt"
	"github.com/vmware/harbor/dao"
	"github.com/vmware/harbor/utils/log"
	"sync"
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

type DummyHandler struct {
	JobID int64
}

func (dh DummyHandler) Enter() (string, error) {
	return "", nil
}

func (dh DummyHandler) Exit() error {
	return nil
}

type StatusUpdater struct {
	DummyHandler
	State string
}

func (su StatusUpdater) Enter() (string, error) {
	err := dao.UpdateJobStatus(su.JobID, su.State)
	if err != nil {
		log.Warningf("Failed to update state of job: %d, state: %s, error: %v", su.JobID, su.State, err)
	}
	var next string = JobContinue
	if su.State == JobStopped || su.State == JobError || su.State == JobFinished {
		next = ""
	}
	return next, err
}

const (
	JobPending  string = "pending"
	JobRunning  string = "running"
	JobError    string = "error"
	JobStopped  string = "stopped"
	JobFinished string = "finished"
	//  statemachine will move to next possible state based on trasition table
	JobContinue string = "_continue"
)

type JobSM struct {
	JobID         int64
	CurrentState  string
	PreviousState string
	//The states that don't have to exist in transition map, such as "Error", "Canceled"
	ForcedStates map[string]struct{}
	Transitions  map[string]map[string]struct{}
	Handlers     map[string]StateHandler
	lock         *sync.Mutex
	desiredState string
}

// EnsterState transit the statemachine from the current state to the state in parameter.
// It returns the next state the statemachine should tranit to.
func (sm *JobSM) EnterState(s string) (string, error) {
	log.Debugf("Trying to transit from State: %s, to State: %s", sm.CurrentState, s)
	targets, ok := sm.Transitions[sm.CurrentState]
	_, exist := targets[s]
	_, isForced := sm.ForcedStates[s]
	if !exist && !isForced {
		return "", fmt.Errorf("Transition from %s to %s does not exist!", sm.CurrentState, s)
	}
	exitHandler, ok := sm.Handlers[sm.CurrentState]
	if ok {
		if err := exitHandler.Exit(); err != nil {
			return "", err
		}
	} else {
		log.Debugf("No handler found for state:%s, skip", sm.CurrentState)
	}
	enterHandler, ok := sm.Handlers[s]
	var next string = JobContinue
	var err error
	if ok {
		if next, err = enterHandler.Enter(); err != nil {
			return "", err
		}
	} else {
		log.Debugf("No handler found for state:%s, skip", s)
	}
	sm.PreviousState = sm.CurrentState
	sm.CurrentState = s
	log.Debugf("Transition succeeded, current state: %s", s)
	return next, nil
}

// Start kicks off the statemachine to transit from current state to s, and moves on
// It will search the transit map if the next state is "_continue", and
// will enter error state if there's more than one possible path when next state is "_continue"
func (sm *JobSM) Start(s string) {
	n, err := sm.EnterState(s)
	log.Debugf("next state from handler: %s", n)
	for len(n) > 0 && err == nil {
		if d := sm.getDesiredState(); len(d) > 0 {
			log.Debugf("Desired state: %s, will ignore the next state from handler")
			n = d
			sm.setDesiredState("")
			continue
		}
		if n == JobContinue && len(sm.Transitions[sm.CurrentState]) == 1 {
			for n = range sm.Transitions[sm.CurrentState] {
				break
			}
			log.Debugf("Continue to state: %s", n)
			continue
		}
		if n == JobContinue && len(sm.Transitions[sm.CurrentState]) != 1 {
			log.Errorf("Next state is continue but there are %d possible next states in transition table", len(sm.Transitions[sm.CurrentState]))
			err = fmt.Errorf("Unable to continue")
			break
		}
		n, err = sm.EnterState(n)
		log.Debugf("next state from handler: %s", n)
	}
	if err != nil {
		log.Warningf("The statemachin will enter error state due to error: %v", err)
		sm.EnterState(JobError)
	}
}

func (sm *JobSM) AddTransition(from string, to string, h StateHandler) {
	_, ok := sm.Transitions[from]
	if !ok {
		sm.Transitions[from] = make(map[string]struct{})
	}
	sm.Transitions[from][to] = struct{}{}
	sm.Handlers[to] = h
}

func (sm *JobSM) RemoveTransition(from string, to string) {
	_, ok := sm.Transitions[from]
	if !ok {
		return
	}
	delete(sm.Transitions[from], to)
}

func (sm *JobSM) Stop() {
	sm.setDesiredState(JobStopped)
}

func (sm *JobSM) getDesiredState() string {
	sm.lock.Lock()
	defer sm.lock.Unlock()
	return sm.desiredState
}

func (sm *JobSM) setDesiredState(s string) {
	sm.lock.Lock()
	defer sm.lock.Unlock()
	sm.desiredState = s
}

func (sm *JobSM) InitJobSM() {
	sm.lock = &sync.Mutex{}
	sm.Handlers = make(map[string]StateHandler)
	sm.Transitions = make(map[string]map[string]struct{})
	sm.CurrentState = JobPending
	sm.AddTransition(JobPending, JobRunning, StatusUpdater{DummyHandler{JobID: sm.JobID}, JobRunning})
	sm.Handlers[JobError] = StatusUpdater{DummyHandler{JobID: sm.JobID}, JobError}
	sm.Handlers[JobStopped] = StatusUpdater{DummyHandler{JobID: sm.JobID}, JobStopped}
}
