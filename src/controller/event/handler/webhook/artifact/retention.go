package artifact

import (
	"errors"
	"fmt"
	"strings"

	"github.com/goharbor/harbor/src/core/api"

	"github.com/goharbor/harbor/src/controller/event"
	"github.com/goharbor/harbor/src/controller/event/handler/util"
	"github.com/goharbor/harbor/src/core/config"
	"github.com/goharbor/harbor/src/lib/log"
	"github.com/goharbor/harbor/src/pkg/notification"
	"github.com/goharbor/harbor/src/pkg/notifier/model"
)

// RetentionHandler preprocess tag retention event data
type RetentionHandler struct {
}

// Handle ...
func (r *RetentionHandler) Handle(value interface{}) error {
	if !config.NotificationEnable() {
		log.Debug("notification feature is not enabled")
		return nil
	}

	trEvent, ok := value.(*event.RetentionEvent)
	if !ok {
		return errors.New("invalid tag retention event type")
	}
	if trEvent == nil {
		return fmt.Errorf("nil tag retention event")
	}

	payload, project, err := constructRetentionPayload(trEvent)
	if err != nil {
		return err
	}

	policies, err := notification.PolicyMgr.GetRelatedPolices(project, trEvent.EventType)
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

func constructRetentionPayload(event *event.RetentionEvent) (*model.Payload, int64, error) {
	task, err := api.RetentionController.GetRetentionExecTask(event.TaskID)
	if err != nil {
		log.Errorf("failed to get retention task %d: error: %v", event.TaskID, err)
		return nil, 0, err
	}
	if task == nil {
		return nil, 0, fmt.Errorf("task %d not found with retention event", event.TaskID)
	}

	execution, err := api.RetentionController.GetRetentionExec(task.ExecutionID)
	if err != nil {
		log.Errorf("failed to get retention execution %d: error: %v", task.ExecutionID, err)
		return nil, 0, err
	}
	if execution == nil {
		return nil, 0, fmt.Errorf("execution %d not found with retention event", task.ExecutionID)
	}

	md, err := api.RetentionController.GetRetention(execution.PolicyID)
	if err != nil {
		log.Errorf("failed to get tag retention policy %d: error: %v", execution.PolicyID, err)
		return nil, 0, err
	}
	if md == nil {
		return nil, 0, fmt.Errorf("policy %d not found with tag retention event", execution.PolicyID)
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
			Retention: &model.Retention{
				Total:             task.Total,
				Retained:          task.Retained,
				HarborHostname:    hostname,
				ProjectName:       event.Deleted[0].Target.Namespace,
				RetentionPolicyID: execution.PolicyID,
				Status:            event.Status,
				RetentionRules:    []*model.RetentionRule{},
			},
		},
	}

	for _, v := range event.Deleted {
		target := v.Target
		succeedArtifact := &model.ArtifactInfo{
			Type:       target.Kind,
			Status:     task.Status,
			NameAndTag: target.Repository + ":" + target.Tags[0],
		}
		payload.EventData.Retention.SuccessfulArtifact = []*model.ArtifactInfo{succeedArtifact}
	}

	for _, v := range md.Rules {
		retentionRule := &model.RetentionRule{
			Template:       v.Template,
			Parameters:     v.Parameters,
			TagSelectors:   v.TagSelectors,
			ScopeSelectors: v.ScopeSelectors,
		}
		payload.EventData.Retention.RetentionRules = append(payload.EventData.Retention.RetentionRules, retentionRule)
	}

	return payload, event.Deleted[0].Target.NamespaceID, nil
}
