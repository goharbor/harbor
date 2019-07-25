package webhook

import (
	"errors"
	"fmt"
	"strings"

	"github.com/goharbor/harbor/src/common/dao"
	"github.com/goharbor/harbor/src/common/models"
	"github.com/goharbor/harbor/src/common/utils/log"
	"github.com/goharbor/harbor/src/core/config"
	"github.com/goharbor/harbor/src/core/notifier"
	"github.com/goharbor/harbor/src/core/notifier/event"
	notifyEvt "github.com/goharbor/harbor/src/core/notifier/event"
	"github.com/goharbor/harbor/src/webhook"
	"github.com/goharbor/harbor/src/webhook/hook"
	"github.com/goharbor/harbor/src/webhook/model"
)

func getNameFromRepoFullName(repo string) string {
	idx := strings.Index(repo, "/")
	return repo[idx+1:]
}

func buildImageResourceURL(extURL, repoName, tag string) (string, error) {
	resURL := fmt.Sprintf("%s/%s:%s", extURL, repoName, tag)
	return resURL, nil
}

func constructImagePayload(event *event.ImageEvent, hookType string) (*model.Payload, error) {
	repoName := event.RepoName
	if repoName == "" {
		return nil, fmt.Errorf("invalid %s event with empty repo name", hookType)
	}

	repoType := models.ProMetaPrivate
	if event.Project.IsPublic() {
		repoType = models.ProMetaPublic
	}

	imageName := getNameFromRepoFullName(repoName)

	payload := &model.Payload{
		Type:    hookType,
		OccurAt: event.OccurAt.Unix(),
		EventData: &model.EventData{
			Repository: &model.Repository{
				Name:         imageName,
				Namespace:    event.Project.Name,
				RepoFullName: repoName,
				RepoType:     repoType,
			},
		},
		Operator: event.Operator,
	}

	repoRecord, err := dao.GetRepositoryByName(repoName)
	if err != nil {
		log.Errorf("failed to get repository with name %s: %v", repoName, err)
		return nil, err
	}
	// once repo has been delete, cannot ensure to get repo record
	if repoRecord == nil {
		log.Debugf("cannot find repository info with repo %s", repoName)
	} else {
		payload.EventData.Repository.DateCreated = repoRecord.CreationTime.Unix()
	}

	extURL, err := config.ExtURL()
	if err != nil {
		return nil, fmt.Errorf("get external endpoint failed: %v", err)
	}

	for _, res := range event.Resource {
		tag := res.Tag
		digest := res.Digest

		if tag == "" {
			log.Errorf("invalid webhook event with empty tag: %v", event)
			continue
		}

		resURL, err := buildImageResourceURL(extURL, event.RepoName, tag)
		if err != nil {
			log.Errorf("get resource URL failed: %v", err)
			continue
		}

		resource := &model.Resource{
			Tag:         tag,
			Digest:      digest,
			ResourceURL: resURL,
		}
		payload.EventData.Resources = append(payload.EventData.Resources, resource)
	}

	return payload, nil
}

// send hook by publishing topic of specified target type
func sendHookWithPolicies(policies []*models.WebhookPolicy, payload *model.Payload, hookType string) error {
	for _, ply := range policies {
		targets := ply.Targets
		for _, target := range targets {
			// publish topic by target type, eg. http, email etc
			if err := notifier.Publish(target.Type, &hook.ScheduleItem{
				PolicyID: ply.ID,
				Target:   &target,
				Payload:  payload,
				HookType: hookType,
			}); err != nil {
				return fmt.Errorf("failed to publish webhook topic by %s: %v", target.Type, err)
			}
			log.Debugf("published image event %s by topic %s", payload.Type, target.Type)
		}
	}
	return nil
}

func resolveImageEventData(value interface{}) (*event.ImageEvent, error) {
	imgEvent, ok := value.(*notifyEvt.ImageEvent)
	if !ok || imgEvent == nil {
		return nil, errors.New("invalid image event")
	}

	if len(imgEvent.Resource) == 0 {
		return nil, fmt.Errorf("empty event resouece data in image event: %v", imgEvent)
	}

	return imgEvent, nil
}

// PreprocessAndSendImageHook preprocess image event data and send hook by webhook policy target
func PreprocessAndSendImageHook(hookType string, value interface{}) error {
	// if global webhook configured disabled, return directly
	if !config.WebhookEnable() {
		log.Debug("webhook feature is not enabled")
		return nil
	}

	imgEvent, err := resolveImageEventData(value)
	if err != nil {
		return err
	}

	policies, err := webhook.PolicyCtl.GetRelatedPolices(imgEvent.Project.ProjectID, hookType)
	if err != nil {
		log.Errorf("failed to find policy for %s event: %v", hookType, err)
		return err
	}
	// if cannot find policy including hook type in project, return directly
	if len(policies) == 0 {
		log.Debugf("cannot find policy for %s event: %v", hookType, imgEvent)
		return nil
	}

	payload, err := constructImagePayload(imgEvent, hookType)
	if err != nil {
		return err
	}

	err = sendHookWithPolicies(policies, payload, hookType)
	if err != nil {
		return err
	}

	return nil

}
