package event

import (
	"github.com/goharbor/harbor/src/common/models"
	notifierModel "github.com/goharbor/harbor/src/pkg/notifier/model"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"testing"
)

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
