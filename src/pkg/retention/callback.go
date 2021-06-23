package retention

import (
	"context"
	"encoding/json"
	"github.com/goharbor/harbor/src/controller/event/metadata"
	"github.com/goharbor/harbor/src/jobservice/job"
	"github.com/goharbor/harbor/src/lib/log"
	"github.com/goharbor/harbor/src/lib/selector"
	"github.com/goharbor/harbor/src/pkg/notifier/event"
	"github.com/goharbor/harbor/src/pkg/task"
)

func init() {

	if err := task.RegisterCheckInProcessor(job.Retention, retentionTaskCheckInProcessor); err != nil {
		log.Fatalf("failed to register the checkin processor for the retention job, error %v", err)
	}

}

func retentionTaskCheckInProcessor(ctx context.Context, t *task.Task, sc *job.StatusChange) (err error) {
	taskID := t.ID
	status := t.Status
	log.Debugf("received retention task status update event: task-%d, status-%s", taskID, status)
	// handle checkin
	if sc.CheckIn != "" {
		var retainObj struct {
			Total    int                `json:"total"`
			Retained int                `json:"retained"`
			Deleted  []*selector.Result `json:"deleted"`
		}
		if err := json.Unmarshal([]byte(sc.CheckIn), &retainObj); err != nil {
			log.Errorf("failed to resolve checkin of retention task %d: %v", taskID, err)

			return err
		}

		t, err := task.Mgr.Get(ctx, taskID)
		if err != nil {
			return err
		}

		t.ExtraAttrs["total"] = retainObj.Total
		t.ExtraAttrs["retained"] = retainObj.Retained

		err = task.Mgr.UpdateExtraAttrs(ctx, taskID, t.ExtraAttrs)
		if err != nil {
			log.G(ctx).WithField("error", err).Errorf("failed to update of retention task %d", taskID)
			return err
		}

		e := &event.Event{}
		metaData := &metadata.RetentionMetaData{
			Total:    retainObj.Total,
			Retained: retainObj.Retained,
			Deleted:  retainObj.Deleted,
			Status:   "SUCCESS",
			TaskID:   taskID,
		}

		if err := e.Build(metaData); err == nil {
			if err := e.Publish(); err != nil {
				log.G(ctx).WithField("error", err).Errorf("tag retention job hook handler: event publish")
			}
		} else {
			log.G(ctx).WithField("error", err).Errorf("tag retention job hook handler: event publish")
		}
	}
	return nil
}
