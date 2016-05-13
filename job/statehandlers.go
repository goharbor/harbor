package job

import (
	"time"

	"github.com/vmware/harbor/dao"
	"github.com/vmware/harbor/job/utils"
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
	err := dao.UpdateRepJobStatus(su.JobID, su.State)
	if err != nil {
		log.Warningf("Failed to update state of job: %d, state: %s, error: %v", su.JobID, su.State, err)
	}
	var next string = models.JobContinue
	if su.State == models.JobStopped || su.State == models.JobError || su.State == models.JobFinished {
		next = ""
	}
	return next, err
}

type ImgPuller struct {
	DummyHandler
	img    string
	logger utils.Logger
}

func (ip ImgPuller) Enter() (string, error) {
	ip.logger.Infof("I'm pretending to pull img:%s, then sleep 30s", ip.img)
	time.Sleep(30 * time.Second)
	ip.logger.Infof("wake up from sleep....")
	return "push-img", nil
}

type ImgPusher struct {
	DummyHandler
	targetURL string
	logger    utils.Logger
}

func (ip ImgPusher) Enter() (string, error) {
	ip.logger.Infof("I'm pretending to push img to:%s, then sleep 30s", ip.targetURL)
	time.Sleep(30 * time.Second)
	ip.logger.Infof("wake up from sleep....")
	return models.JobContinue, nil
}
