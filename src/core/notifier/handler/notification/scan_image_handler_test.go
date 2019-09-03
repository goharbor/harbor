package notification

import (
	"github.com/goharbor/harbor/src/common/dao"
	"github.com/goharbor/harbor/src/common/models"
	"github.com/goharbor/harbor/src/core/config"
	"github.com/goharbor/harbor/src/core/notifier/model"
	"github.com/goharbor/harbor/src/pkg/notification"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"testing"
	"time"
)

func TestScanImagePreprocessHandler_Handle(t *testing.T) {
	PolicyMgr := notification.PolicyMgr
	defer func() {
		notification.PolicyMgr = PolicyMgr
	}()
	notification.PolicyMgr = &fakedPolicyMgr{}

	handler := &ScanImagePreprocessHandler{}
	config.Init()

	name := "project_for_test_scanning_event_preprocess"
	id, _ := config.GlobalProjectMgr.Create(&models.Project{
		Name:    name,
		OwnerID: 1,
		Metadata: map[string]string{
			models.ProMetaEnableContentTrust:   "true",
			models.ProMetaPreventVul:           "true",
			models.ProMetaSeverity:             "low",
			models.ProMetaReuseSysCVEWhitelist: "false",
		},
	})
	defer func(id int64) {
		if err := config.GlobalProjectMgr.Delete(id); err != nil {
			t.Logf("failed to delete project %d: %v", id, err)
		}
	}(id)

	jID, _ := dao.AddScanJob(models.ScanJob{
		Status:       "finished",
		Repository:   "project_for_test_scanning_event_preprocess/testrepo",
		Tag:          "v1.0.0",
		Digest:       "sha256:5a539a2c733ca9efcd62d4561b36ea93d55436c5a86825b8e43ce8303a7a0752",
		CreationTime: time.Now(),
		UpdateTime:   time.Now(),
	})

	type args struct {
		data interface{}
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "ScanImagePreprocessHandler Want Error 1",
			args: args{
				data: nil,
			},
			wantErr: true,
		},
		{
			name: "ScanImagePreprocessHandler Want Error 2",
			args: args{
				data: &model.ScanImageEvent{},
			},
			wantErr: true,
		},
		{
			name: "ScanImagePreprocessHandler Want Error 3",
			args: args{
				data: &model.ScanImageEvent{
					JobID: jID + 1000,
				},
			},
			wantErr: true,
		},
		{
			name: "ScanImagePreprocessHandler Want Error 4",
			args: args{
				data: &model.ScanImageEvent{
					JobID: jID,
				},
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := handler.Handle(tt.args.data)
			if tt.wantErr {
				require.NotNil(t, err, "Error: %v", err)
				return
			}
			assert.Nil(t, err)
		})
	}
}

func TestScanImagePreprocessHandler_IsStateful(t *testing.T) {
	handler := &ScanImagePreprocessHandler{}
	assert.False(t, handler.IsStateful())
}
