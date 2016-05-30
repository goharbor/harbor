package job

import (
	"time"

	"github.com/vmware/harbor/dao"
	"github.com/vmware/harbor/models"
	"github.com/vmware/harbor/utils/log"
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

// DummyHandler is the default implementation of StateHander interface, which has empty Enter and Exit methods.
type DummyHandler struct {
	JobID int64
}

// Enter ...
func (dh DummyHandler) Enter() (string, error) {
	return "", nil
}

// Exit ...
func (dh DummyHandler) Exit() error {
	return nil
}

// StatusUpdater implements the StateHandler interface which updates the status of a job in DB when the job enters
// a status.
type StatusUpdater struct {
	DummyHandler
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

// ImgPuller was for testing
type ImgPuller struct {
	DummyHandler
	img    string
	logger *log.Logger
}

// Enter ...
func (ip ImgPuller) Enter() (string, error) {
	ip.logger.Infof("I'm pretending to pull img:%s, then sleep 30s", ip.img)
	time.Sleep(30 * time.Second)
	ip.logger.Infof("wake up from sleep....")
	return "push-img", nil
}

// ImgPusher is a statehandler for testing
type ImgPusher struct {
	DummyHandler
	targetURL string
	logger    *log.Logger
}

// Enter ...
func (ip ImgPusher) Enter() (string, error) {
	ip.logger.Infof("I'm pretending to push img to:%s, then sleep 30s", ip.targetURL)
	time.Sleep(30 * time.Second)
	ip.logger.Infof("wake up from sleep....")
	return models.JobContinue, nil
}
