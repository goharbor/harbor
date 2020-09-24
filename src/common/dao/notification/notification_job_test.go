package notification

import (
	"testing"

	"github.com/goharbor/harbor/src/common/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var (
	testJob1 = &models.NotificationJob{
		PolicyID:   1111,
		EventType:  "pushImage",
		NotifyType: "http",
		Status:     "pending",
		JobDetail:  "{\"type\":\"pushImage\",\"occur_at\":1563536782,\"event_data\":{\"resources\":[{\"digest\":\"sha256:bf1684a6e3676389ec861c602e97f27b03f14178e5bc3f70dce198f9f160cce9\",\"tag\":\"v1.0\",\"resource_url\":\"10.194.32.23/myproj/alpine:v1.0\"}],\"repository\":{\"date_created\":1563505587,\"name\":\"alpine\",\"namespace\":\"myproj\",\"repo_full_name\":\"myproj/alpine\",\"repo_type\":\"private\"}},\"operator\":\"admin\"}",
		UUID:       "00000000",
	}
	testJob2 = &models.NotificationJob{
		PolicyID:   111,
		EventType:  "pullImage",
		NotifyType: "http",
		Status:     "",
		JobDetail:  "{\"type\":\"pushImage\",\"occur_at\":1563537782,\"event_data\":{\"resources\":[{\"digest\":\"sha256:bf1684a6e3676389ec861c602e97f27b03f14178e5bc3f70dce198f9f160cce9\",\"tag\":\"v1.0\",\"resource_url\":\"10.194.32.23/myproj/alpine:v1.0\"}],\"repository\":{\"date_created\":1563505587,\"name\":\"alpine\",\"namespace\":\"myproj\",\"repo_full_name\":\"myproj/alpine\",\"repo_type\":\"private\"}},\"operator\":\"admin\"}",
		UUID:       "00000000",
	}
	testJob3 = &models.NotificationJob{
		PolicyID:   111,
		EventType:  "deleteImage",
		NotifyType: "http",
		Status:     "pending",
		JobDetail:  "{\"type\":\"pushImage\",\"occur_at\":1563538782,\"event_data\":{\"resources\":[{\"digest\":\"sha256:bf1684a6e3676389ec861c602e97f27b03f14178e5bc3f70dce198f9f160cce9\",\"tag\":\"v1.0\",\"resource_url\":\"10.194.32.23/myproj/alpine:v1.0\"}],\"repository\":{\"date_created\":1563505587,\"name\":\"alpine\",\"namespace\":\"myproj\",\"repo_full_name\":\"myproj/alpine\",\"repo_type\":\"private\"}},\"operator\":\"admin\"}",
		UUID:       "00000000",
	}
)

func TestAddNotificationJob(t *testing.T) {
	tests := []struct {
		name    string
		job     *models.NotificationJob
		want    int64
		wantErr bool
	}{
		{name: "AddNotificationJob nil", job: nil, wantErr: true},
		{name: "AddNotificationJob 1", job: testJob1, want: 1},
		{name: "AddNotificationJob 2", job: testJob2, want: 2},
		{name: "AddNotificationJob 3", job: testJob3, want: 3},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := AddNotificationJob(tt.job)
			if tt.wantErr {
				require.NotNil(t, err, "wantErr: %s", err)
				return
			}
			require.Nil(t, err)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestGetTotalCountOfNotificationJobs(t *testing.T) {
	type args struct {
		query *models.NotificationJobQuery
	}
	tests := []struct {
		name    string
		args    args
		want    int64
		wantErr bool
	}{
		{
			name: "GetTotalCountOfNotificationJobs 1",
			args: args{
				query: &models.NotificationJobQuery{
					PolicyID: 111,
				},
			},
			want: 2,
		},
		{
			name: "GetTotalCountOfNotificationJobs 2",
			args: args{},
			want: 3,
		},
		{
			name: "GetTotalCountOfNotificationJobs 3",
			args: args{
				query: &models.NotificationJobQuery{
					Statuses: []string{"pending"},
				},
			},
			want: 3,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := GetTotalCountOfNotificationJobs(tt.args.query)
			if tt.wantErr {
				require.NotNil(t, err, "wantErr: %s", err)
				return
			}
			require.Nil(t, err)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestGetLastTriggerJobsGroupByEventType(t *testing.T) {
	type args struct {
		policyID int64
	}
	tests := []struct {
		name    string
		args    args
		want    []*models.NotificationJob
		wantErr bool
	}{
		{
			name: "GetLastTriggerJobsGroupByEventType",
			args: args{
				policyID: 111,
			},
			want: []*models.NotificationJob{
				testJob2,
				testJob3,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := GetLastTriggerJobsGroupByEventType(tt.args.policyID)
			if tt.wantErr {
				require.NotNil(t, err, "wantErr: %s", err)
				return
			}
			require.Nil(t, err)
			assert.Equal(t, len(tt.want), len(got))
		})
	}

}

func TestUpdateNotificationJob(t *testing.T) {
	type args struct {
		job   *models.NotificationJob
		props []string
	}
	tests := []struct {
		name    string
		args    args
		want    int64
		wantErr bool
	}{
		{name: "UpdateNotificationJob Want Error 1", args: args{job: nil}, wantErr: true},
		{name: "UpdateNotificationJob Want Error 2", args: args{job: &models.NotificationJob{ID: 0}}, wantErr: true},
		{
			name: "UpdateNotificationJob 1",
			args: args{
				job:   &models.NotificationJob{ID: 1, UUID: "111111111111111"},
				props: []string{"UUID"},
			},
		},
		{
			name: "UpdateNotificationJob 2",
			args: args{
				job:   &models.NotificationJob{ID: 2, UUID: "222222222222222"},
				props: []string{"UUID"},
			},
		},
		{
			name: "UpdateNotificationJob 3",
			args: args{
				job:   &models.NotificationJob{ID: 3, UUID: "333333333333333"},
				props: []string{"UUID"},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := UpdateNotificationJob(tt.args.job, tt.args.props...)
			if tt.wantErr {
				require.NotNil(t, err, "Error: %s", err)
				return
			}

			require.Nil(t, err)
			gotJob, err := GetNotificationJob(tt.args.job.ID)

			require.Nil(t, err)
			assert.Equal(t, tt.args.job.UUID, gotJob.UUID)
		})
	}
}

func TestDeleteNotificationJob(t *testing.T) {
	type args struct {
		id int64
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{name: "DeleteNotificationJob 1", args: args{id: 1}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := DeleteNotificationJob(tt.args.id)

			if tt.wantErr {
				require.NotNil(t, err, "Error: %s", err)
				return
			}

			require.Nil(t, err)
			job, err := GetNotificationJob(tt.args.id)

			require.Nil(t, err)
			assert.Nil(t, job)
		})
	}
}

func TestDeleteAllNotificationJobs(t *testing.T) {
	type args struct {
		policyID int64
		query    []*models.NotificationJobQuery
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "DeleteAllNotificationJobs 1",
			args: args{
				policyID: 111,
				query: []*models.NotificationJobQuery{
					{PolicyID: 111},
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := DeleteAllNotificationJobsByPolicyID(tt.args.policyID)

			if tt.wantErr {
				require.NotNil(t, err, "Error: %s", err)
				return
			}

			require.Nil(t, err)
			jobs, err := GetNotificationJobs(tt.args.query...)

			require.Nil(t, err)
			assert.Equal(t, 0, len(jobs))
		})
	}
}
