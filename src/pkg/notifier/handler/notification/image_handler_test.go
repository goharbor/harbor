package notification

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/goharbor/harbor/src/common/dao"
	"github.com/goharbor/harbor/src/common/models"
	"github.com/goharbor/harbor/src/core/config"
	"github.com/goharbor/harbor/src/pkg/notification"
	"github.com/goharbor/harbor/src/pkg/notifier/model"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/require"
)

type fakedNotificationPlyMgr struct {
}

func (f *fakedNotificationPlyMgr) Create(*models.NotificationPolicy) (int64, error) {
	return 0, nil
}

func (f *fakedNotificationPlyMgr) List(id int64) ([]*models.NotificationPolicy, error) {
	return nil, nil
}

func (f *fakedNotificationPlyMgr) Get(id int64) (*models.NotificationPolicy, error) {
	return nil, nil
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

func (f *fakedNotificationPlyMgr) GetRelatedPolices(id int64, eventType string) ([]*models.NotificationPolicy, error) {
	if id == 1 {
		return []*models.NotificationPolicy{
			{
				ID: 1,
				EventTypes: []string{
					model.EventTypePullImage,
					model.EventTypePushImage,
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
	if id == 2 {
		return nil, nil
	}
	return nil, errors.New("")
}

func TestMain(m *testing.M) {
	dao.PrepareTestForPostgresSQL()
	os.Exit(m.Run())
}

func TestImagePreprocessHandler_Handle(t *testing.T) {
	PolicyMgr := notification.PolicyMgr
	defer func() {
		notification.PolicyMgr = PolicyMgr
	}()
	notification.PolicyMgr = &fakedNotificationPlyMgr{}

	handler := &ImagePreprocessHandler{}
	config.Init()

	type args struct {
		data interface{}
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "ImagePreprocessHandler Want Error 1",
			args: args{
				data: nil,
			},
			wantErr: true,
		},
		{
			name: "ImagePreprocessHandler Want Error 2",
			args: args{
				data: &model.ImageEvent{},
			},
			wantErr: true,
		},
		{
			name: "ImagePreprocessHandler Want Error 3",
			args: args{
				data: &model.ImageEvent{
					Resource: []*model.ImgResource{
						{
							Tag: "v1.0",
						},
					},
					Project: &models.Project{
						ProjectID: 3,
					},
				},
			},
			wantErr: true,
		},
		{
			name: "ImagePreprocessHandler Want Error 4",
			args: args{
				data: &model.ImageEvent{
					Resource: []*model.ImgResource{
						{
							Tag: "v1.0",
						},
					},
					Project: &models.Project{
						ProjectID: 1,
					},
				},
			},
			wantErr: true,
		},
		// No handlers registered for handling topic http
		{
			name: "ImagePreprocessHandler Want Error 5",
			args: args{
				data: &model.ImageEvent{
					RepoName: "test/alpine",
					Resource: []*model.ImgResource{
						{
							Tag: "v1.0",
						},
					},
					Project: &models.Project{
						ProjectID: 1,
					},
				},
			},
			wantErr: true,
		},
		{
			name: "ImagePreprocessHandler 2",
			args: args{
				data: &model.ImageEvent{
					Resource: []*model.ImgResource{
						{
							Tag: "v1.0",
						},
					},
					Project: &models.Project{
						ProjectID: 2,
					},
				},
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := handler.Handle(tt.args.data)
			if tt.wantErr {
				require.NotNil(t, err, "Error: %s", err)
				return
			}
			assert.Nil(t, err)
		})
	}
}

func TestImagePreprocessHandler_IsStateful(t *testing.T) {
	handler := &ImagePreprocessHandler{}
	assert.False(t, handler.IsStateful())
}
