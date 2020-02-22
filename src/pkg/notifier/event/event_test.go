package event

import (
	"testing"
	"time"

	"github.com/goharbor/harbor/src/common/models"
	notifierModel "github.com/goharbor/harbor/src/pkg/notifier/model"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestImagePushEvent_Build(t *testing.T) {
	type args struct {
		imgPushMetadata *ImagePushMetaData
		hookMetadata    *HookMetaData
	}

	tests := []struct {
		name    string
		args    args
		wantErr bool
		want    *Event
	}{
		{
			name: "Build Image Push Event",
			args: args{
				imgPushMetadata: &ImagePushMetaData{
					Project:  &models.Project{ProjectID: 1, Name: "library"},
					Tag:      "v1.0",
					Digest:   "abcd",
					OccurAt:  time.Now(),
					Operator: "admin",
					RepoName: "library/alpine",
				},
			},
			want: &Event{
				Topic: notifierModel.PushImageTopic,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			event := &Event{}
			err := event.Build(tt.args.imgPushMetadata)
			if tt.wantErr {
				require.NotNil(t, err, "Error: %s", err)
				return
			}
			assert.Equal(t, tt.want.Topic, event.Topic)
		})
	}
}

func TestImagePullEvent_Build(t *testing.T) {
	type args struct {
		imgPullMetadata *ImagePullMetaData
	}

	tests := []struct {
		name    string
		args    args
		wantErr bool
		want    *Event
	}{
		{
			name: "Build Image Pull Event",
			args: args{
				imgPullMetadata: &ImagePullMetaData{
					Project:  &models.Project{ProjectID: 1, Name: "library"},
					Tag:      "v1.0",
					Digest:   "abcd",
					OccurAt:  time.Now(),
					Operator: "admin",
					RepoName: "library/alpine",
				},
			},
			want: &Event{
				Topic: notifierModel.PullImageTopic,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			event := &Event{}
			err := event.Build(tt.args.imgPullMetadata)
			if tt.wantErr {
				require.NotNil(t, err, "Error: %s", err)
				return
			}
			assert.Equal(t, tt.want.Topic, event.Topic)
		})
	}
}

func TestImageDelEvent_Build(t *testing.T) {
	type args struct {
		imgDelMetadata *ImageDelMetaData
	}

	tests := []struct {
		name    string
		args    args
		wantErr bool
		want    *Event
	}{
		{
			name: "Build Image Delete Event",
			args: args{
				imgDelMetadata: &ImageDelMetaData{
					Project:  &models.Project{ProjectID: 1, Name: "library"},
					Tags:     []string{"v1.0"},
					OccurAt:  time.Now(),
					Operator: "admin",
					RepoName: "library/alpine",
				},
			},
			want: &Event{
				Topic: notifierModel.DeleteImageTopic,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			event := &Event{}
			err := event.Build(tt.args.imgDelMetadata)
			if tt.wantErr {
				require.NotNil(t, err, "Error: %s", err)
				return
			}
			assert.Equal(t, tt.want.Topic, event.Topic)
		})
	}
}

func TestHookEvent_Build(t *testing.T) {
	type args struct {
		hookMetadata *HookMetaData
	}

	tests := []struct {
		name    string
		args    args
		wantErr bool
		want    *Event
	}{
		{
			name: "Build HTTP Hook Event",
			args: args{
				hookMetadata: &HookMetaData{
					PolicyID:  1,
					EventType: "pushImage",
					Target: &models.EventTarget{
						Type:    "http",
						Address: "http://127.0.0.1",
					},
					Payload: nil,
				},
			},
			want: &Event{
				Topic: notifierModel.WebhookTopic,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			event := &Event{}
			err := event.Build(tt.args.hookMetadata)
			if tt.wantErr {
				require.NotNil(t, err, "Error: %s", err)
				return
			}
			assert.Equal(t, tt.want.Topic, event.Topic)
		})
	}
}

func TestEvent_Publish(t *testing.T) {
	type args struct {
		event *Event
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "Publish Error 1",
			args: args{
				event: &Event{
					Topic: notifierModel.WebhookTopic,
					Data:  nil,
				},
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.args.event.Publish()
			if tt.wantErr {
				require.NotNil(t, err, "Error: %s", err)
				return
			}
		})
	}
}
