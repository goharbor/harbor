package notification

import (
	"errors"
	"fmt"
	"strings"

	"github.com/goharbor/harbor/src/replication/dao/models"

	commonModels "github.com/goharbor/harbor/src/common/models"
	"github.com/goharbor/harbor/src/common/utils/log"
	"github.com/goharbor/harbor/src/core/config"
	"github.com/goharbor/harbor/src/pkg/notification"
	"github.com/goharbor/harbor/src/pkg/notifier/model"
	notifyModel "github.com/goharbor/harbor/src/pkg/notifier/model"
	"github.com/goharbor/harbor/src/replication"
)

// ReplicationPreprocessHandler preprocess replication event data
type ReplicationPreprocessHandler struct {
}

// Handle ...
func (r *ReplicationPreprocessHandler) Handle(value interface{}) error {
	if !config.NotificationEnable() {
		log.Debug("notification feature is not enabled")
		return nil
	}

	rpEvent, ok := value.(*notifyModel.ReplicationEvent)
	if !ok {
		return errors.New("invalid replication event type")
	}
	if rpEvent == nil {
		return fmt.Errorf("nil replication event")
	}

	task, err := replication.OperationCtl.GetTask(rpEvent.ReplicationTaskID)
	if err != nil {
		log.Errorf("failed to get replication task %d: error: %v", rpEvent.ReplicationTaskID, err)
		return err
	}
	if task == nil {
		return fmt.Errorf("task %d not found with replication event", rpEvent.ReplicationTaskID)
	}

	execution, err := replication.OperationCtl.GetExecution(task.ExecutionID)
	if err != nil {
		log.Errorf("failed to get replication execution %d: error: %v", task.ExecutionID, err)
		return err
	}
	if execution == nil {
		return fmt.Errorf("execution %d not found with replication event", task.ExecutionID)
	}

	rpPolicy, err := replication.PolicyCtl.Get(execution.PolicyID)
	if err != nil {
		log.Errorf("failed to get replication policy %d: error: %v", execution.PolicyID, err)
		return err
	}
	if rpPolicy == nil {
		return fmt.Errorf("policy %d not found with replication event", execution.PolicyID)
	}

	payload, project, err := constructReplicationPayload(rpEvent)
	if err != nil {
		return err
	}

	policies, err := notification.PolicyMgr.GetRelatedPolices(project.ProjectID, rpEvent.EventType)
	if err != nil {
		log.Errorf("failed to find policy for %s event: %v", rpEvent.EventType, err)
		return err
	}
	if len(policies) == 0 {
		log.Debugf("cannot find policy for %s event: %v", rpEvent.EventType, rpEvent)
		return nil
	}
	err = sendHookWithPolicies(policies, payload, rpEvent.EventType)
	if err != nil {
		return err
	}
	return nil
}

// IsStateful ...
func (r *ReplicationPreprocessHandler) IsStateful() bool {
	return false
}

func constructReplicationPayload(event *notifyModel.ReplicationEvent) (*model.Payload, *commonModels.Project, error) {
	task, err := replication.OperationCtl.GetTask(event.ReplicationTaskID)
	if err != nil {
		log.Errorf("failed to get replication task %d: error: %v", event.ReplicationTaskID, err)
		return nil, nil, err
	}
	if task == nil {
		return nil, nil, fmt.Errorf("task %d not found with replication event", event.ReplicationTaskID)
	}

	execution, err := replication.OperationCtl.GetExecution(task.ExecutionID)
	if err != nil {
		log.Errorf("failed to get replication execution %d: error: %v", task.ExecutionID, err)
		return nil, nil, err
	}
	if execution == nil {
		return nil, nil, fmt.Errorf("execution %d not found with replication event", task.ExecutionID)
	}

	rpPolicy, err := replication.PolicyCtl.Get(execution.PolicyID)
	if err != nil {
		log.Errorf("failed to get replication policy %d: error: %v", execution.PolicyID, err)
		return nil, nil, err
	}
	if rpPolicy == nil {
		return nil, nil, fmt.Errorf("policy %d not found with replication event", execution.PolicyID)
	}

	registry, err := replication.RegistryMgr.Get(rpPolicy.DestRegistry.ID)
	if err != nil {
		log.Errorf("failed to get replication registry %d: error: %v", rpPolicy.DestRegistry.ID, err)
		return nil, nil, err
	}
	if registry == nil {
		return nil, nil, fmt.Errorf("registry %d not found with replication event", rpPolicy.DestRegistry.ID)
	}

	srcNamespace, srcNameTag := getMetadataFromResource(task.SrcResource)
	destNamespace, _ := getMetadataFromResource(task.DstResource)

	prjName := destNamespace
	// push based replication policy get project from src
	if rpPolicy.DestRegistry.ID > 0 {
		prjName = srcNamespace
	}

	ext, err := config.ExtURL()
	if err != nil {
		log.Errorf("Error while reading external endpoint: %v", err)
	}
	hostname := strings.Split(ext, ":")[0]

	payload := &notifyModel.Payload{
		Type:     event.EventType,
		OccurAt:  event.OccurAt.Unix(),
		Operator: string(execution.Trigger),
		EventData: &model.EventData{
			Replication: &model.Replication{
				HarborHostname:     hostname,
				JobStatus:          event.Status,
				Description:        rpPolicy.Description,
				ArtifactType:       task.ResourceType,
				AuthenticationType: string(registry.Credential.Type),
				OverrideMode:       rpPolicy.Override,
				TriggerType:        string(execution.Trigger),
				ExecutionTimestamp: execution.StartTime.Unix(),
				SrcRegistryType:    string(rpPolicy.SrcRegistry.Type),
				SrcRegistryName:    rpPolicy.SrcRegistry.Name,
				SrcEndpoint:        rpPolicy.SrcRegistry.URL,
				SrcProvider:        "",
				SrcNamespace:       srcNamespace,
				SrcProjectName:     srcNamespace,
				DestRegistryType:   string(rpPolicy.DestRegistry.Type),
				DestRegistryName:   rpPolicy.DestRegistry.Name,
				DestEndpoint:       rpPolicy.DestRegistry.URL,
				DestProvider:       "",
				DestNamespace:      rpPolicy.DestNamespace,
				DestProjectName:    rpPolicy.DestNamespace,
			},
		},
	}

	if task.Status == models.TaskStatusSucceed {
		succeedArtifact := &model.SuccessfulArtifact{
			Type:    task.ResourceType,
			Status:  task.Status,
			NameTag: srcNameTag,
		}
		payload.EventData.Replication.SuccessfulArtifact = append(payload.EventData.Replication.SuccessfulArtifact, succeedArtifact)
	} else {
		failedArtifact := &model.FailedArtifact{
			Type:    task.ResourceType,
			Status:  task.Status,
			NameTag: srcNameTag,
		}
		payload.EventData.Replication.FailedArtifact = append(payload.EventData.Replication.FailedArtifact, failedArtifact)
	}

	project, err := config.GlobalProjectMgr.Get(prjName)
	if err != nil {
		log.Errorf("failed to get project %s, error: %v", prjName, err)
		return nil, nil, err
	}
	if project == nil {
		return nil, nil, fmt.Errorf("project %s not found of replication event", prjName)
	}

	return payload, project, nil
}

func getMetadataFromResource(resource string) (namespace, nameAndTag string) {
	meta := strings.Split(resource, "/")
	return meta[0], meta[1]
}
