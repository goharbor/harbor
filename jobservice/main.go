package main

import (
	"github.com/astaxie/beego"
	"github.com/vmware/harbor/dao"
	"github.com/vmware/harbor/job"
	"github.com/vmware/harbor/utils/log"
	"os"
	"strconv"
)

const defaultMaxWorkers int = 10

type Worker struct {
	ID      int
	RepJobs chan int64
	quit    chan bool
}

func (w *Worker) Start() {
	go func() {
		for {
			WorkerPool <- w
			select {
			case jobID := <-w.RepJobs:
				log.Debugf("worker: %d, will handle job: %d", w.ID, jobID)
				job.HandleRepJob(jobID)
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

var WorkerPool chan *Worker

func main() {
	dao.InitDB()
	initRouters()
	initWorkerPool()
	go dispatch()
	beego.Run()
}

func initWorkerPool() {
	maxWorkersEnv := os.Getenv("MAX_JOB_WORKERS")
	maxWorkers64, err := strconv.ParseInt(maxWorkersEnv, 10, 32)
	maxWorkers := int(maxWorkers64)
	if err != nil {
		log.Warningf("Failed to parse max works setting, error: %v, the default value: %d will be used", err, defaultMaxWorkers)
		maxWorkers = defaultMaxWorkers
	}
	WorkerPool = make(chan *Worker, maxWorkers)
	for i := 0; i < maxWorkers; i++ {
		worker := &Worker{
			ID:      i,
			RepJobs: make(chan int64),
			quit:    make(chan bool),
		}
		worker.Start()
	}
}

func dispatch() {
	for {
		select {
		case job := <-job.JobQueue:
			go func(jobID int64) {
				log.Debugf("Trying to dispatch job: %d", jobID)
				worker := <-WorkerPool
				worker.RepJobs <- jobID
			}(job)
		}
	}
}
