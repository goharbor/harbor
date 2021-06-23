package artifact

import (
	"context"
	"errors"
	"fmt"
	proModels "github.com/goharbor/harbor/src/pkg/project/models"
	"strings"

	"github.com/goharbor/harbor/src/lib/config"

	"github.com/goharbor/harbor/src/controller/event"
	"github.com/goharbor/harbor/src/controller/event/handler/util"
	ctlModel "github.com/goharbor/harbor/src/controller/event/model"
	"github.com/goharbor/harbor/src/controller/project"
	"github.com/goharbor/harbor/src/controller/replication"
	"github.com/goharbor/harbor/src/jobservice/job"
	"github.com/goharbor/harbor/src/lib/log"
	"github.com/goharbor/harbor/src/lib/orm"
	"github.com/goharbor/harbor/src/pkg/notification"
	"github.com/goharbor/harbor/src/pkg/notifier/model"
	"github.com/goharbor/harbor/src/pkg/reg"
	rpModel "github.com/goharbor/harbor/src/pkg/reg/model"
)

// ReplicationHandler preprocess replication event data
type ReplicationHandler struct {
}

// Name ...
func (r *ReplicationHandler) Name() string {
	return "ReplicationWebhook"
}

// Handle ...
func (r *ReplicationHandler) Handle(ctx context.Context, value interface{}) error {
	if !config.NotificationEnable(ctx) {
		log.Debug("notification feature is not enabled")
		return nil
	}

	rpEvent, ok := value.(*event.ReplicationEvent)
	if !ok {
		return errors.New("invalid replication event type")
	}
	if rpEvent == nil {
		return fmt.Errorf("nil replication event")
	}

	payload, project, err := constructReplicationPayload(rpEvent)
	if err != nil {
		return err
	}

	policies, err := notification.PolicyMgr.GetRelatedPolices(orm.Context(), project.ProjectID, rpEvent.EventType)
	if err != nil {
		log.Errorf("failed to find policy for %s event: %v", rpEvent.EventType, err)
		return err
	}
	if len(policies) == 0 {
		log.Debugf("cannot find policy for %s event: %v", rpEvent.EventType, rpEvent)
		return nil
	}
	err = util.SendHookWithPolicies(policies, payload, rpEvent.EventType)
	if err != nil {
		return err
	}
	return nil
}

// IsStateful ...
func (r *ReplicationHandler) IsStateful() bool {
	return false
}

func constructReplicationPayload(event *event.ReplicationEvent) (*model.Payload, *proModels.Project, error) {
	ctx := orm.Context()
	task, err := replication.Ctl.GetTask(ctx, event.ReplicationTaskID)
	if err != nil {
		log.Errorf("failed to get replication task %d: error: %v", event.ReplicationTaskID, err)
		return nil, nil, err
	}

	execution, err := replication.Ctl.GetExecution(ctx, task.ExecutionID)
	if err != nil {
		log.Errorf("failed to get replication execution %d: error: %v", task.ExecutionID, err)
		return nil, nil, err
	}

	rpPolicy, err := replication.Ctl.GetPolicy(ctx, execution.PolicyID)
	if err != nil {
		log.Errorf("failed to get replication policy %d: error: %v", execution.PolicyID, err)
		return nil, nil, err
	}
	if rpPolicy == nil {
		return nil, nil, fmt.Errorf("policy %d not found with replication event", execution.PolicyID)
	}

	var remoteRegID int64
	if rpPolicy.SrcRegistry != nil && rpPolicy.SrcRegistry.ID > 0 {
		remoteRegID = rpPolicy.SrcRegistry.ID
	}

	if rpPolicy.DestRegistry != nil && rpPolicy.DestRegistry.ID > 0 {
		remoteRegID = rpPolicy.DestRegistry.ID
	}

	remoteRegistry, err := reg.Mgr.Get(ctx, remoteRegID)
	if err != nil {
		log.Errorf("failed to get replication remoteRegistry registry %d: error: %v", remoteRegID, err)
		return nil, nil, err
	}

	srcNamespace, srcNameAndTag := getMetadataFromResource(task.SourceResource)
	destNamespace, destNameAndTag := getMetadataFromResource(task.DestinationResource)

	extURL, err := config.ExtURL()
	if err != nil {
		log.Errorf("Error while reading external endpoint URL: %v", err)
	}
	hostname := strings.Split(extURL, ":")[0]

	remoteRes := &ctlModel.ReplicationResource{
		RegistryName: remoteRegistry.Name,
		RegistryType: string(remoteRegistry.Type),
		Endpoint:     remoteRegistry.URL,
		Namespace:    srcNamespace,
	}

	ext, err := config.ExtEndpoint()
	if err != nil {
		log.Errorf("Error while reading external endpoint: %v", err)
	}
	localRes := &ctlModel.ReplicationResource{
		RegistryType: string(rpModel.RegistryTypeHarbor),
		Endpoint:     ext,
		Namespace:    destNamespace,
	}

	payload := &model.Payload{
		Type:     event.EventType,
		OccurAt:  event.OccurAt.Unix(),
		Operator: string(execution.Trigger),
		EventData: &model.EventData{
			Replication: &ctlModel.Replication{
				HarborHostname:     hostname,
				JobStatus:          event.Status,
				Description:        rpPolicy.Description,
				PolicyCreator:      rpPolicy.Creator,
				ArtifactType:       task.ResourceType,
				AuthenticationType: string(remoteRegistry.Credential.Type),
				OverrideMode:       rpPolicy.Override,
				TriggerType:        string(execution.Trigger),
				ExecutionTimestamp: execution.StartTime.Unix(),
			},
		},
	}

	var prjName, nameAndTag string
	// remote(src) -> local harbor(dest)
	if rpPolicy.SrcRegistry != nil {
		payload.EventData.Replication.SrcResource = remoteRes
		payload.EventData.Replication.DestResource = localRes
		prjName = destNamespace
		nameAndTag = destNameAndTag
	}

	// local harbor(src) -> remote(dest)
	if rpPolicy.DestRegistry != nil {
		payload.EventData.Replication.DestResource = remoteRes
		payload.EventData.Replication.SrcResource = localRes
		prjName = srcNamespace
		nameAndTag = srcNameAndTag
	}

	if event.Status == string(job.SuccessStatus) {
		succeedArtifact := &ctlModel.ArtifactInfo{
			Type:       task.ResourceType,
			Status:     task.Status,
			NameAndTag: nameAndTag,
		}
		payload.EventData.Replication.SuccessfulArtifact = []*ctlModel.ArtifactInfo{succeedArtifact}
	}
	if event.Status == string(job.ErrorStatus) {
		failedArtifact := &ctlModel.ArtifactInfo{
			Type:       task.ResourceType,
			Status:     task.Status,
			NameAndTag: nameAndTag,
		}
		payload.EventData.Replication.FailedArtifact = []*ctlModel.ArtifactInfo{failedArtifact}
	}

	prj, err := project.Ctl.GetByName(orm.Context(), prjName, project.Metadata(true))
	if err != nil {
		log.Errorf("failed to get project %s, error: %v", prjName, err)
		return nil, nil, err
	}

	return payload, prj, nil
}

func getMetadataFromResource(resource string) (namespace, nameAndTag string) {
	// Usually resource format likes 'library/busybox:v1', but it could be 'busybox:v1' in docker registry
	meta := strings.Split(resource, "/")
	if len(meta) == 1 {
		return "", meta[0]
	}
	return meta[0], meta[1]
}
