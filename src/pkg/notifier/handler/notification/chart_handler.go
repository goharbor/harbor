package notification

import (
	"errors"
	"fmt"
	"github.com/goharbor/harbor/src/common/models"
	"github.com/goharbor/harbor/src/common/utils/log"
	"github.com/goharbor/harbor/src/core/config"
	"github.com/goharbor/harbor/src/pkg/notification"
	"github.com/goharbor/harbor/src/pkg/notifier/model"
)

// ChartPreprocessHandler preprocess chart event data
type ChartPreprocessHandler struct {
}

// Handle preprocess chart event data and then publish hook event
func (cph *ChartPreprocessHandler) Handle(value interface{}) error {
	// if global notification configured disabled, return directly
	if !config.NotificationEnable() {
		log.Debug("notification feature is not enabled")
		return nil
	}

	chartEvent, ok := value.(*model.ChartEvent)
	if !ok {
		return errors.New("invalid chart event type")
	}

	if chartEvent == nil || len(chartEvent.Versions) == 0 || len(chartEvent.ProjectName) == 0 || len(chartEvent.ChartName) == 0 {
		return fmt.Errorf("data miss in chart event: %v", chartEvent)
	}

	project, err := config.GlobalProjectMgr.Get(chartEvent.ProjectName)
	if err != nil {
		log.Errorf("failed to find project[%s] for chart event: %v", chartEvent.ProjectName, err)
		return err
	}
	if project == nil {
		return fmt.Errorf("project not found for chart event: %s", chartEvent.ProjectName)
	}
	policies, err := notification.PolicyMgr.GetRelatedPolices(project.ProjectID, chartEvent.EventType)
	if err != nil {
		log.Errorf("failed to find policy for %s event: %v", chartEvent.EventType, err)
		return err
	}
	// if cannot find policy including event type in project, return directly
	if len(policies) == 0 {
		log.Debugf("cannot find policy for %s event: %v", chartEvent.EventType, chartEvent)
		return nil
	}

	payload, err := constructChartPayload(chartEvent, project)
	if err != nil {
		return err
	}

	err = sendHookWithPolicies(policies, payload, chartEvent.EventType)
	if err != nil {
		return err
	}

	return nil
}

// IsStateful ...
func (cph *ChartPreprocessHandler) IsStateful() bool {
	return false
}

func constructChartPayload(event *model.ChartEvent, project *models.Project) (*model.Payload, error) {
	repoType := models.ProjectPrivate
	if project.IsPublic() {
		repoType = models.ProjectPublic
	}

	payload := &model.Payload{
		Type:    event.EventType,
		OccurAt: event.OccurAt.Unix(),
		EventData: &model.EventData{
			Repository: &model.Repository{
				Name:         event.ChartName,
				Namespace:    event.ProjectName,
				RepoFullName: fmt.Sprintf("%s/%s", event.ProjectName, event.ChartName),
				RepoType:     repoType,
			},
		},
		Operator: event.Operator,
	}

	extURL, err := config.ExtEndpoint()
	if err != nil {
		return nil, fmt.Errorf("get external endpoint failed: %v", err)
	}

	resourcePrefix := fmt.Sprintf("%s/chartrepo/%s/charts/%s", extURL, event.ProjectName, event.ChartName)
	for _, v := range event.Versions {
		resURL := fmt.Sprintf("%s-%s.tgz", resourcePrefix, v)

		resource := &model.Resource{
			Tag:         v,
			ResourceURL: resURL,
		}
		payload.EventData.Resources = append(payload.EventData.Resources, resource)
	}
	return payload, nil
}
