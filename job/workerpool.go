package job

import (
	"github.com/vmware/harbor/dao"
	"github.com/vmware/harbor/job/config"
	"github.com/vmware/harbor/models"
	"github.com/vmware/harbor/utils/log"
)

var WorkerPool chan *Worker

type Worker struct {
	ID      int
	RepJobs chan int64
	SM      *JobSM
	quit    chan bool
}

func (w *Worker) Start() {
	go func() {
		for {
			WorkerPool <- w
			select {
			case jobID := <-w.RepJobs:
				log.Debugf("worker: %d, will handle job: %d", w.ID, jobID)
				w.handleRepJob(jobID)
			case q := <-w.quit:
				if q {
					log.Debugf("worker: %d, will stop.", w.ID)
					return
				}
			}
		}
	}()
}

func (w *Worker) Stop() {
	go func() {
		w.quit <- true
	}()
}

func (w *Worker) handleRepJob(id int64) {
	err := w.SM.Reset(id)
	if err != nil {
		log.Errorf("Worker %d, failed to re-initialize statemachine, error: %v", w.ID, err)
		err2 := dao.UpdateRepJobStatus(id, models.JobError)
		if err2 != nil {
			log.Errorf("Failed to update job status to ERROR, error:%v", err2)
		}
		return
	}
	if w.SM.Parms.Enabled == 0 {
		log.Debugf("The policy of job:%d is disabled, will cancel the job")
		_ = dao.UpdateRepJobStatus(id, models.JobCanceled)
	} else {
		w.SM.Start(models.JobRunning)
	}
}

func NewWorker(id int) *Worker {
	w := &Worker{
		ID:      id,
		RepJobs: make(chan int64),
		quit:    make(chan bool),
		SM:      &JobSM{},
	}
	w.SM.Init()
	return w
}

func InitWorkerPool() {
	WorkerPool = make(chan *Worker, config.MaxJobWorkers())
	for i := 0; i < config.MaxJobWorkers(); i++ {
		worker := NewWorker(i)
		worker.Start()
		log.Debugf("worker %d started", worker.ID)
	}
}

func Dispatch() {
	for {
		select {
		case job := <-JobQueue:
			go func(jobID int64) {
				log.Debugf("Trying to dispatch job: %d", jobID)
				worker := <-WorkerPool
				worker.RepJobs <- jobID
			}(job)
		}
	}
}
