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
	"github.com/goharbor/harbor/src/server/v2.0/models"
	"github.com/stretchr/testify/assert"
	"reflect"
	"testing"
)

func Test_sortMostMatch(t *testing.T) {
	type args struct {
		input     []*models.UserGroupSearchItem
		matchWord string
		expected  []*models.UserGroupSearchItem
	}
	tests := []struct {
		name string
		args args
	}{
		{"normal", args{[]*models.UserGroupSearchItem{
			{GroupName: "user"}, {GroupName: "harbor_user"}, {GroupName: "admin_user"}, {GroupName: "users"},
		}, "user", []*models.UserGroupSearchItem{
			{GroupName: "user"}, {GroupName: "users"}, {GroupName: "admin_user"}, {GroupName: "harbor_user"},
		}}},
		{"duplicate_item", args{[]*models.UserGroupSearchItem{
			{GroupName: "user"}, {GroupName: "user"}, {GroupName: "harbor_user"}, {GroupName: "admin_user"}, {GroupName: "users"},
		}, "user", []*models.UserGroupSearchItem{
			{GroupName: "user"}, {GroupName: "user"}, {GroupName: "users"}, {GroupName: "admin_user"}, {GroupName: "harbor_user"},
		}}},
		{"miss_exact_match", args{[]*models.UserGroupSearchItem{
			{GroupName: "harbor_user"}, {GroupName: "admin_user"}, {GroupName: "users"},
		}, "user", []*models.UserGroupSearchItem{
			{GroupName: "users"}, {GroupName: "admin_user"}, {GroupName: "harbor_user"},
		}}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sortMostMatch(tt.args.input, tt.args.matchWord)
			assert.True(t, reflect.DeepEqual(tt.args.input, tt.args.expected))
		})
	}
}
