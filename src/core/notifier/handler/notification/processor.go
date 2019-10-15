package notification

import (
	"errors"
	"fmt"
	"strings"

	"github.com/goharbor/harbor/src/common/dao"
	"github.com/goharbor/harbor/src/common/models"
	"github.com/goharbor/harbor/src/common/utils/log"
	"github.com/goharbor/harbor/src/core/config"
	"github.com/goharbor/harbor/src/core/notifier/event"
	notifyModel "github.com/goharbor/harbor/src/core/notifier/model"
	"github.com/goharbor/harbor/src/pkg/notification"
)

// getNameFromImgRepoFullName gets image name from repo full name with format `repoName/imageName`
func getNameFromImgRepoFullName(repo string) string {
	idx := strings.Index(repo, "/")
	return repo[idx+1:]
}

func buildImageResourceURL(extURL, repoName, tag string) (string, error) {
	resURL := fmt.Sprintf("%s/%s:%s", extURL, repoName, tag)
	return resURL, nil
}

func constructImagePayload(event *notifyModel.ImageEvent) (*notifyModel.Payload, error) {
	repoName := event.RepoName
	if repoName == "" {
		return nil, fmt.Errorf("invalid %s event with empty repo name", event.EventType)
	}

	repoType := models.ProjectPrivate
	if event.Project.IsPublic() {
		repoType = models.ProjectPublic
	}

	imageName := getNameFromImgRepoFullName(repoName)

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
			log.Errorf("invalid notification event with empty tag: %v", event)
			continue
		}

		resURL, err := buildImageResourceURL(extURL, event.RepoName, tag)
		if err != nil {
			log.Errorf("get resource URL failed: %v", err)
			continue
		}

		resource := &notifyModel.Resource{
			Tag:         tag,
			Digest:      digest,
			ResourceURL: resURL,
		}
		payload.EventData.Resources = append(payload.EventData.Resources, resource)
	}

	return payload, nil
}

// send hook by publishing topic of specified target type(notify type)
func sendHookWithPolicies(policies []*models.NotificationPolicy, payload *notifyModel.Payload, eventType string) error {
	errRet := false
	for _, ply := range policies {
		targets := ply.Targets
		for _, target := range targets {
			evt := &event.Event{}
			hookMetadata := &event.HookMetaData{
				EventType: eventType,
				PolicyID:  ply.ID,
				Payload:   payload,
				Target:    &target,
			}
			// It should never affect evaluating other policies when one is failed, but error should return
			if err := evt.Build(hookMetadata); err == nil {
				if err := evt.Publish(); err != nil {
					errRet = true
					log.Errorf("failed to publish hook notify event: %v", err)
				}
			} else {
				errRet = true
				log.Errorf("failed to build hook notify event metadata: %v", err)
			}
			log.Debugf("published image event %s by topic %s", payload.Type, target.Type)
		}
	}
	if errRet {
		return errors.New("failed to trigger some of the events")
	}
	return nil
}

func resolveImageEventData(value interface{}) (*notifyModel.ImageEvent, error) {
	imgEvent, ok := value.(*notifyModel.ImageEvent)
	if !ok || imgEvent == nil {
		return nil, errors.New("invalid image event")
	}

	if len(imgEvent.Resource) == 0 {
		return nil, fmt.Errorf("empty event resouece data in image event: %v", imgEvent)
	}

	return imgEvent, nil
}

// preprocessAndSendImageHook preprocess image event data and send hook by notification policy target
func preprocessAndSendImageHook(value interface{}) error {
	// if global notification configured disabled, return directly
	if !config.NotificationEnable() {
		log.Debug("notification feature is not enabled")
		return nil
	}

	imgEvent, err := resolveImageEventData(value)
	if err != nil {
		return err
	}

	policies, err := notification.PolicyMgr.GetRelatedPolices(imgEvent.Project.ProjectID, imgEvent.EventType)
	if err != nil {
		log.Errorf("failed to find policy for %s event: %v", imgEvent.EventType, err)
		return err
	}
	// if cannot find policy including event type in project, return directly
	if len(policies) == 0 {
		log.Debugf("cannot find policy for %s event: %v", imgEvent.EventType, imgEvent)
		return nil
	}

	payload, err := constructImagePayload(imgEvent)
	if err != nil {
		return err
	}

	err = sendHookWithPolicies(policies, payload, imgEvent.EventType)
	if err != nil {
		return err
	}

	return nil
}
