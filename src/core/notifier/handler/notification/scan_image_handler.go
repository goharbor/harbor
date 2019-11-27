package notification

import (
	"time"

	"github.com/goharbor/harbor/src/common/models"
	"github.com/goharbor/harbor/src/common/utils/log"
	"github.com/goharbor/harbor/src/core/config"
	"github.com/goharbor/harbor/src/core/notifier/model"
	"github.com/goharbor/harbor/src/pkg/notification"
	"github.com/goharbor/harbor/src/pkg/scan/api/scan"
	v1 "github.com/goharbor/harbor/src/pkg/scan/rest/v1"
	"github.com/pkg/errors"
)

// ScanImagePreprocessHandler preprocess chart event data
type ScanImagePreprocessHandler struct {
}

// Handle preprocess chart event data and then publish hook event
func (si *ScanImagePreprocessHandler) Handle(value interface{}) error {
	// if global notification configured disabled, return directly
	if !config.NotificationEnable() {
		log.Debug("notification feature is not enabled")
		return nil
	}

	if value == nil {
		return errors.New("empty scan image event")
	}

	e, ok := value.(*model.ScanImageEvent)
	if !ok {
		return errors.New("invalid scan image event type")
	}

	policies, err := notification.PolicyMgr.GetRelatedPolices(e.Artifact.NamespaceID, e.EventType)
	if err != nil {
		return errors.Wrap(err, "scan preprocess handler")
	}

	// If we cannot find policy including event type in project, return directly
	if len(policies) == 0 {
		log.Debugf("Cannot find policy for %s event: %v", e.EventType, e)
		return nil
	}

	// Get project
	project, err := config.GlobalProjectMgr.Get(e.Artifact.NamespaceID)
	if err != nil {
		return errors.Wrap(err, "scan preprocess handler")
	}

	payload, err := constructScanImagePayload(e, project)
	if err != nil {
		return errors.Wrap(err, "scan preprocess handler")
	}

	err = sendHookWithPolicies(policies, payload, e.EventType)
	if err != nil {
		return errors.Wrap(err, "scan preprocess handler")
	}

	return nil
}

// IsStateful ...
func (si *ScanImagePreprocessHandler) IsStateful() bool {
	return false
}

func constructScanImagePayload(event *model.ScanImageEvent, project *models.Project) (*model.Payload, error) {
	repoType := models.ProjectPrivate
	if project.IsPublic() {
		repoType = models.ProjectPublic
	}

	repoName := getNameFromImgRepoFullName(event.Artifact.Repository)

	payload := &model.Payload{
		Type:    event.EventType,
		OccurAt: event.OccurAt.Unix(),
		EventData: &model.EventData{
			Repository: &model.Repository{
				Name:         repoName,
				Namespace:    project.Name,
				RepoFullName: event.Artifact.Repository,
				RepoType:     repoType,
			},
		},
		Operator: event.Operator,
	}

	extURL, err := config.ExtURL()
	if err != nil {
		return nil, errors.Wrap(err, "construct scan payload")
	}

	resURL, err := buildImageResourceURL(extURL, event.Artifact.Repository, event.Artifact.Tag)
	if err != nil {
		return nil, errors.Wrap(err, "construct scan payload")
	}

	// Wait for reasonable time to make sure the report is ready
	// Interval=500ms and total time = 5s
	// If the report is still not ready in the total time, then failed at then
	for i := 0; i < 10; i++ {
		// First check in case it is ready
		if re, err := scan.DefaultController.GetReport(event.Artifact, []string{v1.MimeTypeNativeReport}); err == nil {
			if len(re) > 0 && len(re[0].Report) > 0 {
				break
			}
		} else {
			log.Error(errors.Wrap(err, "construct scan payload: wait report ready loop"))
		}

		time.Sleep(500 * time.Millisecond)
	}

	// Add scan overview
	summaries, err := scan.DefaultController.GetSummary(event.Artifact, []string{v1.MimeTypeNativeReport})
	if err != nil {
		return nil, errors.Wrap(err, "construct scan payload")
	}

	resource := &model.Resource{
		Tag:          event.Artifact.Tag,
		Digest:       event.Artifact.Digest,
		ResourceURL:  resURL,
		ScanOverview: summaries,
	}
	payload.EventData.Resources = append(payload.EventData.Resources, resource)

	return payload, nil
}
