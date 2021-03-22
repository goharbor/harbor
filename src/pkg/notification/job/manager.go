package job

import (
	"context"
	"github.com/goharbor/harbor/src/lib/q"
	"github.com/goharbor/harbor/src/pkg/notification/job/dao"
	"github.com/goharbor/harbor/src/pkg/notification/job/model"
)

var (
	// Mgr is a global variable for the default notification job
	Mgr = NewManager()
)

// Manager manages notification jobs recorded in database
type Manager interface {
	// Create create a notification job
	Create(ctx context.Context, job *model.Job) (int64, error)

	// List list notification jobs
	List(ctx context.Context, query *q.Query) ([]*model.Job, error)

	// Update update notification job
	Update(ctx context.Context, job *model.Job, props ...string) error

	// ListJobsGroupByEventType lists last triggered jobs group by event type
	ListJobsGroupByEventType(ctx context.Context, policyID int64) ([]*model.Job, error)

	// Count ...
	Count(ctx context.Context, query *q.Query) (total int64, err error)
}

var _ Manager = &manager{}

type manager struct {
	dao dao.DAO
}

// NewManager ...
func NewManager() Manager {
	return &manager{
		dao: dao.New(),
	}
}

// Create ...
func (d *manager) Create(ctx context.Context, job *model.Job) (int64, error) {
	return d.dao.Create(ctx, job)
}

// Count ...
func (d *manager) Count(ctx context.Context, query *q.Query) (int64, error) {
	return d.dao.Count(ctx, query)
}

// List ...
func (d *manager) List(ctx context.Context, query *q.Query) ([]*model.Job, error) {
	return d.dao.List(ctx, query)
}

// Update ...
func (d *manager) Update(ctx context.Context, job *model.Job, props ...string) error {
	return d.dao.Update(ctx, job, props...)
}

// ListJobsGroupByEventType lists last triggered jobs group by event type
func (d *manager) ListJobsGroupByEventType(ctx context.Context, policyID int64) ([]*model.Job, error) {
	return d.dao.GetLastTriggerJobsGroupByEventType(ctx, policyID)
}
