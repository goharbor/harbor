// Copyright Project Harbor Authors
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

package handler

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/goharbor/harbor/src/lib/errors"
	"github.com/goharbor/harbor/src/pkg/scheduler"

	"github.com/go-openapi/runtime/middleware"
	"github.com/go-openapi/strfmt"

	"github.com/goharbor/harbor/src/common/rbac"
	"github.com/goharbor/harbor/src/controller/jobmonitor"
	jm "github.com/goharbor/harbor/src/pkg/jobmonitor"
	"github.com/goharbor/harbor/src/server/v2.0/models"
	"github.com/goharbor/harbor/src/server/v2.0/restapi/operations/jobservice"
)

type jobServiceAPI struct {
	BaseAPI
	jobCtr jobmonitor.MonitorController
}

func newJobServiceAPI() *jobServiceAPI {
	return &jobServiceAPI{jobCtr: jobmonitor.Ctl}
}

func (j *jobServiceAPI) GetWorkerPools(ctx context.Context, _ jobservice.GetWorkerPoolsParams) middleware.Responder {
	if err := j.RequireSystemAccess(ctx, rbac.ActionList, rbac.ResourceJobServiceMonitor); err != nil {
		return j.SendError(ctx, err)
	}
	workPools, err := j.jobCtr.ListPools(ctx)
	if err != nil {
		return j.SendError(ctx, err)
	}
	return jobservice.NewGetWorkerPoolsOK().WithPayload(toWorkerPoolResponse(workPools))
}

func (j *jobServiceAPI) GetWorkers(ctx context.Context, params jobservice.GetWorkersParams) middleware.Responder {
	if err := j.RequireSystemAccess(ctx, rbac.ActionList, rbac.ResourceJobServiceMonitor); err != nil {
		return j.SendError(ctx, err)
	}
	workers, err := j.jobCtr.ListWorkers(ctx, params.PoolID)
	if err != nil {
		return j.SendError(ctx, err)
	}
	return jobservice.NewGetWorkersOK().WithPayload(toWorkerResponse(workers))
}

func (j *jobServiceAPI) StopRunningJob(ctx context.Context, params jobservice.StopRunningJobParams) middleware.Responder {
	if err := j.RequireSystemAccess(ctx, rbac.ActionStop, rbac.ResourceJobServiceMonitor); err != nil {
		return j.SendError(ctx, err)
	}
	err := j.jobCtr.StopRunningJobs(ctx, params.JobID)
	if err != nil {
		return j.SendError(ctx, err)
	}
	return jobservice.NewStopRunningJobOK()
}

func toWorkerResponse(wks []*jm.Worker) []*models.Worker {
	workers := make([]*models.Worker, 0)
	for _, w := range wks {
		if len(w.JobID) == 0 {
			workers = append(workers, &models.Worker{
				ID:     w.ID,
				PoolID: w.PoolID,
			})
		} else {
			var startAtTime, checkInAtTime *strfmt.DateTime
			if w.StartedAt != 0 {
				t := covertTime(w.StartedAt)
				startAtTime = &t
			}
			if w.CheckInAt != 0 {
				t := covertTime(w.CheckInAt)
				checkInAtTime = &t
			}
			workers = append(workers, &models.Worker{
				ID:        w.ID,
				JobName:   w.JobName,
				JobID:     w.JobID,
				PoolID:    w.PoolID,
				StartAt:   startAtTime,
				CheckinAt: checkInAtTime,
			})
		}
	}
	return workers
}

func toWorkerPoolResponse(wps []*jm.WorkerPool) []*models.WorkerPool {
	pools := make([]*models.WorkerPool, 0)
	for _, wp := range wps {
		p := &models.WorkerPool{
			Pid:          int64(wp.PID),
			HeartbeatAt:  covertTime(wp.HeartbeatAt),
			Concurrency:  int64(wp.Concurrency),
			WorkerPoolID: wp.ID,
			StartAt:      covertTime(wp.StartAt),
		}
		pools = append(pools, p)
	}
	return pools
}

func covertTime(t int64) strfmt.DateTime {
	if t == 0 {
		return strfmt.NewDateTime()
	}
	uxt := time.Unix(int64(t), 0)
	return strfmt.DateTime(uxt)
}

func toScheduleResponse(schs []*scheduler.Schedule) []*models.ScheduleTask {
	result := make([]*models.ScheduleTask, 0)
	for _, s := range schs {
		result = append(result, &models.ScheduleTask{
			ID:         s.ID,
			VendorType: s.VendorType,
			VendorID:   s.VendorID,
			Cron:       s.CRON,
			UpdateTime: strfmt.DateTime(s.UpdateTime),
		})
	}
	return result
}

func (j *jobServiceAPI) ListJobQueues(ctx context.Context, _ jobservice.ListJobQueuesParams) middleware.Responder {
	if err := j.RequireSystemAccess(ctx, rbac.ActionList, rbac.ResourceJobServiceMonitor); err != nil {
		return j.SendError(ctx, err)
	}
	queues, err := j.jobCtr.ListQueues(ctx)
	if err != nil {
		return j.SendError(ctx, err)
	}
	return jobservice.NewListJobQueuesOK().WithPayload(toQueueResponse(queues))
}

func toQueueResponse(queues []*jm.Queue) []*models.JobQueue {
	result := make([]*models.JobQueue, 0)
	for _, q := range queues {
		result = append(result, &models.JobQueue{
			JobType: q.JobType,
			Count:   q.Count,
			Latency: q.Latency,
			Paused:  q.Paused,
		})
	}
	return result
}

func (j *jobServiceAPI) ActionPendingJobs(ctx context.Context, params jobservice.ActionPendingJobsParams) middleware.Responder {
	if err := j.RequireSystemAccess(ctx, rbac.ActionStop, rbac.ResourceJobServiceMonitor); err != nil {
		return j.SendError(ctx, err)
	}
	jobType := strings.ToUpper(params.JobType)
	action := strings.ToLower(params.ActionRequest.Action)
	if !strings.EqualFold(action, "stop") && !strings.EqualFold(action, "resume") && !strings.EqualFold(action, "pause") {
		return j.SendError(ctx, errors.BadRequestError(fmt.Errorf("the action is not supported")))
	}
	if strings.EqualFold(action, "stop") {
		err := j.jobCtr.StopPendingJobs(ctx, jobType)
		if err != nil {
			return j.SendError(ctx, err)
		}
	}
	if strings.EqualFold(action, "pause") {
		err := j.jobCtr.PauseJobQueues(ctx, jobType)
		if err != nil {
			return j.SendError(ctx, err)
		}
	}
	if strings.EqualFold(action, "resume") {
		err := j.jobCtr.ResumeJobQueues(ctx, jobType)
		if err != nil {
			return j.SendError(ctx, err)
		}
	}
	return jobservice.NewActionPendingJobsOK()
}

func (j *jobServiceAPI) ActionGetJobLog(ctx context.Context, params jobservice.ActionGetJobLogParams) middleware.Responder {
	if err := j.RequireSystemAccess(ctx, rbac.ActionList, rbac.ResourceJobServiceMonitor); err != nil {
		return j.SendError(ctx, err)
	}
	log, err := j.jobCtr.GetJobLog(ctx, params.JobID)
	if err != nil {
		return j.SendError(ctx, err)
	}
	return jobservice.NewActionGetJobLogOK().WithContentType("text/plain").WithPayload(string(log))
}
