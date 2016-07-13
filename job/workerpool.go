/*
   Copyright (c) 2016 VMware, Inc. All Rights Reserved.
   Licensed under the Apache License, Version 2.0 (the "License");
   you may not use this file except in compliance with the License.
   You may obtain a copy of the License at

       http://www.apache.org/licenses/LICENSE-2.0

   Unless required by applicable law or agreed to in writing, software
   distributed under the License is distributed on an "AS IS" BASIS,
   WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
   See the License for the specific language governing permissions and
   limitations under the License.
*/

package job

import (
	"github.com/vmware/harbor/dao"
	"github.com/vmware/harbor/job/config"
	"github.com/vmware/harbor/models"
	"github.com/vmware/harbor/utils/log"
)

type workerPool struct {
	workerChan chan *Worker
	workerList []*Worker
}

// WorkerPool is a set of workers each worker is associate to a statemachine for handling jobs.
// it consists of a channel for free workers and a list to all workers
var WorkerPool *workerPool

// StopJobs accepts a list of jobs and will try to stop them if any of them is being executed by the worker.
func (wp *workerPool) StopJobs(jobs []int64) {
	log.Debugf("Works working on jobs: %v will be stopped", jobs)
	for _, id := range jobs {
		for _, w := range wp.workerList {
			if w.SM.JobID == id {
				log.Debugf("found a worker whose job ID is %d, will try to stop it", id)
				w.SM.Stop(id)
			}
		}
	}
}

// Worker consists of a channel for job from which worker gets the next job to handle, and a pointer to a statemachine,
// the actual work to handle the job is done via state machine.
type Worker struct {
	ID      int
	RepJobs chan int64
	SM      *SM
	quit    chan bool
}

// Start is a loop worker gets id from its channel and handle it.
func (w *Worker) Start() {
	go func() {
		for {
			WorkerPool.workerChan <- w
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

// Stop ...
func (w *Worker) Stop() {
	go func() {
		w.quit <- true
	}()
}

func (w *Worker) handleRepJob(id int64) {
	err := w.SM.Reset(id)
	if err != nil {
		log.Errorf("Worker %d, failed to re-initialize statemachine for job: %d, error: %v", w.ID, id, err)
		err2 := dao.UpdateRepJobStatus(id, models.JobError)
		if err2 != nil {
			log.Errorf("Failed to update job status to ERROR, job: %d, error:%v", id, err2)
		}
		return
	}
	if w.SM.Parms.Enabled == 0 {
		log.Debugf("The policy of job:%d is disabled, will cancel the job", id)
		_ = dao.UpdateRepJobStatus(id, models.JobCanceled)
		w.SM.Logger.Info("The job has been canceled")
	} else {
		w.SM.Start(models.JobRunning)
	}
}

// NewWorker returns a pointer to new instance of worker
func NewWorker(id int) *Worker {
	w := &Worker{
		ID:      id,
		RepJobs: make(chan int64),
		quit:    make(chan bool),
		SM:      &SM{},
	}
	w.SM.Init()
	return w
}

// InitWorkerPool create workers according to configuration.
func InitWorkerPool() {
	WorkerPool = &workerPool{
		workerChan: make(chan *Worker, config.MaxJobWorkers()),
		workerList: make([]*Worker, 0, config.MaxJobWorkers()),
	}
	for i := 0; i < config.MaxJobWorkers(); i++ {
		worker := NewWorker(i)
		WorkerPool.workerList = append(WorkerPool.workerList, worker)
		worker.Start()
		log.Debugf("worker %d started", worker.ID)
	}
}

// Dispatch will listen to the jobQueue of job service and try to pick a free worker from the worker pool and assign the job to it.
func Dispatch() {
	for {
		select {
		case job := <-jobQueue:
			go func(jobID int64) {
				log.Debugf("Trying to dispatch job: %d", jobID)
				worker := <-WorkerPool.workerChan
				worker.RepJobs <- jobID
			}(job)
		}
	}
}
