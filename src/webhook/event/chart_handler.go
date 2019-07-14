package event

import (
	"errors"
	"fmt"

	"github.com/goharbor/harbor/src/common/models"
	"github.com/goharbor/harbor/src/common/utils/log"
	"github.com/goharbor/harbor/src/core/config"
	"github.com/goharbor/harbor/src/core/notifier"
	"github.com/goharbor/harbor/src/webhook/hook"
	"github.com/goharbor/harbor/src/webhook/model"
)

// ChartWebhookHandler handle chart webhook event
type ChartWebhookHandler struct {
}

// Handle chart related webhook event
func (cwh *ChartWebhookHandler) Handle(value interface{}) error {
	if !config.WebhookEnable() {
		log.Debug("webhook feature is not enabled")
		return nil
	}

	if value == nil {
		return errors.New("ChartWebhookHandler cannot handle nil value")
	}

	crtEvent, ok := value.(*ChartEvent)
	if !ok || crtEvent == nil {
		return errors.New("invalid chart webhook event")
	}

	if len(crtEvent.ChartVersions) == 0 {
		return fmt.Errorf("empty chart version: %v", crtEvent)
	}
	projName := crtEvent.ProjectName
	chartName := crtEvent.ChartName
	if projName == "" || chartName == "" {
		return errors.New("project name or chart name cannot be empty")
	}

	proj, err := config.GlobalProjectMgr.Get(projName)
	if err != nil {
		log.Errorf("Get project %s failed: %v", projName, err)
		return err
	}
	payload, err := cwh.constructChartPayload(proj, crtEvent)
	if err != nil {
		return err
	}

	for _, ver := range crtEvent.ChartVersions {
		chartURL, err := getChartResourceURL(projName, chartName, ver)
		if err != nil {
			log.Errorf("Get chart %s version %s resource URL failed: %v", chartName, ver, err)
			return err
		}

		eventData := &model.EventData{
			Tag:         ver,
			ResourceURL: chartURL,
		}

		payload.EventData = append(payload.EventData, eventData)
	}

	policies, err := getRelatedPolices(proj.ProjectID, payload.Type)
	if err != nil {
		return err
	}
	if len(policies) == 0 {
		log.Debug("cannot find policy for chart event %v", crtEvent)
		return nil
	}

	for _, ply := range policies {
		targets := ply.Targets
		for _, target := range targets {
			if err := notifier.Publish(target.Type, &hook.ScheduleItem{
				PolicyID: ply.ID,
				Target:   &target,
				Payload:  payload,
				HookType: crtEvent.HookType,
			}); err != nil {
				return fmt.Errorf("failed to publish chart webhook topic by %s: %v", target.Type, err)
			}
			log.Debugf("published chart %s event by topic %s", payload.Type, target.Type)
		}
	}
	return nil
}

// IsStateful ...
func (cwh *ChartWebhookHandler) IsStateful() bool {
	return false
}

func (cwh *ChartWebhookHandler) constructChartPayload(proj *models.Project, event *ChartEvent) (*model.Payload, error) {
	repoType := "private"
	if proj.IsPublic() {
		repoType = "public"
	}

	payload := &model.Payload{
		Type:      event.HookType,
		OccurAt:   event.OccurTime.Unix(),
		MediaType: MediaTypeHelmChart,
		Repository: &model.Repository{
			Name:         event.ChartName,
			Namespace:    event.ProjectName,
			RepoFullName: fmt.Sprintf("%s/%s", event.ProjectName, event.ChartName),
			RepoType:     repoType,
			DateCreated:  event.RepoCreateTime.Unix(),
		},
		Operator: event.Operator,
	}
	return payload, nil
}

func getChartResourceURL(projName, repoName, version string) (string, error) {
	extURL, err := config.ExtURL()
	if err != nil {
		return "", fmt.Errorf("get external endpoint failed: %v", err)
	}
	resURL := fmt.Sprintf("%s/chartrepo/%s/charts/%s-%s.tgz", extURL, projName, repoName, version)
	return resURL, nil
}
