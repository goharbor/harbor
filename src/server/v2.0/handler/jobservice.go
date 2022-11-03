//  Copyright Project Harbor Authors
//
//  Licensed under the Apache License, Version 2.0 (the "License");
//  you may not use this file except in compliance with the License.
//  You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
//  Unless required by applicable law or agreed to in writing, software
//  distributed under the License is distributed on an "AS IS" BASIS,
//  WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
//  See the License for the specific language governing permissions and
//  limitations under the License.

package handler

import (
	"context"
	"time"

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

func (j *jobServiceAPI) GetWorkerPools(ctx context.Context, params jobservice.GetWorkerPoolsParams) middleware.Responder {
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
	err := j.jobCtr.StopRunningJob(ctx, params.JobID)
	if err != nil {
		return j.SendError(ctx, err)
	}
	return jobservice.NewStopRunningJobOK()
}

func toWorkerResponse(wks []*jm.Worker) []*models.Worker {
	workers := make([]*models.Worker, 0)
	for _, w := range wks {
		p := &models.Worker{
			ID:        w.ID,
			JobName:   w.JobName,
			JobID:     w.JobID,
			PoolID:    w.PoolID,
			Args:      w.Args,
			StartAt:   covertTime(w.StartedAt),
			CheckinAt: covertTime(w.CheckInAt),
		}
		workers = append(workers, p)
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
	uxt := time.Unix(int64(t), 0)
	return strfmt.DateTime(uxt)
}
