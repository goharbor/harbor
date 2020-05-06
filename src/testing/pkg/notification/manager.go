package notification

import (
	"github.com/goharbor/harbor/src/common/models"
	"github.com/goharbor/harbor/src/controller/event"
)

type FakedPolicyMgr struct {
}

func (f *FakedPolicyMgr) Create(*models.NotificationPolicy) (int64, error) {
	return 0, nil
}

func (f *FakedPolicyMgr) List(id int64) ([]*models.NotificationPolicy, error) {
	return nil, nil
}

func (f *FakedPolicyMgr) Get(id int64) (*models.NotificationPolicy, error) {
	return nil, nil
}

func (f *FakedPolicyMgr) GetByNameAndProjectID(string, int64) (*models.NotificationPolicy, error) {
	return nil, nil
}

func (f *FakedPolicyMgr) Update(*models.NotificationPolicy) error {
	return nil
}

func (f *FakedPolicyMgr) Delete(int64) error {
	return nil
}

func (f *FakedPolicyMgr) Test(*models.NotificationPolicy) error {
	return nil
}

func (f *FakedPolicyMgr) GetRelatedPolices(id int64, eventType string) ([]*models.NotificationPolicy, error) {
	return []*models.NotificationPolicy{
		{
			ID: 1,
			EventTypes: []string{
				event.TopicUploadChart,
				event.TopicDownloadChart,
				event.TopicDeleteChart,
				event.TopicPushArtifact,
				event.TopicPullArtifact,
				event.TopicDeleteArtifact,
				event.TopicScanningFailed,
				event.TopicScanningCompleted,
				event.TopicQuotaExceed,
			},
			Targets: []models.EventTarget{
				{
					Type:    "http",
					Address: "http://127.0.0.1:8080",
				},
			},
		},
	}, nil
}
