package gc

import (
	"context"
	"encoding/json"
	"github.com/goharbor/harbor/src/jobservice/job"
	"github.com/goharbor/harbor/src/lib/errors"
	"github.com/goharbor/harbor/src/lib/log"
	"github.com/goharbor/harbor/src/lib/q"
	"github.com/goharbor/harbor/src/pkg/scheduler"
	"github.com/goharbor/harbor/src/pkg/task"
)

var (
	// GCCtl is a global garbage collection controller instance
	GCCtl = NewController()
)

const (
	// SchedulerCallback ...
	SchedulerCallback = "GARBAGE_COLLECTION"
	// gcVendorType ...
	gcVendorType = "GARBAGE_COLLECTION"
)

// Controller manages the tags
type Controller interface {
	// Start start a manual gc job
	Start(ctx context.Context, parameters map[string]interface{}) error
	// Stop stop a gc job
	Stop(ctx context.Context, taskID int64) error
	// GetLog get the gc log by id
	GetLog(ctx context.Context, id int64) ([]byte, error)
	// History list all of gc executions
	History(ctx context.Context, query *q.Query) ([]*History, error)
	// Count count the gc executions
	Count(ctx context.Context, query *q.Query) (int64, error)
	// GetSchedule get the current gc schedule
	GetSchedule(ctx context.Context) (*scheduler.Schedule, error)
	// CreateSchedule create the gc schedule with cron string
	CreateSchedule(ctx context.Context, cron string, parameters map[string]interface{}) (int64, error)
	// DeleteSchedule remove the gc schedule
	DeleteSchedule(ctx context.Context) error
}

// NewController creates an instance of the default repository controller
func NewController() Controller {
	return &controller{
		taskMgr:      task.NewManager(),
		exeMgr:       task.NewExecutionManager(),
		schedulerMgr: scheduler.New(),
	}
}

type controller struct {
	taskMgr      task.Manager
	exeMgr       task.ExecutionManager
	schedulerMgr scheduler.Scheduler
}

// Start starts the manual GC
func (c *controller) Start(ctx context.Context, parameters map[string]interface{}) error {
	execID, err := c.exeMgr.Create(ctx, gcVendorType, -1, task.ExecutionTriggerManual, parameters)
	if err != nil {
		return err
	}
	taskID, err := c.taskMgr.Create(ctx, execID, &task.Job{
		Name: job.ImageGC,
		Metadata: &job.Metadata{
			JobKind: job.KindGeneric,
		},
		Parameters: parameters,
	})
	if err != nil {
		return err
	}
	defer func() {
		if err == nil {
			return
		}
		if err := c.taskMgr.Stop(ctx, taskID); err != nil {
			log.Errorf("failed to stop the task %d: %v", taskID, err)
		}
	}()
	return nil
}

// Stop ...
func (c *controller) Stop(ctx context.Context, taskID int64) error {
	return c.taskMgr.Stop(ctx, taskID)
}

// GetLog ...
func (c *controller) GetLog(ctx context.Context, executionID int64) ([]byte, error) {
	tasks, err := c.taskMgr.List(ctx, q.New(q.KeyWords{"ExecutionID": executionID}))
	if err != nil {
		return nil, err
	}
	if len(tasks) == 0 {
		return nil, errors.New(nil).WithCode(errors.NotFoundCode).WithMessage("no gc task is found")
	}
	return c.taskMgr.GetLog(ctx, tasks[0].ID)
}

// Count ...
func (c *controller) Count(ctx context.Context, query *q.Query) (int64, error) {
	query.Keywords["VendorType"] = gcVendorType
	return c.exeMgr.Count(ctx, query)
}

// History ...
func (c *controller) History(ctx context.Context, query *q.Query) ([]*History, error) {
	var hs []*History

	query.Keywords["VendorType"] = gcVendorType
	exes, err := c.exeMgr.List(ctx, query)
	if err != nil {
		return nil, err
	}
	for _, exe := range exes {
		tasks, err := c.taskMgr.List(ctx, q.New(q.KeyWords{"ExecutionID": exe.ID}))
		if err != nil {
			return nil, err
		}
		if len(tasks) == 0 {
			continue
		}

		extraAttrsString, err := json.Marshal(exe.ExtraAttrs)
		if err != nil {
			return nil, err
		}
		hs = append(hs, &History{
			ID:         exe.ID,
			Name:       gcVendorType,
			Kind:       exe.Trigger,
			Parameters: string(extraAttrsString),
			Deleted:    false,
			Schedule: Schedule{Schedule: &ScheduleParam{
				Type: exe.Trigger,
			}},
			Status:       tasks[0].Status,
			CreationTime: tasks[0].CreationTime,
			UpdateTime:   tasks[0].UpdateTime,
		})
	}
	return hs, nil
}

// GetSchedule ...
func (c *controller) GetSchedule(ctx context.Context) (*scheduler.Schedule, error) {
	sch, err := c.schedulerMgr.ListSchedules(ctx, q.New(q.KeyWords{"VendorType": gcVendorType}))
	if err != nil {
		return nil, err
	}
	if len(sch) == 0 {
		return nil, errors.New(nil).WithCode(errors.NotFoundCode).WithMessage("no gc schedule is found")
	}
	if sch[0] == nil {
		return nil, errors.New(nil).WithCode(errors.NotFoundCode).WithMessage("no gc schedule is found")
	}
	return sch[0], nil
}

// CreateSchedule ...
func (c *controller) CreateSchedule(ctx context.Context, cron string, parameters map[string]interface{}) (int64, error) {
	return c.schedulerMgr.Schedule(ctx, gcVendorType, -1, cron, SchedulerCallback, parameters)
}

// DeleteSchedule ...
func (c *controller) DeleteSchedule(ctx context.Context) error {
	return c.schedulerMgr.UnScheduleByVendor(ctx, gcVendorType, -1)
}
