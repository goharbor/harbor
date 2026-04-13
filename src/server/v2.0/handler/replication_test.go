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

package handler

import (
	"testing"
	"time"

	"github.com/go-openapi/strfmt"
	"github.com/stretchr/testify/assert"

	"github.com/goharbor/harbor/src/controller/replication"
	repctlmodel "github.com/goharbor/harbor/src/controller/replication/model"
	"github.com/goharbor/harbor/src/jobservice/job"
	"github.com/goharbor/harbor/src/pkg/reg/model"
	"github.com/goharbor/harbor/src/pkg/task"
	taskdao "github.com/goharbor/harbor/src/pkg/task/dao"
	"github.com/goharbor/harbor/src/server/v2.0/models"
)

func TestConvertReplicationPolicy(t *testing.T) {
	testTime := time.Now()
	
	tests := []struct {
		name     string
		policy   *repctlmodel.Policy
		expected *models.ReplicationPolicy
		shouldPanic bool
	}{
		{
			name: "policy with string filter values",
			policy: &repctlmodel.Policy{
				ID:          1,
				Name:        "test-policy",
				Description: "test description",
				CreationTime: testTime,
				UpdateTime:   testTime,
				Enabled:     true,
				Override:    true,
				ReplicateDeletion: false,
				DestNamespace: "test-namespace",
				DestNamespaceReplaceCount: 2,
				Speed: 1024,
				CopyByChunk: true,
				SingleActiveReplication: false,
				SrcRegistry: &model.Registry{
					ID:   1,
					Name: "src-registry",
					URL:  "https://src.example.com",
				},
				DestRegistry: &model.Registry{
					ID:   2,
					Name: "dest-registry",
					URL:  "https://dest.example.com",
				},
				Filters: []*model.Filter{
					{
						Type:       model.FilterTypeTag,
						Value:      "latest",
						Decoration: "matches",
					},
					{
						Type:       model.FilterTypeLabel,
						Value:      []string{"env=prod", "team=backend"},
						Decoration: "matches",
					},
					{
						Type:  model.FilterTypeName,
						Value: "library/*",
					},
				},
				Trigger: &model.Trigger{
					Type: model.TriggerTypeManual,
					Settings: &model.TriggerSettings{
						Cron: "0 0 * * *",
					},
				},
			},
			expected: &models.ReplicationPolicy{
				ID:                        1,
				Name:                      "test-policy",
				Description:               "test description",
				CreationTime:              strfmt.DateTime(testTime),
				UpdateTime:                strfmt.DateTime(testTime),
				Enabled:                   true,
				Override:                  true,
				Deletion:                  false,
				ReplicateDeletion:         false,
				DestNamespace:             "test-namespace",
				DestNamespaceReplaceCount: func() *int8 { v := int8(2); return &v }(),
				Speed:                     func() *int32 { v := int32(1024); return &v }(),
				CopyByChunk:               func() *bool { v := true; return &v }(),
				SingleActiveReplication:   func() *bool { v := false; return &v }(),
				SrcRegistry: &models.Registry{
					ID:   1,
					Name: "src-registry",
					URL:  "https://src.example.com",
				},
				DestRegistry: &models.Registry{
					ID:   2,
					Name: "dest-registry",
					URL:  "https://dest.example.com",
				},
				Filters: []*models.ReplicationFilter{
					{
						Type:       "tag",
						Value:      "latest",
						Decoration: "matches",
					},
					{
						Type:       "label",
						Value:      []string{"env=prod", "team=backend"},
						Decoration: "matches",
					},
					{
						Type:  "name",
						Value: "library/*",
					},
				},
				Trigger: &models.ReplicationTrigger{
					Type: "manual",
					TriggerSettings: &models.ReplicationTriggerSettings{
						Cron: "0 0 * * *",
					},
				},
			},
		},
		{
			name: "policy with empty filters",
			policy: &repctlmodel.Policy{
				ID:          2,
				Name:        "empty-filter-policy",
				Description: "policy with no filters",
				CreationTime: testTime,
				UpdateTime:   testTime,
				Enabled:     true,
				DestNamespaceReplaceCount: -1,
				Speed: 0,
				CopyByChunk: false,
				SingleActiveReplication: true,
				Filters: []*model.Filter{},
			},
			expected: &models.ReplicationPolicy{
				ID:                        2,
				Name:                      "empty-filter-policy",
				Description:               "policy with no filters",
				CreationTime:              strfmt.DateTime(testTime),
				UpdateTime:                strfmt.DateTime(testTime),
				Enabled:                   true,
				DestNamespaceReplaceCount: func() *int8 { v := int8(-1); return &v }(),
				Speed:                     func() *int32 { v := int32(0); return &v }(),
				CopyByChunk:               func() *bool { v := false; return &v }(),
				SingleActiveReplication:   func() *bool { v := true; return &v }(),
				Filters:                   []*models.ReplicationFilter{},
			},
		},
		{
			name: "policy with nil registries and trigger",
			policy: &repctlmodel.Policy{
				ID:          3,
				Name:        "minimal-policy",
				Description: "minimal policy configuration",
				CreationTime: testTime,
				UpdateTime:   testTime,
				Enabled:     false,
				DestNamespaceReplaceCount: -1,
				Speed: 0,
				CopyByChunk: false,
				SingleActiveReplication: false,
			},
			expected: &models.ReplicationPolicy{
				ID:                        3,
				Name:                      "minimal-policy",
				Description:               "minimal policy configuration",
				CreationTime:              strfmt.DateTime(testTime),
				UpdateTime:                strfmt.DateTime(testTime),
				Enabled:                   false,
				DestNamespaceReplaceCount: func() *int8 { v := int8(-1); return &v }(),
				Speed:                     func() *int32 { v := int32(0); return &v }(),
				CopyByChunk:               func() *bool { v := false; return &v }(),
				SingleActiveReplication:   func() *bool { v := false; return &v }(),
			},
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.shouldPanic {
				assert.Panics(t, func() {
					convertReplicationPolicy(tt.policy)
				})
			} else {
				result := convertReplicationPolicy(tt.policy)
				assert.Equal(t, tt.expected.ID, result.ID)
				assert.Equal(t, tt.expected.Name, result.Name)
				assert.Equal(t, tt.expected.Description, result.Description)
				assert.Equal(t, tt.expected.Enabled, result.Enabled)
				assert.Equal(t, tt.expected.Override, result.Override)
				assert.Equal(t, tt.expected.Deletion, result.Deletion)
				assert.Equal(t, tt.expected.ReplicateDeletion, result.ReplicateDeletion)
				assert.Equal(t, tt.expected.DestNamespace, result.DestNamespace)
				
				if tt.expected.DestNamespaceReplaceCount != nil {
					assert.Equal(t, *tt.expected.DestNamespaceReplaceCount, *result.DestNamespaceReplaceCount)
				}
				if tt.expected.Speed != nil {
					assert.Equal(t, *tt.expected.Speed, *result.Speed)
				}
				if tt.expected.CopyByChunk != nil {
					assert.Equal(t, *tt.expected.CopyByChunk, *result.CopyByChunk)
				}
				if tt.expected.SingleActiveReplication != nil {
					assert.Equal(t, *tt.expected.SingleActiveReplication, *result.SingleActiveReplication)
				}
				
				// Check filters
				if len(tt.expected.Filters) > 0 {
					assert.Len(t, result.Filters, len(tt.expected.Filters))
					for i, expectedFilter := range tt.expected.Filters {
						assert.Equal(t, expectedFilter.Type, result.Filters[i].Type)
						assert.Equal(t, expectedFilter.Value, result.Filters[i].Value)
						assert.Equal(t, expectedFilter.Decoration, result.Filters[i].Decoration)
					}
				}
				
				// Check registries
				if tt.expected.SrcRegistry != nil {
					assert.Equal(t, tt.expected.SrcRegistry.ID, result.SrcRegistry.ID)
					assert.Equal(t, tt.expected.SrcRegistry.Name, result.SrcRegistry.Name)
					assert.Equal(t, tt.expected.SrcRegistry.URL, result.SrcRegistry.URL)
				}
				if tt.expected.DestRegistry != nil {
					assert.Equal(t, tt.expected.DestRegistry.ID, result.DestRegistry.ID)
					assert.Equal(t, tt.expected.DestRegistry.Name, result.DestRegistry.Name)
					assert.Equal(t, tt.expected.DestRegistry.URL, result.DestRegistry.URL)
				}
				
				// Check trigger
				if tt.expected.Trigger != nil {
					assert.Equal(t, tt.expected.Trigger.Type, result.Trigger.Type)
					if tt.expected.Trigger.TriggerSettings != nil {
						assert.Equal(t, tt.expected.Trigger.TriggerSettings.Cron, result.Trigger.TriggerSettings.Cron)
					}
				}
			}
		})
	}
}

func TestConvertReplicationPolicyFilterTypeAssertion(t *testing.T) {
	testTime := time.Now()
	
	// Test case where filter.Value is not a string but should be handled gracefully
	policy := &repctlmodel.Policy{
		ID:          1,
		Name:        "test-policy",
		CreationTime: testTime,
		UpdateTime:   testTime,
		Enabled:     true,
		DestNamespaceReplaceCount: -1,
		Speed: 0,
		CopyByChunk: false,
		SingleActiveReplication: false,
		Filters: []*model.Filter{
			{
				Type:  model.FilterTypeTag,
				Value: 123,
			},
		},
	}
	
	// The function should handle non-string values by preserving their types
	result := convertReplicationPolicy(policy)
	assert.NotNil(t, result)
	assert.Len(t, result.Filters, 1)
	assert.Equal(t, 123, result.Filters[0].Value, "Expected non-string value to be preserved as original type")
}

// Test helper functions
func TestConvertExecution(t *testing.T) {
	execution := &replication.Execution{
		ID:       1,
		PolicyID: 2,
		Status:   job.RunningStatus.String(),
		Trigger:  task.ExecutionTriggerManual,
		StartTime: time.Now(),
		EndTime:   time.Now(),
		StatusMessage: "test execution",
		Metrics: &taskdao.Metrics{
			TaskCount:        10,
			SuccessTaskCount: 8,
			ErrorTaskCount:   1,
			PendingTaskCount: 1,
		},
	}
	
	result := convertExecution(execution)
	
	assert.Equal(t, int64(1), result.ID)
	assert.Equal(t, int64(2), result.PolicyID)
	assert.Equal(t, "InProgress", result.Status)
	assert.Equal(t, "manual", result.Trigger)
	assert.Equal(t, int64(10), result.Total)
	assert.Equal(t, int64(8), result.Succeed)
	assert.Equal(t, int64(1), result.Failed)
	assert.Equal(t, int64(1), result.InProgress)
}

func TestConvertTask(t *testing.T) {
	task := &replication.Task{
		ID:                1,
		ExecutionID:       2,
		JobID:             "job-123",
		Status:            job.SuccessStatus.String(),
		Operation:         "copy",
		ResourceType:      "image",
		SourceResource:    "source/repo:tag",
		DestinationResource: "dest/repo:tag",
		StartTime:         time.Now(),
		EndTime:           time.Now(),
	}
	
	result := convertTask(task)
	
	assert.Equal(t, int64(1), result.ID)
	assert.Equal(t, int64(2), result.ExecutionID)
	assert.Equal(t, "job-123", result.JobID)
	assert.Equal(t, "Succeed", result.Status)
	assert.Equal(t, "copy", result.Operation)
	assert.Equal(t, "image", result.ResourceType)
	assert.Equal(t, "source/repo:tag", result.SrcResource)
	assert.Equal(t, "dest/repo:tag", result.DstResource)
}