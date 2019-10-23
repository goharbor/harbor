package models

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNotificationPolicy_ConvertFromDBModel(t *testing.T) {
	tests := []struct {
		name    string
		policy  *NotificationPolicy
		want    *NotificationPolicy
		wantErr bool
	}{
		{
			name: "ConvertFromDBModel want error 1",
			policy: &NotificationPolicy{
				TargetsDB: "[{{\"type\":\"http\",\"address\":\"http://10.173.32.58:9009\"}]",
			},
			wantErr: true,
		},
		{
			name: "ConvertFromDBModel want error 2",
			policy: &NotificationPolicy{
				EventTypesDB: "[{\"pushImage\",\"pullImage\",\"deleteImage\"]",
			},
			wantErr: true,
		},
		{
			name: "ConvertFromDBModel 1",
			policy: &NotificationPolicy{
				TargetsDB:    "[{\"type\":\"http\",\"address\":\"http://10.173.32.58:9009\"}]",
				EventTypesDB: "[\"pushImage\",\"pullImage\",\"deleteImage\"]",
			},
			want: &NotificationPolicy{
				Targets: []EventTarget{
					{
						Type:    "http",
						Address: "http://10.173.32.58:9009",
					},
				},
				EventTypes: []EventType{{Type: "pullImage", Enable: true}, {Type: "pushImage", Enable: true},
					{Type: "deleteImage", Enable: true}},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.policy.ConvertFromDBModel()
			if tt.wantErr {
				require.NotNil(t, err, "wantErr: %s", err)
				return
			}
			require.Nil(t, err)
			assert.Equal(t, tt.want.Targets, tt.policy.Targets)
			assert.Equal(t, tt.want.EventTypes, tt.policy.EventTypes)
		})
	}
}

func TestNotificationPolicy_ConvertToDBModel(t *testing.T) {
	tests := []struct {
		name    string
		policy  *NotificationPolicy
		want    *NotificationPolicy
		wantErr bool
	}{
		{
			name: "ConvertToDBModel 1",
			policy: &NotificationPolicy{
				Targets: []EventTarget{
					{
						Type:           "http",
						Address:        "http://127.0.0.1",
						SkipCertVerify: false,
					},
				},
				EventTypes: []EventType{{Type: "pullImage", Enable: true}, {Type: "pushImage", Enable: true},
					{Type: "deleteImage", Enable: true}},
			},
			want: &NotificationPolicy{
				TargetsDB:    "[{\"type\":\"http\",\"address\":\"http://127.0.0.1\",\"skip_cert_verify\":false}]",
				EventTypesDB: "[\"pushImage\",\"pullImage\",\"deleteImage\"]",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.policy.ConvertToDBModel()
			if tt.wantErr {
				require.NotNil(t, err, "wantErr: %s", err)
				return
			}
			require.Nil(t, err)
			assert.Equal(t, tt.want.TargetsDB, tt.policy.TargetsDB)
			assert.Equal(t, tt.want.EventTypesDB, tt.policy.EventTypesDB)
		})
	}
}

func TestNotificationJob_TableName(t *testing.T) {
	job := &NotificationJob{}
	got := job.TableName()
	assert.Equal(t, NotificationJobTable, got)
}

func TestNotificationPolicy_TableName(t *testing.T) {
	policy := &NotificationPolicy{}
	got := policy.TableName()
	assert.Equal(t, NotificationPolicyTable, got)

}
