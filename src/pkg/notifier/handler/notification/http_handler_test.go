package notification

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/goharbor/harbor/src/common/job/models"
	cModels "github.com/goharbor/harbor/src/common/models"
	"github.com/goharbor/harbor/src/pkg/notification"
	"github.com/goharbor/harbor/src/pkg/notifier/event"
	"github.com/goharbor/harbor/src/pkg/notifier/model"
	"github.com/stretchr/testify/require"
)

type fakedHookManager struct {
}

func (f *fakedHookManager) StartHook(event *model.HookEvent, job *models.JobData) error {
	return nil
}

func TestHTTPHandler_Handle(t *testing.T) {
	hookMgr := notification.HookManager
	defer func() {
		notification.HookManager = hookMgr
	}()
	notification.HookManager = &fakedHookManager{}

	handler := &HTTPHandler{}

	type args struct {
		event *event.Event
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "HTTPHandler_Handle Want Error 1",
			args: args{
				event: &event.Event{
					Topic: "http",
					Data:  nil,
				},
			},
			wantErr: true,
		},
		{
			name: "HTTPHandler_Handle Want Error 2",
			args: args{
				event: &event.Event{
					Topic: "http",
					Data:  &model.EventData{},
				},
			},
			wantErr: true,
		},
		{
			name: "HTTPHandler_Handle 1",
			args: args{
				event: &event.Event{
					Topic: "http",
					Data: &model.HookEvent{
						PolicyID:  1,
						EventType: "pushImage",
						Target: &cModels.EventTarget{
							Type:    "http",
							Address: "http://127.0.0.1:8080",
						},
						Payload: &model.Payload{
							OccurAt: time.Now().Unix(),
						},
					},
				},
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := handler.Handle(tt.args.event.Data)
			if tt.wantErr {
				require.NotNil(t, err, "Error: %s", err)
				return
			}
		})
	}
}

func TestHTTPHandler_IsStateful(t *testing.T) {
	handler := &HTTPHandler{}
	assert.False(t, handler.IsStateful())
}
