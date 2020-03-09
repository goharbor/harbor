package api

import (
	"net/http"
	"testing"

	"github.com/pkg/errors"

	"github.com/goharbor/harbor/src/pkg/notifier/model"

	"github.com/goharbor/harbor/src/common/models"
	"github.com/goharbor/harbor/src/pkg/notification"
)

type fakedNotificationPlyMgr struct {
}

func (f *fakedNotificationPlyMgr) Create(*models.NotificationPolicy) (int64, error) {
	return 0, nil
}

func (f *fakedNotificationPlyMgr) List(id int64) ([]*models.NotificationPolicy, error) {
	return []*models.NotificationPolicy{
		{
			ID: 1,
			EventTypes: []string{
				model.EventTypePullImage,
				model.EventTypePushImage,
			},
		},
	}, nil
}

func (f *fakedNotificationPlyMgr) Get(id int64) (*models.NotificationPolicy, error) {
	switch id {
	case 1:
		return &models.NotificationPolicy{ID: 1, ProjectID: 1}, nil
	case 2:
		return &models.NotificationPolicy{ID: 2, ProjectID: 222}, nil
	case 3:
		return nil, errors.New("")
	default:
		return nil, nil
	}
}

func (f *fakedNotificationPlyMgr) GetByNameAndProjectID(string, int64) (*models.NotificationPolicy, error) {
	return nil, nil
}

func (f *fakedNotificationPlyMgr) Update(*models.NotificationPolicy) error {

	return nil
}

func (f *fakedNotificationPlyMgr) Delete(int64) error {
	return nil
}

func (f *fakedNotificationPlyMgr) Test(*models.NotificationPolicy) error {
	return nil
}

func (f *fakedNotificationPlyMgr) GetRelatedPolices(int64, string) ([]*models.NotificationPolicy, error) {
	return nil, nil
}

func TestNotificationPolicyAPI_List(t *testing.T) {
	policyCtl := notification.PolicyMgr
	defer func() {
		notification.PolicyMgr = policyCtl
	}()

	notification.PolicyMgr = &fakedNotificationPlyMgr{}

	cases := []*codeCheckingCase{
		// 401
		{
			request: &testingRequest{
				method: http.MethodGet,
				url:    "/api/projects/1/webhook/policies",
			},
			code: http.StatusUnauthorized,
		},
		// 403
		{
			request: &testingRequest{
				method:     http.MethodGet,
				url:        "/api/projects/1/webhook/policies",
				credential: nonSysAdmin,
			},
			code: http.StatusForbidden,
		},
		// 404
		{
			request: &testingRequest{
				method:     http.MethodGet,
				url:        "/api/projects/123/webhook/policies",
				credential: sysAdmin,
			},
			code: http.StatusNotFound,
		},
		// 200
		{
			request: &testingRequest{
				method:     http.MethodGet,
				url:        "/api/projects/1/webhook/policies",
				credential: sysAdmin,
			},
			code: http.StatusOK,
		},
	}
	runCodeCheckingCases(t, cases...)

}

func TestNotificationPolicyAPI_Post(t *testing.T) {
	policyCtl := notification.PolicyMgr
	defer func() {
		notification.PolicyMgr = policyCtl
	}()

	notification.PolicyMgr = &fakedNotificationPlyMgr{}

	cases := []*codeCheckingCase{
		// 401
		{
			request: &testingRequest{
				method: http.MethodPost,
				url:    "/api/projects/1/webhook/policies",
			},
			code: http.StatusUnauthorized,
		},
		// 403
		{
			request: &testingRequest{
				method:     http.MethodPost,
				url:        "/api/projects/1/webhook/policies",
				credential: nonSysAdmin,
			},
			code: http.StatusForbidden,
		},
		// 400 invalid json body
		{
			request: &testingRequest{
				method:     http.MethodPost,
				url:        "/api/projects/1/webhook/policies",
				credential: sysAdmin,
				bodyJSON:   "invalid json body",
			},
			code: http.StatusBadRequest,
		},
		// 400 empty targets
		{
			request: &testingRequest{
				method:     http.MethodPost,
				url:        "/api/projects/1/webhook/policies",
				credential: sysAdmin,
				bodyJSON: &models.NotificationPolicy{
					Targets: []models.EventTarget{},
				}},
			code: http.StatusBadRequest,
		},
		// 400 invalid event target address
		{
			request: &testingRequest{
				method:     http.MethodPost,
				url:        "/api/projects/1/webhook/policies",
				credential: sysAdmin,
				bodyJSON: &models.NotificationPolicy{
					EventTypes: []string{"pullImage", "pushImage", "deleteImage"},
					Targets: []models.EventTarget{
						{
							Address: "tcp://127.0.0.1:8080",
						},
					},
				}},
			code: http.StatusBadRequest,
		},
		// 400 invalid event target type
		{
			request: &testingRequest{
				method:     http.MethodPost,
				url:        "/api/projects/1/webhook/policies",
				credential: sysAdmin,
				bodyJSON: &models.NotificationPolicy{
					EventTypes: []string{"pullImage", "pushImage", "deleteImage"},
					Targets: []models.EventTarget{
						{
							Type:    "smn",
							Address: "http://127.0.0.1:8080",
						},
					},
				}},
			code: http.StatusBadRequest,
		},
		// 400 invalid event type
		{
			request: &testingRequest{
				method:     http.MethodPost,
				url:        "/api/projects/1/webhook/policies",
				credential: sysAdmin,
				bodyJSON: &models.NotificationPolicy{
					EventTypes: []string{"invalidType"},
					Targets: []models.EventTarget{
						{
							Address: "tcp://127.0.0.1:8080",
						},
					},
				}},
			code: http.StatusBadRequest,
		},
		// 400 policy ID != 0
		{
			request: &testingRequest{
				method:     http.MethodPost,
				url:        "/api/projects/1/webhook/policies",
				credential: sysAdmin,
				bodyJSON: &models.NotificationPolicy{
					ID:         111,
					EventTypes: []string{"pullImage", "pushImage", "deleteImage"},
					Targets: []models.EventTarget{
						{
							Type:           "http",
							Address:        "http://10.173.32.58:9009",
							AuthHeader:     "xxxxxxxxx",
							SkipCertVerify: true,
						},
					},
				},
			},
			code: http.StatusBadRequest,
		},
		// 201
		{
			request: &testingRequest{
				method:     http.MethodPost,
				url:        "/api/projects/1/webhook/policies",
				credential: sysAdmin,
				bodyJSON: &models.NotificationPolicy{
					EventTypes: []string{"pullImage", "pushImage", "deleteImage"},
					Targets: []models.EventTarget{
						{
							Type:           "http",
							Address:        "http://10.173.32.58:9009",
							AuthHeader:     "xxxxxxxxx",
							SkipCertVerify: true,
						},
					},
				},
			},
			code: http.StatusCreated,
		},
	}
	runCodeCheckingCases(t, cases...)
}

func TestNotificationPolicyAPI_Get(t *testing.T) {
	policyCtl := notification.PolicyMgr
	defer func() {
		notification.PolicyMgr = policyCtl
	}()

	notification.PolicyMgr = &fakedNotificationPlyMgr{}

	cases := []*codeCheckingCase{
		// 401
		{
			request: &testingRequest{
				method: http.MethodGet,
				url:    "/api/projects/1/webhook/policies/111",
			},
			code: http.StatusUnauthorized,
		},
		// 403
		{
			request: &testingRequest{
				method:     http.MethodGet,
				url:        "/api/projects/1/webhook/policies/111",
				credential: nonSysAdmin,
			},
			code: http.StatusForbidden,
		},
		// 404
		{
			request: &testingRequest{
				method:     http.MethodGet,
				url:        "/api/projects/1/webhook/policies/1234",
				credential: sysAdmin,
			},
			code: http.StatusNotFound,
		},
		// 400 projectID not match with projectID in URL
		{
			request: &testingRequest{
				method:     http.MethodGet,
				url:        "/api/projects/1/webhook/policies/2",
				credential: sysAdmin,
			},
			code: http.StatusBadRequest,
		},
		// 500
		{
			request: &testingRequest{
				method:     http.MethodGet,
				url:        "/api/projects/1/webhook/policies/3",
				credential: sysAdmin,
			},
			code: http.StatusInternalServerError,
		},
		// 200
		{
			request: &testingRequest{
				method:     http.MethodGet,
				url:        "/api/projects/1/webhook/policies/1",
				credential: sysAdmin,
			},
			code: http.StatusOK,
		},
	}
	runCodeCheckingCases(t, cases...)
}

func TestNotificationPolicyAPI_Put(t *testing.T) {
	policyCtl := notification.PolicyMgr
	defer func() {
		notification.PolicyMgr = policyCtl
	}()

	notification.PolicyMgr = &fakedNotificationPlyMgr{}

	cases := []*codeCheckingCase{
		// 401
		{
			request: &testingRequest{
				method: http.MethodPut,
				url:    "/api/projects/1/webhook/policies/111",
			},
			code: http.StatusUnauthorized,
		},
		// 403
		{
			request: &testingRequest{
				method:     http.MethodPut,
				url:        "/api/projects/1/webhook/policies/111",
				credential: nonSysAdmin,
			},
			code: http.StatusForbidden,
		},
		// 404
		{
			request: &testingRequest{
				method:     http.MethodPut,
				url:        "/api/projects/1/webhook/policies/1234",
				credential: sysAdmin,
			},
			code: http.StatusNotFound,
		},
		// 400 invalid json body
		{
			request: &testingRequest{
				method:     http.MethodPut,
				url:        "/api/projects/1/webhook/policies/1",
				credential: sysAdmin,
				bodyJSON:   "invalidJSONBody",
			},
			code: http.StatusBadRequest,
		},
		// 400 empty targets
		{
			request: &testingRequest{
				method:     http.MethodPut,
				url:        "/api/projects/1/webhook/policies/1",
				credential: sysAdmin,
				bodyJSON: &models.NotificationPolicy{
					EventTypes: []string{"pullImage", "pushImage", "deleteImage"},
					Targets:    []models.EventTarget{},
				}},
			code: http.StatusBadRequest,
		},
		// 400 invalid event target address
		{
			request: &testingRequest{
				method:     http.MethodPut,
				url:        "/api/projects/1/webhook/policies/1",
				credential: sysAdmin,
				bodyJSON: &models.NotificationPolicy{
					EventTypes: []string{"pullImage", "pushImage", "deleteImage"},
					Targets: []models.EventTarget{
						{
							Address: "tcp://127.0.0.1:8080",
						},
					},
				}},
			code: http.StatusBadRequest,
		},
		// 400 invalid event target type
		{
			request: &testingRequest{
				method:     http.MethodPut,
				url:        "/api/projects/1/webhook/policies/1",
				credential: sysAdmin,
				bodyJSON: &models.NotificationPolicy{
					EventTypes: []string{"pullImage", "pushImage", "deleteImage"},
					Targets: []models.EventTarget{
						{
							Type:    "smn",
							Address: "http://127.0.0.1:8080",
						},
					},
				}},
			code: http.StatusBadRequest,
		},
		// 400 invalid event type
		{
			request: &testingRequest{
				method:     http.MethodPut,
				url:        "/api/projects/1/webhook/policies/1",
				credential: sysAdmin,
				bodyJSON: &models.NotificationPolicy{
					EventTypes: []string{"invalidType"},
					Targets: []models.EventTarget{
						{
							Address: "tcp://127.0.0.1:8080",
						},
					},
				}},
			code: http.StatusBadRequest,
		},
		// 200
		{
			request: &testingRequest{
				method:     http.MethodPut,
				url:        "/api/projects/1/webhook/policies/1",
				credential: sysAdmin,
				bodyJSON: &models.NotificationPolicy{
					Name:       "imagePolicyTest",
					EventTypes: []string{"pullImage", "pushImage", "deleteImage"},
					Targets: []models.EventTarget{
						{
							Type:           "http",
							Address:        "http://10.173.32.58:9009",
							AuthHeader:     "xxxxxxxxx",
							SkipCertVerify: true,
						},
					},
				},
			},
			code: http.StatusOK,
		},
	}
	runCodeCheckingCases(t, cases...)
}

func TestNotificationPolicyAPI_Test(t *testing.T) {
	policyCtl := notification.PolicyMgr
	defer func() {
		notification.PolicyMgr = policyCtl
	}()

	notification.PolicyMgr = &fakedNotificationPlyMgr{}
	cases := []*codeCheckingCase{
		// 401
		{
			request: &testingRequest{
				method: http.MethodPost,
				url:    "/api/projects/1/webhook/policies/test",
			},
			code: http.StatusUnauthorized,
		},
		// 403
		{
			request: &testingRequest{
				method:     http.MethodPost,
				url:        "/api/projects/1/webhook/policies/test",
				credential: nonSysAdmin,
			},
			code: http.StatusForbidden,
		},
		// 404
		{
			request: &testingRequest{
				method:     http.MethodPost,
				url:        "/api/projects/123/webhook/policies/test",
				credential: sysAdmin,
			},
			code: http.StatusNotFound,
		},
		// 400 invalid json body
		{
			request: &testingRequest{
				method:     http.MethodPost,
				url:        "/api/projects/1/webhook/policies/test",
				credential: sysAdmin,
				bodyJSON:   1234125,
			},
			code: http.StatusBadRequest,
		},
		// 200
		{
			request: &testingRequest{
				method:     http.MethodPost,
				url:        "/api/projects/1/webhook/policies/test",
				credential: sysAdmin,
				bodyJSON: &models.NotificationPolicy{
					Targets: []models.EventTarget{
						{
							Type:           "http",
							Address:        "http://10.173.32.58:9009",
							AuthHeader:     "xxxxxxxxx",
							SkipCertVerify: true,
						},
					},
				},
			},
			code: http.StatusOK,
		},
	}
	runCodeCheckingCases(t, cases...)
}

func TestNotificationPolicyAPI_ListGroupByEventType(t *testing.T) {
	policyCtl := notification.PolicyMgr
	jobMgr := notification.JobMgr
	defer func() {
		notification.PolicyMgr = policyCtl
		notification.JobMgr = jobMgr
	}()

	notification.PolicyMgr = &fakedNotificationPlyMgr{}
	notification.JobMgr = &fakedNotificationJobMgr{}

	cases := []*codeCheckingCase{
		// 401
		{
			request: &testingRequest{
				method: http.MethodGet,
				url:    "/api/projects/1/webhook/lasttrigger",
			},
			code: http.StatusUnauthorized,
		},
		// 403
		{
			request: &testingRequest{
				method:     http.MethodGet,
				url:        "/api/projects/1/webhook/lasttrigger",
				credential: nonSysAdmin,
			},
			code: http.StatusForbidden,
		},
		// 404
		{
			request: &testingRequest{
				method:     http.MethodGet,
				url:        "/api/projects/123/webhook/lasttrigger",
				credential: sysAdmin,
			},
			code: http.StatusNotFound,
		},
		// 200
		{
			request: &testingRequest{
				method:     http.MethodGet,
				url:        "/api/projects/1/webhook/lasttrigger",
				credential: sysAdmin,
			},
			code: http.StatusOK,
		},
	}
	runCodeCheckingCases(t, cases...)
}

func TestNotificationPolicyAPI_Delete(t *testing.T) {
	policyCtl := notification.PolicyMgr
	defer func() {
		notification.PolicyMgr = policyCtl
	}()

	notification.PolicyMgr = &fakedNotificationPlyMgr{}

	cases := []*codeCheckingCase{
		// 401
		{
			request: &testingRequest{
				method: http.MethodDelete,
				url:    "/api/projects/1/webhook/policies/111",
			},
			code: http.StatusUnauthorized,
		},
		// 403
		{
			request: &testingRequest{
				method:     http.MethodDelete,
				url:        "/api/projects/1/webhook/policies/111",
				credential: nonSysAdmin,
			},
			code: http.StatusForbidden,
		},
		// 404
		{
			request: &testingRequest{
				method:     http.MethodDelete,
				url:        "/api/projects/1/webhook/policies/1234",
				credential: sysAdmin,
			},
			code: http.StatusNotFound,
		},
		// 400 projectID not match
		{
			request: &testingRequest{
				method:     http.MethodDelete,
				url:        "/api/projects/1/webhook/policies/2",
				credential: sysAdmin,
			},
			code: http.StatusBadRequest,
		},
		// 500 failed to get policy
		{
			request: &testingRequest{
				method:     http.MethodDelete,
				url:        "/api/projects/1/webhook/policies/3",
				credential: sysAdmin,
			},
			code: http.StatusInternalServerError,
		},
		// 200
		{
			request: &testingRequest{
				method:     http.MethodDelete,
				url:        "/api/projects/1/webhook/policies/1",
				credential: sysAdmin,
			},
			code: http.StatusOK,
		},
	}
	runCodeCheckingCases(t, cases...)
}
