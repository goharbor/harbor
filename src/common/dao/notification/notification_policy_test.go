package notification

import (
	"testing"
	"time"

	"github.com/goharbor/harbor/src/common/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var (
	testPly1 = &models.NotificationPolicy{
		Name:         "webhook test policy1",
		Description:  "webhook test policy1 description",
		ProjectID:    111,
		TargetsDB:    "[{\"type\":\"http\",\"address\":\"http://10.173.32.58:9009\",\"token\":\"xxxxxxxxx\",\"skip_cert_verify\":true}]",
		EventTypesDB: "[\"pushImage\",\"pullImage\",\"deleteImage\",\"uploadChart\",\"deleteChart\",\"downloadChart\",\"scanningFailed\",\"scanningCompleted\"]",
		Creator:      "no one",
		CreationTime: time.Now(),
		UpdateTime:   time.Now(),
		Enabled:      true,
	}
)

var (
	testPly2 = &models.NotificationPolicy{
		Name:         "webhook test policy2",
		Description:  "webhook test policy2 description",
		ProjectID:    111,
		TargetsDB:    "[{\"type\":\"http\",\"address\":\"http://10.173.32.58:9009\",\"token\":\"xxxxxxxxx\",\"skip_cert_verify\":true}]",
		EventTypesDB: "[\"pushImage\",\"pullImage\",\"deleteImage\",\"uploadChart\",\"deleteChart\",\"downloadChart\",\"scanningFailed\",\"scanningCompleted\"]",
		Creator:      "no one",
		CreationTime: time.Now(),
		UpdateTime:   time.Now(),
		Enabled:      true,
	}
)

var (
	testPly3 = &models.NotificationPolicy{
		Name:         "webhook test policy3",
		Description:  "webhook test policy3 description",
		ProjectID:    111,
		TargetsDB:    "[{\"type\":\"http\",\"address\":\"http://10.173.32.58:9009\",\"token\":\"xxxxxxxxx\",\"skip_cert_verify\":true}]",
		EventTypesDB: "[\"pushImage\",\"pullImage\",\"deleteImage\",\"uploadChart\",\"deleteChart\",\"downloadChart\",\"scanningFailed\",\"scanningCompleted\"]",
		Creator:      "no one",
		CreationTime: time.Now(),
		UpdateTime:   time.Now(),
		Enabled:      true,
	}
)

func TestAddNotificationPolicy(t *testing.T) {
	tests := []struct {
		name    string
		policy  *models.NotificationPolicy
		want    int64
		wantErr bool
	}{
		{name: "AddNotificationPolicy nil", policy: nil, wantErr: true},
		{name: "AddNotificationPolicy 1", policy: testPly1, want: 1},
		{name: "AddNotificationPolicy 2", policy: testPly2, want: 2},
		{name: "AddNotificationPolicy 3", policy: testPly3, want: 3},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := AddNotificationPolicy(tt.policy)

			if tt.wantErr {
				require.NotNil(t, err, "wantErr: %s", err)
				return
			}
			require.Nil(t, err)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestGetNotificationPolicies(t *testing.T) {
	tests := []struct {
		name         string
		projectID    int64
		wantPolicies []*models.NotificationPolicy
		wantErr      bool
	}{
		{name: "GetNotificationPolicies nil", projectID: 0, wantPolicies: []*models.NotificationPolicy{}},
		{name: "GetNotificationPolicies 1", projectID: 111, wantPolicies: []*models.NotificationPolicy{testPly1, testPly2, testPly3}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotPolicies, err := GetNotificationPolicies(tt.projectID)
			if tt.wantErr {
				require.NotNil(t, err, "wantErr: %s", err)
				return
			}

			require.Nil(t, err)
			for i, gotPolicy := range gotPolicies {
				assert.Equal(t, tt.wantPolicies[i].Name, gotPolicy.Name)
				assert.Equal(t, tt.wantPolicies[i].ID, gotPolicy.ID)
				assert.Equal(t, tt.wantPolicies[i].EventTypesDB, gotPolicy.EventTypesDB)
				assert.Equal(t, tt.wantPolicies[i].TargetsDB, gotPolicy.TargetsDB)
				assert.Equal(t, tt.wantPolicies[i].Creator, gotPolicy.Creator)
				assert.Equal(t, tt.wantPolicies[i].Enabled, gotPolicy.Enabled)
				assert.Equal(t, tt.wantPolicies[i].Description, gotPolicy.Description)
			}
		})
	}
}

func TestGetNotificationPolicy(t *testing.T) {
	tests := []struct {
		name       string
		id         int64
		wantPolicy *models.NotificationPolicy
		wantErr    bool
	}{
		{name: "GetRepPolicy 1", id: 1, wantPolicy: testPly1},
		{name: "GetRepPolicy 2", id: 2, wantPolicy: testPly2},
		{name: "GetRepPolicy 3", id: 3, wantPolicy: testPly3},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotPolicy, err := GetNotificationPolicy(tt.id)
			if tt.wantErr {
				require.NotNil(t, err, "wantErr: %s", err)
				return
			}
			require.Nil(t, err)
			assert.Equal(t, tt.wantPolicy.Name, gotPolicy.Name)
			assert.Equal(t, tt.wantPolicy.ID, gotPolicy.ID)
			assert.Equal(t, tt.wantPolicy.EventTypesDB, gotPolicy.EventTypesDB)
			assert.Equal(t, tt.wantPolicy.TargetsDB, gotPolicy.TargetsDB)
			assert.Equal(t, tt.wantPolicy.Creator, gotPolicy.Creator)
			assert.Equal(t, tt.wantPolicy.Enabled, gotPolicy.Enabled)
			assert.Equal(t, tt.wantPolicy.Description, gotPolicy.Description)
		})
	}
}

func TestGetNotificationPolicyByName(t *testing.T) {
	type args struct {
		name      string
		projectID int64
	}
	tests := []struct {
		name       string
		args       args
		wantPolicy *models.NotificationPolicy
		wantErr    bool
	}{
		{name: "GetNotificationPolicyByName 1", args: args{name: testPly1.Name, projectID: testPly1.ProjectID}, wantPolicy: testPly1},
		{name: "GetNotificationPolicyByName 2", args: args{name: testPly2.Name, projectID: testPly2.ProjectID}, wantPolicy: testPly2},
		{name: "GetNotificationPolicyByName 3", args: args{name: testPly3.Name, projectID: testPly3.ProjectID}, wantPolicy: testPly3},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotPolicy, err := GetNotificationPolicyByName(tt.args.name, tt.args.projectID)
			if tt.wantErr {
				require.NotNil(t, err, "wantErr: %s", err)
				return
			}
			require.Nil(t, err)
			assert.Equal(t, tt.wantPolicy.Name, gotPolicy.Name)
			assert.Equal(t, tt.wantPolicy.ID, gotPolicy.ID)
			assert.Equal(t, tt.wantPolicy.EventTypesDB, gotPolicy.EventTypesDB)
			assert.Equal(t, tt.wantPolicy.TargetsDB, gotPolicy.TargetsDB)
			assert.Equal(t, tt.wantPolicy.Creator, gotPolicy.Creator)
			assert.Equal(t, tt.wantPolicy.Enabled, gotPolicy.Enabled)
			assert.Equal(t, tt.wantPolicy.Description, gotPolicy.Description)
		})
	}

}

func TestUpdateNotificationPolicy(t *testing.T) {
	type args struct {
		policy *models.NotificationPolicy
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "UpdateNotificationPolicy nil",
			args: args{
				policy: nil,
			},
			wantErr: true,
		},

		{
			name: "UpdateNotificationPolicy 1",
			args: args{
				policy: &models.NotificationPolicy{
					ID:           1,
					Name:         "webhook test policy1 new",
					Description:  "webhook test policy1 description new",
					ProjectID:    111,
					TargetsDB:    "[{\"type\":\"http\",\"address\":\"http://10.173.32.58:9009\",\"token\":\"xxxxxxxxx\",\"skip_cert_verify\":true}]",
					EventTypesDB: "[\"pushImage\",\"pullImage\",\"deleteImage\",\"uploadChart\",\"deleteChart\",\"downloadChart\",\"scanningFailed\",\"scanningCompleted\"]",
					Creator:      "no one",
					CreationTime: time.Now(),
					UpdateTime:   time.Now(),
					Enabled:      true,
				},
			},
		},
		{
			name: "UpdateNotificationPolicy 2",
			args: args{
				policy: &models.NotificationPolicy{
					ID:           1,
					Name:         "webhook test policy2 new",
					Description:  "webhook test policy2 description new",
					ProjectID:    111,
					TargetsDB:    "[{\"type\":\"http\",\"address\":\"http://10.173.32.58:9009\",\"token\":\"xxxxxxxxx\",\"skip_cert_verify\":true}]",
					EventTypesDB: "[\"pushImage\",\"pullImage\",\"deleteImage\",\"uploadChart\",\"deleteChart\",\"downloadChart\",\"scanningFailed\",\"scanningCompleted\"]",
					Creator:      "no one",
					CreationTime: time.Now(),
					UpdateTime:   time.Now(),
					Enabled:      true,
				},
			},
		},
		{
			name: "UpdateNotificationPolicy 3",
			args: args{
				policy: &models.NotificationPolicy{
					ID:           1,
					Name:         "webhook test policy3 new",
					Description:  "webhook test policy3 description new",
					ProjectID:    111,
					TargetsDB:    "[{\"type\":\"http\",\"address\":\"http://10.173.32.58:9009\",\"token\":\"xxxxxxxxx\",\"skip_cert_verify\":true}]",
					EventTypesDB: "[\"pushImage\",\"pullImage\",\"deleteImage\",\"uploadChart\",\"deleteChart\",\"downloadChart\",\"scanningFailed\",\"scanningCompleted\"]",
					Creator:      "no one",
					CreationTime: time.Now(),
					UpdateTime:   time.Now(),
					Enabled:      true,
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := UpdateNotificationPolicy(tt.args.policy)

			if tt.wantErr {
				require.NotNil(t, err, "Error: %s", err)
				return
			}

			require.Nil(t, err)
			gotPolicy, err := GetNotificationPolicy(tt.args.policy.ID)

			require.Nil(t, err)
			assert.Equal(t, tt.args.policy.Description, gotPolicy.Description)
			assert.Equal(t, tt.args.policy.Name, gotPolicy.Name)
		})
	}

}

func TestDeleteNotificationPolicy(t *testing.T) {
	tests := []struct {
		name    string
		id      int64
		wantErr bool
	}{
		{name: "DeleteNotificationPolicy 1", id: 1, wantErr: false},
		{name: "DeleteNotificationPolicy 2", id: 2, wantErr: false},
		{name: "DeleteNotificationPolicy 3", id: 3, wantErr: false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := DeleteNotificationPolicy(tt.id)
			if tt.wantErr {
				require.NotNil(t, err, "wantErr: %s", err)
				return
			}
			require.Nil(t, err)
			policy, err := GetNotificationPolicy(tt.id)
			require.Nil(t, err)
			assert.Nil(t, policy)
		})
	}
}
