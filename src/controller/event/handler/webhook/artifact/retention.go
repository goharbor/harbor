package artifact

import (
	"context"
	"fmt"
	"strings"

	"github.com/goharbor/harbor/src/controller/retention"
	"github.com/goharbor/harbor/src/lib/config"
	"github.com/goharbor/harbor/src/lib/orm"

	"github.com/goharbor/harbor/src/controller/event"
	"github.com/goharbor/harbor/src/controller/event/handler/util"
	evtModel "github.com/goharbor/harbor/src/controller/event/model"
	"github.com/goharbor/harbor/src/lib/errors"
	"github.com/goharbor/harbor/src/lib/log"
	"github.com/goharbor/harbor/src/pkg/notification"
	"github.com/goharbor/harbor/src/pkg/notifier/model"
)

// RetentionHandler preprocess tag retention event data
type RetentionHandler struct {
}

// Name ...
func (r *RetentionHandler) Name() string {
	return "RetentionWebhook"
}

// Handle ...
func (r *RetentionHandler) Handle(ctx context.Context, value interface{}) error {
	if !config.NotificationEnable(ctx) {
		log.Debug("notification feature is not enabled")
		return nil
	}

	trEvent, ok := value.(*event.RetentionEvent)
	if !ok {
		return errors.New("invalid tag retention event type")
	}
	if trEvent == nil {
		return errors.New("nil tag retention event")
	}
	if len(trEvent.Deleted) == 0 {
		log.Debugf("empty delete info of retention event")
		return nil
	}

	payload, dryRun, project, err := r.constructRetentionPayload(trEvent)
	if err != nil {
		return err
	}
	// if dry run, do not trigger webhook
	if dryRun {
		log.Debugf("retention task %v is dry run", trEvent.TaskID)
		return nil
	}

	policies, err := notification.PolicyMgr.GetRelatedPolices(ctx, project, trEvent.EventType)
	if err != nil {
		log.Errorf("failed to find policy for %s event: %v", trEvent.EventType, err)
		return err
	}
	if len(policies) == 0 {
		log.Debugf("cannot find policy for %s event: %v", trEvent.EventType, trEvent)
		return nil
	}
	err = util.SendHookWithPolicies(policies, payload, trEvent.EventType)
	if err != nil {
		return err
	}
	return nil
}

// IsStateful ...
func (r *RetentionHandler) IsStateful() bool {
	return false
}

func (r *RetentionHandler) constructRetentionPayload(event *event.RetentionEvent) (*model.Payload, bool, int64, error) {
	ctx := orm.Context()
	task, err := retention.Ctl.GetRetentionExecTask(ctx, event.TaskID)
	if err != nil {
		log.Errorf("failed to get retention task %d: error: %v", event.TaskID, err)
		return nil, false, 0, err
	}
	if task == nil {
		return nil, false, 0, fmt.Errorf("task %d not found with retention event", event.TaskID)
	}

	execution, err := retention.Ctl.GetRetentionExec(ctx, task.ExecutionID)
	if err != nil {
		log.Errorf("failed to get retention execution %d: error: %v", task.ExecutionID, err)
		return nil, false, 0, err
	}
	if execution == nil {
		return nil, false, 0, fmt.Errorf("execution %d not found with retention event", task.ExecutionID)
	}

	if execution.DryRun {
		return nil, true, 0, nil
	}

	md, err := retention.Ctl.GetRetention(ctx, execution.PolicyID)
	if err != nil {
		log.Errorf("failed to get tag retention policy %d: error: %v", execution.PolicyID, err)
		return nil, false, 0, err
	}
	if md == nil {
		return nil, false, 0, fmt.Errorf("policy %d not found with tag retention event", execution.PolicyID)
	}

	extURL, err := config.ExtURL()
	if err != nil {
		log.Errorf("Error while reading external endpoint URL: %v", err)
	}
	hostname := strings.Split(extURL, ":")[0]

	payload := &model.Payload{
		Type:     event.EventType,
		OccurAt:  event.OccurAt.Unix(),
		Operator: execution.Trigger,
		EventData: &model.EventData{
			Retention: &evtModel.Retention{
				Total:             task.Total,
				Retained:          task.Retained,
				HarborHostname:    hostname,
				ProjectName:       event.Deleted[0].Target.Namespace,
				RetentionPolicyID: execution.PolicyID,
				Status:            event.Status,
				RetentionRules:    []*evtModel.RetentionRule{},
			},
		},
	}

	for _, v := range event.Deleted {
		target := v.Target
		deletedArtifact := &evtModel.ArtifactInfo{
			Type:   target.Kind,
			Status: event.Status,
		}
		if len(target.Tags) != 0 {
			deletedArtifact.NameAndTag = target.Repository + ":" + target.Tags[0]
		}
		payload.EventData.Retention.DeletedArtifact = []*evtModel.ArtifactInfo{deletedArtifact}
	}

	for _, v := range md.Rules {
		retentionRule := &evtModel.RetentionRule{
			Template:       v.Template,
			Parameters:     v.Parameters,
			TagSelectors:   v.TagSelectors,
			ScopeSelectors: v.ScopeSelectors,
		}
		payload.EventData.Retention.RetentionRules = append(payload.EventData.Retention.RetentionRules, retentionRule)
	}

	return payload, false, event.Deleted[0].Target.NamespaceID, nil
}
