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

package dao

import (
	"context"

	"github.com/goharbor/harbor/src/lib/errors"
	"github.com/goharbor/harbor/src/lib/orm"
	"github.com/goharbor/harbor/src/lib/q"
	"github.com/goharbor/harbor/src/pkg/queuestatus/model"
)

// ErrQueueTypeDup ...
var ErrQueueTypeDup = errors.ConflictError(nil).WithMessage("duplicated queue type")

func init() {
	orm.RegisterModel(
		new(model.JobQueueStatus),
	)
}

// DAO the dao for queue status
type DAO interface {
	// Query query queue status
	Query(ctx context.Context, query *q.Query) ([]*model.JobQueueStatus, error)
	// GetByJobType get queue status by JobType
	GetByJobType(ctx context.Context, jobType string) (*model.JobQueueStatus, error)
	// UpdateStatus update queue status
	UpdateStatus(ctx context.Context, jobType string, paused bool) error
	// InsertOrUpdate create a queue status or update it if it already exists
	InsertOrUpdate(ctx context.Context, queue *model.JobQueueStatus) (int64, error)
}

type dao struct {
}

// New create queue status DAO
func New() DAO {
	return &dao{}
}

func (d *dao) Query(ctx context.Context, query *q.Query) ([]*model.JobQueueStatus, error) {
	query = q.MustClone(query)
	qs, err := orm.QuerySetter(ctx, &model.JobQueueStatus{}, query)
	if err != nil {
		return nil, err
	}
	var queueStatusList []*model.JobQueueStatus
	if _, err := qs.All(&queueStatusList); err != nil {
		return nil, err
	}
	return queueStatusList, nil
}

func (d *dao) GetByJobType(ctx context.Context, jobType string) (*model.JobQueueStatus, error) {
	queueList, err := d.Query(ctx, q.New(q.KeyWords{"JobType": jobType}))
	if err != nil {
		return nil, err
	}
	if len(queueList) > 0 {
		return queueList[0], nil
	}
	return nil, nil
}

func (d *dao) UpdateStatus(ctx context.Context, jobType string, paused bool) error {
	_, err := d.InsertOrUpdate(ctx, &model.JobQueueStatus{JobType: jobType, Paused: paused})
	return err
}

func (d *dao) InsertOrUpdate(ctx context.Context, queue *model.JobQueueStatus) (int64, error) {
	o, err := orm.FromContext(ctx)
	if err != nil {
		return 0, err
	}
	return o.InsertOrUpdate(queue, "job_type")
}
