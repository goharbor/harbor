package job

import (
	"fmt"
	"github.com/vmware/harbor/dao"
	"github.com/vmware/harbor/utils/log"
)

type StateHandler interface {
	Enter() error
	//Exit should be idempotent
	Exit() error
}

type DummyHandler struct {
	JobID int64
}

func (dh DummyHandler) Enter() error {
	return nil
}

func (dh DummyHandler) Exit() error {
	return nil
}

type StatusUpdater struct {
	DummyHandler
	State string
}

func (su StatusUpdater) Enter() error {
	err := dao.UpdateJobStatus(su.JobID, su.State)
	if err != nil {
		log.Warningf("Failed to update state of job: %d, state: %s, error: %v", su.JobID, su.State, err)
	}
	return err
}

type JobSM struct {
	JobID         int64
	CurrentState  string
	PreviousState string
	//The states that don't have to exist in transition map, such as "Error", "Canceled"
	ForcedStates map[string]struct{}
	Transitions  map[string]map[string]struct{}
	Handlers     map[string]StateHandler
}

func (sm *JobSM) EnterState(s string) error {
	log.Debugf("Trying to transit from State: %s, to State: %s", sm.CurrentState, s)
	targets, ok := sm.Transitions[sm.CurrentState]
	_, exist := targets[s]
	_, isForced := sm.ForcedStates[s]
	if !exist && !isForced {
		return fmt.Errorf("Transition from %s to %s does not exist!", sm.CurrentState, s)
	}
	exitHandler, ok := sm.Handlers[sm.CurrentState]
	if ok {
		if err := exitHandler.Exit(); err != nil {
			return err
		}
	} else {
		log.Debugf("No handler found for state:%s, skip", sm.CurrentState)
	}
	enterHandler, ok := sm.Handlers[s]
	if ok {
		if err := enterHandler.Enter(); err != nil {
			return err
		}
	} else {
		log.Debugf("No handler found for state:%s, skip", s)
	}
	sm.PreviousState = sm.CurrentState
	sm.CurrentState = s
	log.Debugf("Transition succeeded, current state: %s", s)
	return nil
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

func (sm *JobSM) InitJobSM() {
	sm.Handlers = make(map[string]StateHandler)
	sm.Transitions = make(map[string]map[string]struct{})
	sm.CurrentState = dao.JobPending
	log.Debugf("sm.Handlers: %v", sm.Handlers)
	sm.AddTransition(dao.JobPending, dao.JobRunning, StatusUpdater{DummyHandler{JobID: sm.JobID}, dao.JobRunning})
	sm.Handlers[dao.JobError] = StatusUpdater{DummyHandler{JobID: sm.JobID}, dao.JobError}
}
