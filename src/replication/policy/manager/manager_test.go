// Copyright Project Harbor Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//    http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package manager

import (
	"reflect"
	"testing"

	persist_models "github.com/goharbor/harbor/src/replication/dao/models"
	"github.com/goharbor/harbor/src/replication/model"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func Test_convertFromPersistModel(t *testing.T) {
	tests := []struct {
		name    string
		from    *persist_models.RepPolicy
		want    *model.Policy
		wantErr bool
	}{
		{
			name:    "Nil Persist Model",
			from:    nil,
			want:    nil,
			wantErr: false,
		},
		{
			name: "parse Filters Error",
			from: &persist_models.RepPolicy{Filters: "abc"},
			want: nil, wantErr: true,
		},
		{
			name: "parse Trigger Error",
			from: &persist_models.RepPolicy{Trigger: "abc"},
			want: nil, wantErr: true,
		},
		{
			name: "Persist Model", from: &persist_models.RepPolicy{
				ID:                999,
				Name:              "Policy Test",
				Description:       "Policy Description",
				Creator:           "someone",
				SrcRegistryID:     123,
				DestRegistryID:    456,
				DestNamespace:     "target_ns",
				ReplicateDeletion: true,
				Override:          true,
				Enabled:           true,
				Trigger:           "",
				Filters:           "[]",
			}, want: &model.Policy{
				ID:          999,
				Name:        "Policy Test",
				Description: "Policy Description",
				Creator:     "someone",
				SrcRegistry: &model.Registry{
					ID: 123,
				},
				DestRegistry: &model.Registry{
					ID: 456,
				},
				DestNamespace: "target_ns",
				Deletion:      true,
				Override:      true,
				Enabled:       true,
				Trigger:       nil,
				Filters:       []*model.Filter{},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := convertFromPersistModel(tt.from)

			if tt.wantErr {
				require.NotNil(t, err)
				return
			}

			if tt.want == nil {
				require.Nil(t, got)
				return
			}

			require.Nil(t, err, tt.name)
			assert.Equal(t, tt.want.ID, got.ID)
			assert.Equal(t, tt.want.Name, got.Name)
			assert.Equal(t, tt.want.Description, got.Description)
			assert.Equal(t, tt.want.Creator, got.Creator)
			assert.Equal(t, tt.want.SrcRegistry.ID, got.SrcRegistry.ID)
			assert.Equal(t, tt.want.DestRegistry.ID, got.DestRegistry.ID)
			assert.Equal(t, tt.want.DestNamespace, got.DestNamespace)
			assert.Equal(t, tt.want.Deletion, got.Deletion)
			assert.Equal(t, tt.want.Override, got.Override)
			assert.Equal(t, tt.want.Enabled, got.Enabled)
			assert.Equal(t, tt.want.Trigger, got.Trigger)
			assert.Equal(t, tt.want.Filters, got.Filters)

		})
	}
}

func Test_convertToPersistModel(t *testing.T) {
	tests := []struct {
		name    string
		from    *model.Policy
		want    *persist_models.RepPolicy
		wantErr bool
	}{
		{name: "Nil Model", from: nil, want: nil, wantErr: true},
		{
			name: "Persist Model", from: &model.Policy{
				ID:          999,
				Name:        "Policy Test",
				Description: "Policy Description",
				Creator:     "someone",
				SrcRegistry: &model.Registry{
					ID: 123,
				},
				DestRegistry: &model.Registry{
					ID: 456,
				},
				DestNamespace: "target_ns",
				Deletion:      true,
				Override:      true,
				Enabled:       true,
				Trigger:       &model.Trigger{},
				Filters:       []*model.Filter{{Type: "registry", Value: "abc"}},
			}, want: &persist_models.RepPolicy{
				ID:                999,
				Name:              "Policy Test",
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
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := convertToPersistModel(tt.from)

			if tt.wantErr {
				assert.Equal(t, err, errNilPolicyModel)
				return
			}

			require.Nil(t, err, tt.name)
			assert.Equal(t, tt.want.ID, got.ID)
			assert.Equal(t, tt.want.Name, got.Name)
			assert.Equal(t, tt.want.Description, got.Description)
			assert.Equal(t, tt.want.Creator, got.Creator)
			assert.Equal(t, tt.want.SrcRegistryID, got.SrcRegistryID)
			assert.Equal(t, tt.want.DestRegistryID, got.DestRegistryID)
			assert.Equal(t, tt.want.DestNamespace, got.DestNamespace)
			assert.Equal(t, tt.want.ReplicateDeletion, got.ReplicateDeletion)
			assert.Equal(t, tt.want.Override, got.Override)
			assert.Equal(t, tt.want.Enabled, got.Enabled)
			assert.Equal(t, tt.want.Trigger, got.Trigger)
			assert.Equal(t, tt.want.Filters, got.Filters)

		})
	}
}

func TestNewDefaultManager(t *testing.T) {
	tests := []struct {
		name string
		want *DefaultManager
	}{
		{want: &DefaultManager{}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := NewDefaultManager(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("NewDefaultManager() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestParseFilters(t *testing.T) {
	// nil filter string
	str := ""
	filters, err := parseFilters(str)
	require.Nil(t, err)
	assert.Nil(t, filters)
	// only contains the fields that introduced in the latest version
	str = `[{"type":"name","value":"library/hello-world"}]`
	filters, err = parseFilters(str)
	require.Nil(t, err)
	require.Equal(t, 1, len(filters))
	assert.Equal(t, model.FilterTypeName, filters[0].Type)
	assert.Equal(t, "library/hello-world", filters[0].Value.(string))
	// contains "kind" from previous versions
	str = `[{"kind":"repository","value":"library/hello-world"}]`
	filters, err = parseFilters(str)
	require.Nil(t, err)
	require.Equal(t, 1, len(filters))
	assert.Equal(t, model.FilterTypeName, filters[0].Type)
	assert.Equal(t, "library/hello-world", filters[0].Value.(string))
	// contains "pattern" from previous versions
	str = `[{"kind":"repository","pattern":"library/hello-world"}]`
	filters, err = parseFilters(str)
	require.Nil(t, err)
	require.Equal(t, 1, len(filters))
	assert.Equal(t, model.FilterTypeName, filters[0].Type)
	assert.Equal(t, "library/hello-world", filters[0].Value.(string))
}

func TestParseTrigger(t *testing.T) {
	// nil trigger string
	str := ""
	trigger, err := parseTrigger(str)
	require.Nil(t, err)
	assert.Nil(t, trigger)
	// only contains the fields that introduced in the latest version
	str = `{"type":"scheduled", "trigger_settings":{"cron":"1 * * * * *"}}`
	trigger, err = parseTrigger(str)
	require.Nil(t, err)
	assert.Equal(t, model.TriggerTypeScheduled, trigger.Type)
	assert.Equal(t, "1 * * * * *", trigger.Settings.Cron)
	// contains "kind" from previous versions
	str = `{"kind":"Manual"}`
	trigger, err = parseTrigger(str)
	require.Nil(t, err)
	assert.Equal(t, model.TriggerTypeManual, trigger.Type)
	assert.Nil(t, trigger.Settings)
	// contains "kind" from previous versions
	str = `{"kind":"Immediate"}`
	trigger, err = parseTrigger(str)
	require.Nil(t, err)
	assert.Equal(t, model.TriggerTypeEventBased, trigger.Type)
	assert.Nil(t, trigger.Settings)
	// contains "schedule_param" from previous versions
	str = `{"kind":"Scheduled","schedule_param":{"type":"Weekly","weekday":1,"offtime":0}}`
	trigger, err = parseTrigger(str)
	require.Nil(t, err)
	assert.Equal(t, model.TriggerTypeScheduled, trigger.Type)
	assert.Equal(t, "0 0 0 * * 1", trigger.Settings.Cron)
	// contains "schedule_param" from previous versions
	str = `{"kind":"Scheduled","schedule_param":{"type":"Daily","offtime":0}}`
	trigger, err = parseTrigger(str)
	require.Nil(t, err)
	assert.Equal(t, model.TriggerTypeScheduled, trigger.Type)
	assert.Equal(t, "0 0 0 * * *", trigger.Settings.Cron)
}
