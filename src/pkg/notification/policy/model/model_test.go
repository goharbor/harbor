package model

import (
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestPolicy_ConvertFromDBModel(t *testing.T) {
	tests := []struct {
		name    string
		policy  *Policy
		want    *Policy
		wantErr bool
	}{
		{
			name: "ConvertFromDBModel want error 1",
			policy: &Policy{
				TargetsDB: "[{{\"type\":\"http\",\"address\":\"http://10.173.32.58:9009\"}]",
			},
			wantErr: true,
		},
		{
			name: "ConvertFromDBModel want error 2",
			policy: &Policy{
				EventTypesDB: "[{\"pushImage\",\"pullImage\",\"deleteImage\"]",
			},
			wantErr: true,
		},
		{
			name: "ConvertFromDBModel 1",
			policy: &Policy{
				TargetsDB:    "[{\"type\":\"http\",\"address\":\"http://10.173.32.58:9009\"}]",
				EventTypesDB: "[\"pushImage\",\"pullImage\",\"deleteImage\"]",
			},
			want: &Policy{
				Targets: []EventTarget{
					{
						Type:    "http",
						Address: "http://10.173.32.58:9009",
					},
				},
				EventTypes: []string{"pushImage", "pullImage", "deleteImage"},
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

func TestPolicy_ConvertToDBModel(t *testing.T) {
	tests := []struct {
		name    string
		policy  *Policy
		want    *Policy
		wantErr bool
	}{
		{
			name: "ConvertToDBModel 1",
			policy: &Policy{
				Targets: []EventTarget{
					{
						Type:           "http",
						Address:        "http://127.0.0.1",
						SkipCertVerify: false,
					},
				},
				EventTypes: []string{"pushImage", "pullImage", "deleteImage"},
			},
			want: &Policy{
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
