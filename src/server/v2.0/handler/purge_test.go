//  Copyright Project Harbor Authors
//
//  Licensed under the Apache License, Version 2.0 (the "License");
//  you may not use this file except in compliance with the License.
//  You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
//  Unless required by applicable law or agreed to in writing, software
//  distributed under the License is distributed on an "AS IS" BASIS,
//  WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
//  See the License for the specific language governing permissions and
//  limitations under the License.

package handler

import (
	"fmt"
	"testing"

	"github.com/goharbor/harbor/src/common"
	"github.com/goharbor/harbor/src/server/v2.0/models"
	"github.com/goharbor/harbor/src/server/v2.0/restapi/operations/purge"
	"github.com/stretchr/testify/assert"
)

func Test_verifyUpdateRequest(t *testing.T) {
	type args struct {
		params purge.UpdatePurgeScheduleParams
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{"normal", args{purge.UpdatePurgeScheduleParams{Schedule: &models.Schedule{Schedule: &models.ScheduleObj{}, Parameters: map[string]interface{}{common.PurgeAuditRetentionHour: "168", common.PurgeAuditIncludeOperations: "pull"}}}}, false},
		{"missing_schedule", args{purge.UpdatePurgeScheduleParams{Schedule: &models.Schedule{Parameters: map[string]interface{}{common.PurgeAuditRetentionHour: "168", common.PurgeAuditIncludeOperations: "pull"}}}}, true},
		{"missing_retention_hour", args{purge.UpdatePurgeScheduleParams{Schedule: &models.Schedule{Schedule: &models.ScheduleObj{}, Parameters: map[string]interface{}{common.PurgeAuditIncludeOperations: "pull"}}}}, true},
		{"missing_operations", args{purge.UpdatePurgeScheduleParams{Schedule: &models.Schedule{Schedule: &models.ScheduleObj{}, Parameters: map[string]interface{}{common.PurgeAuditRetentionHour: "168"}}}}, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := verifyUpdateRequest(tt.args.params)
			if tt.wantErr != (err != nil) {
				t.Error("test failed")
			}
		})
	}
}
func Test_verifyCreateRequest(t *testing.T) {
	type args struct {
		params purge.CreatePurgeScheduleParams
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{"normal", args{purge.CreatePurgeScheduleParams{Schedule: &models.Schedule{Schedule: &models.ScheduleObj{}, Parameters: map[string]interface{}{common.PurgeAuditRetentionHour: "168", common.PurgeAuditIncludeOperations: "pull"}}}}, false},
		{"missing_schedule", args{purge.CreatePurgeScheduleParams{Schedule: &models.Schedule{Parameters: map[string]interface{}{common.PurgeAuditRetentionHour: "168", common.PurgeAuditIncludeOperations: "pull"}}}}, true},
		{"missing_retention_hour", args{purge.CreatePurgeScheduleParams{Schedule: &models.Schedule{Schedule: &models.ScheduleObj{}, Parameters: map[string]interface{}{common.PurgeAuditIncludeOperations: "pull"}}}}, true},
		{"missing_operations", args{purge.CreatePurgeScheduleParams{Schedule: &models.Schedule{Schedule: &models.ScheduleObj{}, Parameters: map[string]interface{}{common.PurgeAuditRetentionHour: "168"}}}}, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := verifyCreateRequest(tt.args.params)
			if tt.wantErr != (err != nil) {
				t.Error("test failed")
			}
		})
	}
}

func Test_checkRetentionHour(t *testing.T) {
	type args struct {
		m map[string]interface{}
	}
	tests := []struct {
		name    string
		args    args
		want    int
		wantErr assert.ErrorAssertionFunc
	}{
		{"normal", args{map[string]interface{}{common.PurgeAuditRetentionHour: 24}}, 24, func(t assert.TestingT, err error, i ...interface{}) bool { return false }},
		{"overflow", args{map[string]interface{}{common.PurgeAuditRetentionHour: 250000}}, 0, func(t assert.TestingT, err error, i ...interface{}) bool { return true }},
		{"equal", args{map[string]interface{}{common.PurgeAuditRetentionHour: 240000}}, 240000, func(t assert.TestingT, err error, i ...interface{}) bool { return false }},
		{"wrong type", args{map[string]interface{}{common.PurgeAuditRetentionHour: "wrong type"}}, 0, func(t assert.TestingT, err error, i ...interface{}) bool { return true }},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := retentionHour(tt.args.m)
			if !tt.wantErr(t, err, fmt.Sprintf("retentionHour(%v)", tt.args.m)) {
				return
			}
			assert.Equalf(t, tt.want, got, "retentionHour(%v)", tt.args.m)
		})
	}
}
