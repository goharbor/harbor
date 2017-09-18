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
)

// workerPool is a set of workers each worker is associate to a statemachine for handling jobs.
// it consists of a channel for free workers and a list to all workers
type workerPool struct {
	poolType   Type
	workerChan chan *Worker
	workerList []*Worker
}

// WorkerPools is a map contains workerpools for different types of jobs.
var WorkerPools map[Type]*workerPool

// For WorkerPools initialization.
var once sync.Once

//TODO: remove the hard code?
const maxScanWorker = 3

// StopJobs accepts a list of jobs and will try to stop them if any of them is being executed by the worker.
func (wp *workerPool) StopJobs(jobs []Job) {
	log.Debugf("Works working on jobs: %v will be stopped", jobs)
	for _, j := range jobs {
		for _, w := range wp.workerList {
			if w.SM.CurrentJob.ID() == j.ID() {
				log.Debugf("found a worker whose job ID is %d, type: %v, will try to stop it", j.ID(), j.Type())
				w.SM.Stop(j)
			}
		}
	}
}

// Worker consists of a channel for job from which worker gets the next job to handle, and a pointer to a statemachine,
// the actual work to handle the job is done via state machine.
type Worker struct {
	ID    int
	Type  Type
	Jobs  chan Job
	queue chan *Worker
	SM    *SM
	quit  chan bool
}

// String ...
func (w *Worker) String() string {
	return fmt.Sprintf("{ID: %d, Type: %v}", w.ID, w.Type)
}

// Start is a loop worker gets id from its channel and handle it.
func (w *Worker) Start() {
	go func() {
		for {
			w.queue <- w
			select {
			case job := <-w.Jobs:
				log.Debugf("worker: %v, will handle job: %v", w, job)
				w.handle(job)
			case q := <-w.quit:
				if q {
					log.Debugf("worker: %v, will stop.", w)
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

func (w *Worker) handle(job Job) {
	err := w.SM.Reset(job)
	if err != nil {
		log.Errorf("Worker %v, failed to re-initialize statemachine for job: %v, error: %v", w, job, err)
		err2 := job.UpdateStatus(models.JobError)
		if err2 != nil {
			log.Errorf("Failed to update job status to ERROR, job: %v, error:%v", job, err2)
		}
	}
}

// NewWorker returns a pointer to new instance of worker
func NewWorker(id int, t Type, wp *workerPool) *Worker {
	w := &Worker{
		ID:    id,
		Type:  t,
		Jobs:  make(chan Job),
		quit:  make(chan bool),
		queue: wp.workerChan,
		SM:    &SM{},
	}
	w.SM.Init()
	return w
}

// InitWorkerPools create worker pools for different types of jobs.
func InitWorkerPools() error {
	maxRepWorker, err := config.MaxJobWorkers()
	if err != nil {
		return err
	}
	once.Do(func() {
		WorkerPools = make(map[Type]*workerPool)
		WorkerPools[ReplicationType] = createWorkerPool(maxRepWorker, ReplicationType)
		WorkerPools[ScanType] = createWorkerPool(maxScanWorker, ScanType)
	})
	return nil
}

//createWorkerPool create workers according to parm
func createWorkerPool(n int, t Type) *workerPool {
	wp := &workerPool{
		workerChan: make(chan *Worker, n),
		workerList: make([]*Worker, 0, n),
	}
	for i := 0; i < n; i++ {
		worker := NewWorker(i, t, wp)
		wp.workerList = append(wp.workerList, worker)
		worker.Start()
		log.Debugf("worker %v started", worker)
	}
	return wp
}

// Dispatch will listen to the jobQueue of job service and try to pick a free worker from the worker pool and assign the job to it.
func Dispatch() {
	for {
		select {
		case job := <-jobQueue:
			go func(job Job) {
				log.Debugf("Trying to dispatch job: %v", job)
				worker := <-WorkerPools[job.Type()].workerChan
				worker.Jobs <- job
			}(job)
		}
	}
}
