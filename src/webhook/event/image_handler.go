package event

import (
	"errors"
	"fmt"
	"strings"

	"github.com/goharbor/harbor/src/common/dao"
	"github.com/goharbor/harbor/src/common/models"
	"github.com/goharbor/harbor/src/common/utils/log"
	"github.com/goharbor/harbor/src/core/config"
	"github.com/goharbor/harbor/src/core/notifier"
	"github.com/goharbor/harbor/src/webhook/hook"
	"github.com/goharbor/harbor/src/webhook/model"
)

// ImageWebhookHandler handle image webhook event
type ImageWebhookHandler struct {
}

// Handle image related webhook event
func (iwh *ImageWebhookHandler) Handle(value interface{}) error {
	if !config.WebhookEnable() {
		log.Debug("webhook feature is not enabled")
		return nil
	}

	if value == nil {
		return errors.New("ImageWebhookHandler cannot handle nil value")
	}

	imgEvent, ok := value.(*ImageEvent)
	if !ok || imgEvent == nil {
		return errors.New("invalid image webhook event")
	}

	if len(imgEvent.Events) == 0 {
		return fmt.Errorf("empty image webhook event: %v", imgEvent)
	}

	payload, err := iwh.constructImagePayload(imgEvent)
	if err != nil {
		return err
	}

	for _, event := range imgEvent.Events {
		tag := event.Target.Tag
		digest := event.Target.Digest

		if tag == "" || digest == "" {
			log.Errorf("invalid webhook event: %v", event)
			continue
		}

		resURL, err := getImageResourceURL(imgEvent.RepoName, tag)
		if err != nil {
			log.Errorf("get resource URL failed: %v", err)
			continue
		}

		eventData := &model.EventData{
			Tag:         tag,
			Digest:      digest,
			ResourceURL: resURL,
		}
		payload.EventData = append(payload.EventData, eventData)
	}

	policies, err := getRelatedPolices(imgEvent.ProjectID, imgEvent.HookType)
	if err != nil {
		return err
	}
	if len(policies) == 0 {
		log.Debug("cannot find policy for image event: %v", imgEvent)
		return nil
	}

	for _, ply := range policies {
		targets := ply.Targets
		for _, target := range targets {
			if err := notifier.Publish(target.Type, &hook.ScheduleItem{
				PolicyID: ply.ID,
				Target:   &target,
				Payload:  payload,
				HookType: imgEvent.HookType,
			}); err != nil {
				return fmt.Errorf("failed to publish webhook topic by %s: %v", target.Type, err)
			}
			log.Debugf("published image event %s by topic %s", payload.Type, target.Type)
		}
	}
	return nil
}

// IsStateful ...
func (iwh *ImageWebhookHandler) IsStateful() bool {
	return false
}

func (iwh *ImageWebhookHandler) constructImagePayload(event *ImageEvent) (*model.Payload, error) {
	repoName := event.RepoName
	if repoName == "" {
		return nil, errors.New("invalid webhook event")
	}

	repoType := "private"
	if event.ProjectPublic {
		repoType = "public"
	}

	repoRecord, err := getRepository(event.ProjectID, repoName)
	if err != nil {
		return nil, err
	}

	imageName := getNameFromRepoFullName(repoName)

	payload := &model.Payload{
		Type:      event.HookType,
		OccurAt:   event.OccurAt.Unix(),
		MediaType: MediaTypeContainerImage,
		Repository: &model.Repository{
			DateCreated:  repoRecord.CreationTime.Unix(),
			Name:         imageName,
			Namespace:    event.ProjectName,
			RepoFullName: repoName,
			RepoType:     repoType,
		},
		Operator: event.Operator,
	}

	return payload, nil
}

func getRepository(projectID int64, repoName string) (*models.RepoRecord, error) {
	query := &models.RepositoryQuery{
		ProjectIDs: []int64{projectID},
		Name:       repoName,
	}
	repositories, err := dao.GetRepositories(query)
	if err != nil {
		log.Errorf("get repository failed projectID %d, name %s: %v", projectID, repoName, err)
		return nil, err
	}
	if len(repositories) == 0 {
		return nil, fmt.Errorf("get empty repository projectID %d, name %s", projectID, repoName)
	}
	return repositories[0], nil
}

func getNameFromRepoFullName(repo string) string {
	idx := strings.Index(repo, "/")
	return repo[idx+1:]
}

func getImageResourceURL(repoName, tag string) (string, error) {
	extURL, err := config.ExtURL()
	if err != nil {
		return "", fmt.Errorf("get external endpoint failed: %v", err)
	}

	resURL := fmt.Sprintf("%s/%s:%s", extURL, repoName, tag)
	return resURL, nil
}
