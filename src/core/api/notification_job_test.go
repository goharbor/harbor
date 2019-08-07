package api

import (
	"net/http"
	"testing"
	"time"

	"github.com/goharbor/harbor/src/common/models"
	"github.com/goharbor/harbor/src/pkg/notification"
	"github.com/goharbor/harbor/src/pkg/notification/model"
)

type fakedNotificationJobMgr struct {
}

func (f *fakedNotificationJobMgr) Create(job *models.NotificationJob) (int64, error) {
	return 1, nil
}

func (f *fakedNotificationJobMgr) List(...*models.NotificationJobQuery) (int64, []*models.NotificationJob, error) {
	return 0, nil, nil
}

func (f *fakedNotificationJobMgr) Update(job *models.NotificationJob, props ...string) error {
	return nil
}

func (f *fakedNotificationJobMgr) ListJobsGroupByEventType(policyID int64) ([]*models.NotificationJob, error) {
	return []*models.NotificationJob{
		{
			EventType:    model.EventTypePullImage,
			CreationTime: time.Now(),
		},
		{
			EventType:    model.EventTypeDeleteImage,
			CreationTime: time.Now(),
		},
	}, nil
}

func TestNotificationJobAPI_List(t *testing.T) {
	policyMgr := notification.PolicyMgr
	jobMgr := notification.JobMgr
	defer func() {
		notification.PolicyMgr = policyMgr
		notification.JobMgr = jobMgr
	}()
	notification.PolicyMgr = &fakedNotificationPlyMgr{}
	notification.JobMgr = &fakedNotificationJobMgr{}

	cases := []*codeCheckingCase{
		// 401
		{
			request: &testingRequest{
				method: http.MethodGet,
				url:    "/api/projects/1/webhook/jobs?policy_id=1",
			},
			code: http.StatusUnauthorized,
		},
		// 403
		{
			request: &testingRequest{
				method:     http.MethodGet,
				url:        "/api/projects/1/webhook/jobs?policy_id=1",
				credential: nonSysAdmin,
			},
			code: http.StatusForbidden,
		},
		// 400 policyID invalid
		{
			request: &testingRequest{
				method:     http.MethodGet,
				url:        "/api/projects/1/webhook/jobs?policy_id=0",
				credential: sysAdmin,
			},
			code: http.StatusBadRequest,
		},
		// 400 policyID not found
		{
			request: &testingRequest{
				method:     http.MethodGet,
				url:        "/api/projects/1/webhook/jobs?policy_id=123",
				credential: sysAdmin,
			},
			code: http.StatusBadRequest,
		},
		// 404 project not found
		{
			request: &testingRequest{
				method:     http.MethodGet,
				url:        "/api/projects/123/webhook/jobs?policy_id=1",
				credential: sysAdmin,
			},
			code: http.StatusNotFound,
		},
		// 200
		{
			request: &testingRequest{
				method:     http.MethodGet,
				url:        "/api/projects/1/webhook/jobs?policy_id=1",
				credential: sysAdmin,
			},
			code: http.StatusOK,
		},
	}
	runCodeCheckingCases(t, cases...)
}
