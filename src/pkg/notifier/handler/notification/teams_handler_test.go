package notification

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/goharbor/harbor/src/pkg/notification"
	policy_model "github.com/goharbor/harbor/src/pkg/notification/policy/model"
	"github.com/goharbor/harbor/src/pkg/notifier/event"
	"github.com/goharbor/harbor/src/pkg/notifier/model"
)

func TestTeamsHandler_Handle(t *testing.T) {
	hookMgr := notification.HookManager
	defer func() {
		notification.HookManager = hookMgr
	}()
	notification.HookManager = &fakedHookManager{}

	handler := &TeamsHandler{}

	type args struct {
		event *event.Event
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "TeamsHandler_Handle Want Error 1",
			args: args{
				event: &event.Event{
					Topic: "teams",
					Data:  nil,
				},
			},
			wantErr: true,
		},
		{
			name: "TeamsHandler_Handle Want Error 2",
			args: args{
				event: &event.Event{
					Topic: "teams",
					Data:  &model.EventData{},
				},
			},
			wantErr: true,
		},
		{
			name: "TeamsHandler_Handle 1",
			args: args{
				event: &event.Event{
					Topic: "teams",
					Data: &model.HookEvent{
						PolicyID:  1,
						EventType: "pushImage",
						Target: &policy_model.EventTarget{
							Type:    "teams",
							Address: "http://127.0.0.1:8080",
						},
						Payload: &model.Payload{
							OccurAt:  time.Now().Unix(),
							Type:     "pushImage",
							Operator: "admin",
							EventData: &model.EventData{
								Resources: []*model.Resource{
									{
										Tag: "v9.0",
									},
								},
								Repository: &model.Repository{
									Name: "library/debian",
								},
							},
						},
					},
				},
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := handler.Handle(context.TODO(), tt.args.event.Data)
			if tt.wantErr {
				require.NotNil(t, err, "Error: %s", err)
				return
			}
		})
	}
}

func TestTeamsHandler_IsStateful(t *testing.T) {
	handler := &TeamsHandler{}
	assert.False(t, handler.IsStateful())
}

func TestTeamsHandler_Name(t *testing.T) {
	handler := &TeamsHandler{}
	assert.Equal(t, "Teams", handler.Name())
}
