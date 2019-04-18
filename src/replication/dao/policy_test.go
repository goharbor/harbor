package dao

import (
	"testing"

	common_models "github.com/goharbor/harbor/src/common/models"
	"github.com/goharbor/harbor/src/replication/dao/models"
	"github.com/goharbor/harbor/src/replication/model"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var (
	testPolic1 = &models.RepPolicy{
		// ID:                999,
		Name:              "Policy Test 1",
		Description:       "Policy Description",
		Creator:           "someone",
		SrcRegistryID:     123,
		DestRegistryID:    456,
		DestNamespace:     "target_ns",
		ReplicateDeletion: true,
		Override:          true,
		Enabled:           true,
		Trigger:           "{\"type\":\"\",\"trigger_settings\":null}",
		Filters:           "[{\"type\":\"registry\",\"value\":\"abc\"}]",
	}

	testPolic2 = &models.RepPolicy{
		// ID:                999,
		Name:              "Policy Test 2",
		Description:       "Policy Description",
		Creator:           "someone",
		SrcRegistryID:     123,
		DestRegistryID:    456,
		DestNamespace:     "target_ns",
		ReplicateDeletion: true,
		Override:          true,
		Enabled:           true,
		Trigger:           "{\"type\":\"\",\"trigger_settings\":null}",
		Filters:           "[{\"type\":\"registry\",\"value\":\"abc\"}]",
	}

	testPolic3 = &models.RepPolicy{
		// ID:                999,
		Name:              "Policy Test 3",
		Description:       "Policy Description",
		Creator:           "someone",
		SrcRegistryID:     123,
		DestRegistryID:    456,
		DestNamespace:     "target_ns",
		ReplicateDeletion: true,
		Override:          true,
		Enabled:           true,
		Trigger:           "{\"type\":\"\",\"trigger_settings\":null}",
		Filters:           "[{\"type\":\"registry\",\"value\":\"abc\"}]",
	}
)

func TestAddRepPolicy(t *testing.T) {
	tests := []struct {
		name    string
		policy  *models.RepPolicy
		want    int64
		wantErr bool
	}{
		{name: "AddRepPolicy 1", policy: testPolic1, want: 1},
		{name: "AddRepPolicy 2", policy: testPolic2, want: 2},
		{name: "AddRepPolicy 3", policy: testPolic3, want: 3},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := AddRepPolicy(tt.policy)

			if tt.wantErr {
				require.NotNil(t, err, "wantErr: %s", err)
				return
			}

			require.Nil(t, err)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestGetPolicies(t *testing.T) {
	type args struct {
		name      string
		namespace string
		page      int64
		pageSize  int64
	}
	tests := []struct {
		name         string
		args         args
		wantPolicies []*models.RepPolicy
		wantErr      bool
	}{
		{name: "GetTotalOfRepPolicies nil", args: args{name: "Test 0"}, wantPolicies: []*models.RepPolicy{}},
		{name: "GetTotalOfRepPolicies 1", args: args{name: "Test 1"}, wantPolicies: []*models.RepPolicy{testPolic1}},
		{name: "GetTotalOfRepPolicies 2", args: args{name: "Test", page: 1, pageSize: 2}, wantPolicies: []*models.RepPolicy{testPolic1, testPolic2}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, gotPolicies, err := GetPolicies([]*model.PolicyQuery{
				{
					Name:      tt.args.name,
					Namespace: tt.args.namespace,
					Pagination: common_models.Pagination{
						Page: tt.args.page,
						Size: tt.args.pageSize,
					},
				},
			}...)
			if tt.wantErr {
				require.NotNil(t, err, "wantErr: %s", err)
				return
			}

			require.Nil(t, err)
			for i, gotPolicy := range gotPolicies {
				assert.Equal(t, tt.wantPolicies[i].Name, gotPolicy.Name)
				assert.Equal(t, tt.wantPolicies[i].Description, gotPolicy.Description)
				assert.Equal(t, tt.wantPolicies[i].Creator, gotPolicy.Creator)
				assert.Equal(t, tt.wantPolicies[i].SrcRegistryID, gotPolicy.SrcRegistryID)
				assert.Equal(t, tt.wantPolicies[i].DestRegistryID, gotPolicy.DestRegistryID)
				assert.Equal(t, tt.wantPolicies[i].DestNamespace, gotPolicy.DestNamespace)
				assert.Equal(t, tt.wantPolicies[i].ReplicateDeletion, gotPolicy.ReplicateDeletion)
				assert.Equal(t, tt.wantPolicies[i].Override, gotPolicy.Override)
				assert.Equal(t, tt.wantPolicies[i].Enabled, gotPolicy.Enabled)
				assert.Equal(t, tt.wantPolicies[i].Trigger, gotPolicy.Trigger)
				assert.Equal(t, tt.wantPolicies[i].Filters, gotPolicy.Filters)
			}
		})
	}
}

func TestGetRepPolicy(t *testing.T) {
	tests := []struct {
		name       string
		id         int64
		wantPolicy *models.RepPolicy
		wantErr    bool
	}{
		{name: "GetRepPolicy 1", id: 1, wantPolicy: testPolic1},
		{name: "GetRepPolicy 2", id: 2, wantPolicy: testPolic2},
		{name: "GetRepPolicy 3", id: 3, wantPolicy: testPolic3},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotPolicy, err := GetRepPolicy(tt.id)
			if tt.wantErr {
				require.NotNil(t, err, "wantErr: %s", err)
				return
			}

			require.Nil(t, err)
			assert.Equal(t, tt.wantPolicy.Name, gotPolicy.Name)
			assert.Equal(t, tt.wantPolicy.Description, gotPolicy.Description)
			assert.Equal(t, tt.wantPolicy.Creator, gotPolicy.Creator)
			assert.Equal(t, tt.wantPolicy.SrcRegistryID, gotPolicy.SrcRegistryID)
			assert.Equal(t, tt.wantPolicy.DestRegistryID, gotPolicy.DestRegistryID)
			assert.Equal(t, tt.wantPolicy.DestNamespace, gotPolicy.DestNamespace)
			assert.Equal(t, tt.wantPolicy.ReplicateDeletion, gotPolicy.ReplicateDeletion)
			assert.Equal(t, tt.wantPolicy.Override, gotPolicy.Override)
			assert.Equal(t, tt.wantPolicy.Enabled, gotPolicy.Enabled)
			assert.Equal(t, tt.wantPolicy.Trigger, gotPolicy.Trigger)
			assert.Equal(t, tt.wantPolicy.Filters, gotPolicy.Filters)
		})
	}
}

func TestGetRepPolicyByName(t *testing.T) {
	type args struct {
		name string
	}
	tests := []struct {
		name       string
		args       args
		wantPolicy *models.RepPolicy
		wantErr    bool
	}{
		{name: "GetRepPolicyByName 1", args: args{name: testPolic1.Name}, wantPolicy: testPolic1},
		{name: "GetRepPolicyByName 2", args: args{name: testPolic2.Name}, wantPolicy: testPolic2},
		{name: "GetRepPolicyByName 3", args: args{name: testPolic3.Name}, wantPolicy: testPolic3},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotPolicy, err := GetRepPolicyByName(tt.args.name)
			if tt.wantErr {
				require.NotNil(t, err, "wantErr: %s", err)
				return
			}

			require.Nil(t, err)
			assert.Equal(t, tt.wantPolicy.Name, gotPolicy.Name)
			assert.Equal(t, tt.wantPolicy.Description, gotPolicy.Description)
			assert.Equal(t, tt.wantPolicy.Creator, gotPolicy.Creator)
			assert.Equal(t, tt.wantPolicy.SrcRegistryID, gotPolicy.SrcRegistryID)
			assert.Equal(t, tt.wantPolicy.DestRegistryID, gotPolicy.DestRegistryID)
			assert.Equal(t, tt.wantPolicy.DestNamespace, gotPolicy.DestNamespace)
			assert.Equal(t, tt.wantPolicy.ReplicateDeletion, gotPolicy.ReplicateDeletion)
			assert.Equal(t, tt.wantPolicy.Override, gotPolicy.Override)
			assert.Equal(t, tt.wantPolicy.Enabled, gotPolicy.Enabled)
			assert.Equal(t, tt.wantPolicy.Trigger, gotPolicy.Trigger)
			assert.Equal(t, tt.wantPolicy.Filters, gotPolicy.Filters)
		})
	}
}

func TestUpdateRepPolicy(t *testing.T) {
	type args struct {
		policy *models.RepPolicy
		props  []string
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{name: "UpdateRepPolicy Want Error", args: args{policy: nil}, wantErr: true},
		{
			name: "UpdateRepPolicy 1",
			args: args{
				policy: &models.RepPolicy{ID: 1, Description: "Policy Description 1", Creator: "Someone 1"},
				props:  []string{"description", "creator"},
			},
		},
		{
			name: "UpdateRepPolicy 2",
			args: args{
				policy: &models.RepPolicy{ID: 2, Description: "Policy Description 2", Creator: "Someone 2"},
				props:  []string{"description", "creator"},
			},
		},
		{
			name: "UpdateRepPolicy 3",
			args: args{
				policy: &models.RepPolicy{ID: 3, Description: "Policy Description 3", Creator: "Someone 3"},
				props:  []string{"description", "creator"},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := UpdateRepPolicy(tt.args.policy, tt.args.props...)

			if tt.wantErr {
				require.NotNil(t, err, "Error: %s", err)
				return
			}

			require.Nil(t, err)
			gotPolicy, err := GetRepPolicy(tt.args.policy.ID)

			require.Nil(t, err)
			assert.Equal(t, tt.args.policy.Description, gotPolicy.Description)
			assert.Equal(t, tt.args.policy.Creator, gotPolicy.Creator)
		})
	}
}

func TestDeleteRepPolicy(t *testing.T) {
	tests := []struct {
		name    string
		id      int64
		wantErr bool
	}{
		{name: "DeleteRepPolicy 1", id: 1, wantErr: false},
		{name: "DeleteRepPolicy 2", id: 2, wantErr: false},
		{name: "DeleteRepPolicy 3", id: 3, wantErr: false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := DeleteRepPolicy(tt.id)
			if tt.wantErr {
				require.NotNil(t, err, "wantErr: %s", err)
				return
			}

			require.Nil(t, err)
			policy, err := GetRepPolicy(tt.id)
			require.Nil(t, err)
			assert.Nil(t, policy)
		})
	}
}
