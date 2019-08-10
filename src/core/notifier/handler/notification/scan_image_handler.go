package notification

import (
	"errors"
	"fmt"
	"github.com/goharbor/harbor/src/common/dao"
	"github.com/goharbor/harbor/src/common/models"
	"github.com/goharbor/harbor/src/common/utils/log"
	"github.com/goharbor/harbor/src/core/config"
	"github.com/goharbor/harbor/src/core/notifier/model"
	"github.com/goharbor/harbor/src/pkg/notification"
	"strings"
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

	e, ok := value.(*model.ScanImageEvent)
	if !ok {
		return errors.New("invalid scan image event type")
	}

	if e == nil {
		return errors.New("empty scan image event")
	}

	job, err := dao.GetScanJob(e.JobID)
	if err != nil {
		log.Errorf("failed to find scan job[%d] for scanning webhook: %v", e.JobID, err)
		return err
	}
	if job == nil {
		log.Errorf("can't find scan job[%d] for scanning webhook", e.JobID)
		return fmt.Errorf("scan job for scanning webhook not found: %d", e.JobID)
	}

	rs := strings.SplitN(job.Repository, "/", 2)
	projectName := rs[0]
	repoName := rs[1]

	project, err := config.GlobalProjectMgr.Get(projectName)
	if err != nil {
		log.Errorf("failed to find project[%s] for scan image event: %v", projectName, err)
		return err
	}
	policies, err := notification.PolicyMgr.GetRelatedPolices(project.ProjectID, e.EventType)
	if err != nil {
		log.Errorf("failed to find policy for %s event: %v", e.EventType, err)
		return err
	}
	// if cannot find policy including event type in project, return directly
	if len(policies) == 0 {
		log.Debugf("cannot find policy for %s event: %v", e.EventType, e)
		return nil
	}

	payload, err := constructScanImagePayload(e, job, project, projectName, repoName)
	if err != nil {
		return err
	}

	err = sendHookWithPolicies(policies, payload, e.EventType)
	if err != nil {
		return err
	}

	return nil
}

// IsStateful ...
func (si *ScanImagePreprocessHandler) IsStateful() bool {
	return false
}

func constructScanImagePayload(event *model.ScanImageEvent, job *models.ScanJob, project *models.Project, projectName, repoName string) (*model.Payload, error) {
	repoType := models.ProjectPrivate
	if project.IsPublic() {
		repoType = models.ProjectPublic
	}

	payload := &model.Payload{
		Type:    event.EventType,
		OccurAt: event.OccurAt.Unix(),
		EventData: &model.EventData{
			Repository: &model.Repository{
				Name:         repoName,
				Namespace:    projectName,
				RepoFullName: job.Repository,
				RepoType:     repoType,
			},
		},
		Operator: event.Operator,
	}

	extURL, err := config.ExtURL()
	if err != nil {
		return nil, fmt.Errorf("get external endpoint failed: %v", err)
	}
	resURL, _ := buildImageResourceURL(extURL, job.Repository, job.Tag)

	// Add scan overview
	scanOverview := getScanOverview(job.Digest, job.Tag, event.EventType)
	if scanOverview == nil {
		scanOverview = &models.ImgScanOverview{
			JobID:        job.ID,
			Status:       job.Status,
			CreationTime: job.CreationTime,
		}
	}
	resource := &model.Resource{
		Tag:          job.Tag,
		Digest:       job.Digest,
		ResourceURL:  resURL,
		ScanOverview: scanOverview,
	}
	payload.EventData.Resources = append(payload.EventData.Resources, resource)
	return payload, nil
}
