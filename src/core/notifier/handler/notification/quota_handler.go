package notification

import (
	"errors"
	"fmt"
	"github.com/goharbor/harbor/src/common/models"
	"github.com/goharbor/harbor/src/common/utils/log"
	"github.com/goharbor/harbor/src/core/config"
	"github.com/goharbor/harbor/src/core/notifier/model"
	notifyModel "github.com/goharbor/harbor/src/core/notifier/model"
	"github.com/goharbor/harbor/src/pkg/notification"
)

// QuotaPreprocessHandler preprocess image event data
type QuotaPreprocessHandler struct {
}

// Handle ...
func (qp *QuotaPreprocessHandler) Handle(value interface{}) error {
	if !config.NotificationEnable() {
		log.Debug("notification feature is not enabled")
		return nil
	}

	quotaEvent, ok := value.(*model.QuotaEvent)
	if !ok {
		return errors.New("invalid quota event type")
	}
	if quotaEvent == nil {
		return fmt.Errorf("nil quota event")
	}

	project, err := config.GlobalProjectMgr.Get(quotaEvent.Project.Name)
	if err != nil {
		log.Errorf("failed to get project:%s, error: %v", quotaEvent.Project.Name, err)
		return err
	}
	if project == nil {
		return fmt.Errorf("project not found of quota event: %s", quotaEvent.Project.Name)
	}
	policies, err := notification.PolicyMgr.GetRelatedPolices(project.ProjectID, quotaEvent.EventType)
	if err != nil {
		log.Errorf("failed to find policy for %s event: %v", quotaEvent.EventType, err)
		return err
	}
	if len(policies) == 0 {
		log.Debugf("cannot find policy for %s event: %v", quotaEvent.EventType, quotaEvent)
		return nil
	}

	payload, err := constructQuotaPayload(quotaEvent)
	if err != nil {
		return err
	}

	err = sendHookWithPolicies(policies, payload, quotaEvent.EventType)
	if err != nil {
		return err
	}
	return nil
}

// IsStateful ...
func (qp *QuotaPreprocessHandler) IsStateful() bool {
	return false
}

func constructQuotaPayload(event *model.QuotaEvent) (*model.Payload, error) {
	repoName := event.RepoName
	if repoName == "" {
		return nil, fmt.Errorf("invalid %s event with empty repo name", event.EventType)
	}

	repoType := models.ProjectPrivate
	if event.Project.IsPublic() {
		repoType = models.ProjectPublic
	}

	imageName := getNameFromImgRepoFullName(repoName)
	quotaCustom := make(map[string]string)
	quotaCustom["Details"] = event.Msg

	payload := &notifyModel.Payload{
		Type:    event.EventType,
		OccurAt: event.OccurAt.Unix(),
		EventData: &notifyModel.EventData{
			Repository: &notifyModel.Repository{
				Name:         imageName,
				Namespace:    event.Project.Name,
				RepoFullName: repoName,
				RepoType:     repoType,
			},
			Custom: quotaCustom,
		},
	}
	resource := &notifyModel.Resource{
		Tag:    event.Resource.Tag,
		Digest: event.Resource.Digest,
	}
	payload.EventData.Resources = append(payload.EventData.Resources, resource)

	return payload, nil
}
