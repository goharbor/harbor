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

package quota

import (
	"testing"

	"github.com/goharbor/harbor/src/pkg/types"
)

func Test_isSafe(t *testing.T) {
	type args struct {
		hardLimits   types.ResourceList
		currentUsed  types.ResourceList
		newUsed      types.ResourceList
		skipOverflow bool
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			"unlimited",
			args{
				types.ResourceList{types.ResourceStorage: types.UNLIMITED},
				types.ResourceList{types.ResourceStorage: 1000},
				types.ResourceList{types.ResourceStorage: 1000},
				false,
			},
			false,
		},
		{
			"ok",
			args{
				types.ResourceList{types.ResourceStorage: 100},
				types.ResourceList{types.ResourceStorage: 10},
				types.ResourceList{types.ResourceStorage: 1},
				false,
			},
			false,
		},
		{
			"over the hard limit",
			args{
				types.ResourceList{types.ResourceStorage: 100},
				types.ResourceList{types.ResourceStorage: 0},
				types.ResourceList{types.ResourceStorage: 200},
				false,
			},
			true,
		},
		{
			"skip overflow",
			args{
				types.ResourceList{types.ResourceStorage: 100},
				types.ResourceList{types.ResourceStorage: 0},
				types.ResourceList{types.ResourceStorage: 200},
				true,
			},
			false,
		},
		{
			"hard limit not found",
			args{
				types.ResourceList{types.ResourceStorage: 100},
				types.ResourceList{types.ResourceCount: 0},
				types.ResourceList{types.ResourceCount: 1},
				false,
			},
			true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := isSafe(tt.args.hardLimits, tt.args.currentUsed, tt.args.newUsed, tt.args.skipOverflow); (err != nil) != tt.wantErr {
				t.Errorf("isSafe() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
