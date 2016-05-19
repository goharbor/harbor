package replication

/*
import (
	"encoding/json"
	//"github.com/vmware/harbor/dao"
	"github.com/vmware/harbor/job"
	"github.com/vmware/harbor/models"
	"time"
)

const (
	jobType = "transfer_img_out"
)

type Runner struct {
	job.JobSM
	Logger job.Logger
	parm   ImgOutParm
}

type ImgPuller struct {
	job.DummyHandler
	img    string
	logger job.Logger
}

func (ip ImgPuller) Enter() (string, error) {
	ip.logger.Infof("I'm pretending to pull img:%s, then sleep 30s", ip.img)
	time.Sleep(30 * time.Second)
	ip.logger.Infof("wake up from sleep....")
	return "push-img", nil
}

type ImgPusher struct {
	job.DummyHandler
	targetURL string
	logger    job.Logger
}

func (ip ImgPusher) Enter() (string, error) {
	ip.logger.Infof("I'm pretending to push img to:%s, then sleep 30s", ip.targetURL)
	time.Sleep(30 * time.Second)
	ip.logger.Infof("wake up from sleep....")
	return job.JobContinue, nil
}

func init() {
	job.Register(jobType, Runner{})
}

func (r Runner) Run(je models.JobEntry) error {
	err := r.init(je)
	if err != nil {
		return err
	}
	r.Start(job.JobRunning)
	return nil
}

func (r *Runner) init(je models.JobEntry) error {
	r.JobID = je.ID
	r.InitJobSM()
	err := json.Unmarshal([]byte(je.ParmsStr), &r.parm)
	if err != nil {
		return err
	}
	r.Logger = job.Logger{je.ID}
	r.AddTransition(job.JobRunning, "pull-img", ImgPuller{DummyHandler: job.DummyHandler{JobID: r.JobID}, img: r.parm.Image, logger: r.Logger})
	//only handle on target for now
	url := r.parm.Targets[0].URL
	r.AddTransition("pull-img", "push-img", ImgPusher{DummyHandler: job.DummyHandler{JobID: r.JobID}, targetURL: url, logger: r.Logger})
	r.AddTransition("push-img", job.JobFinished, job.StatusUpdater{job.DummyHandler{JobID: r.JobID}, job.JobFinished})
	return nil
}
*/
