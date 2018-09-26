package notifier

import (
	"errors"
	"fmt"
	"net/http"

	"github.com/goharbor/harbor/src/common/dao"
	common_http "github.com/goharbor/harbor/src/common/http"
	"github.com/goharbor/harbor/src/common/job"
	"github.com/goharbor/harbor/src/common/models"
	common_utils "github.com/goharbor/harbor/src/common/utils"
	"github.com/goharbor/harbor/src/common/utils/log"
	"github.com/goharbor/harbor/src/core/utils"
)

const (
	// PolicyTypeDaily specify the policy type is "daily"
	PolicyTypeDaily = "daily"
	// PolicyTypeNone specify the policy type is "none"
	PolicyTypeNone = "none"
)

// ScanPolicyNotification is defined for pass the policy change data.
type ScanPolicyNotification struct {
	// Type is used to keep the scan policy type: "none","daily" and "refresh".
	Type string

	// DailyTime is used when the type is 'daily', the offset with UTC time 00:00.
	DailyTime int64
}

// ScanPolicyNotificationHandler is defined to handle the changes of scanning
// policy.
type ScanPolicyNotificationHandler struct{}

// IsStateful to indicate this handler is stateful.
func (s *ScanPolicyNotificationHandler) IsStateful() bool {
	// Policy change should be done one by one.
	return true
}

// Handle the policy change notification.
func (s *ScanPolicyNotificationHandler) Handle(value interface{}) error {
	notification, ok := value.(ScanPolicyNotification)
	if !ok {
		return errors.New("ScanPolicyNotificationHandler can not handle value with invalid type")
	}

	if notification.Type == PolicyTypeDaily {
		if err := cancelScanAllJobs(); err != nil {
			return fmt.Errorf("Failed to cancel scan_all jobs, error: %v", err)
		}
		h, m, s := common_utils.ParseOfftime(notification.DailyTime)
		cron := fmt.Sprintf("%d %d %d * * *", s, m, h)
		if err := utils.ScheduleScanAllImages(cron); err != nil {
			return fmt.Errorf("Failed to schedule scan_all job, error: %v", err)
		}
	} else if notification.Type == PolicyTypeNone {
		if err := cancelScanAllJobs(); err != nil {
			return fmt.Errorf("Failed to cancel scan_all jobs, error: %v", err)
		}
	} else {
		return fmt.Errorf("Notification type %s is not supported", notification.Type)
	}

	return nil
}

func cancelScanAllJobs(c ...job.Client) error {
	var client job.Client
	if c == nil || len(c) == 0 {
		client = utils.GetJobServiceClient()
	} else {
		client = c[0]
	}
	q := &models.AdminJobQuery{
		Name: job.ImageScanAllJob,
		Kind: job.JobKindPeriodic,
	}
	jobs, err := dao.GetAdminJobs(q)
	if err != nil {
		log.Errorf("Failed to query sheduled scan_all jobs, error: %v", err)
		return err
	}
	if len(jobs) > 1 {
		log.Warningf("Got more than one scheduled scan_all jobs: %+v", jobs)
	}
	for _, j := range jobs {
		if err := dao.DeleteAdminJob(j.ID); err != nil {
			log.Warningf("Failed to delete scan_all job from DB, job ID: %d, job UUID: %s, error: %v", j.ID, j.UUID, err)
		}
		if err := client.PostAction(j.UUID, job.JobActionStop); err != nil {
			if e, ok := err.(*common_http.Error); ok && e.Code == http.StatusNotFound {
				log.Warningf("scan_all job not found on jobservice, UUID: %s, skip", j.UUID)
			} else {
				log.Errorf("Failed to stop scan_all job, UUID: %s, error: %v", j.UUID, e)
				return e
			}
		}
		log.Infof("scan_all job canceled, uuid: %s, id: %d", j.UUID, j.ID)
	}
	return nil
}
