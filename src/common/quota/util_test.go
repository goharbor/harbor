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
	"errors"
	"testing"
)

func TestIsUnsafeError(t *testing.T) {
	type args struct {
		err error
	}
	tests := []struct {
		name string
		args args
		want bool
	}{
		{
			"is unsafe error",
			args{err: newUnsafe("unsafe")},
			true,
		},
		{
			"is not unsafe error",
			args{err: errors.New("unsafe")},
			false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := IsUnsafeError(tt.args.err); got != tt.want {
				t.Errorf("IsUnsafeError() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_checkQuotas(t *testing.T) {
	type args struct {
		hardLimits ResourceList
		used       ResourceList
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			"unlimited",
			args{hardLimits: ResourceList{ResourceStorage: UNLIMITED}, used: ResourceList{ResourceStorage: 1000}},
			false,
		},
		{
			"ok",
			args{hardLimits: ResourceList{ResourceStorage: 100}, used: ResourceList{ResourceStorage: 1}},
			false,
		},
		{
			"bad used value",
			args{hardLimits: ResourceList{ResourceStorage: 100}, used: ResourceList{ResourceStorage: -1}},
			true,
		},
		{
			"over the hard limit",
			args{hardLimits: ResourceList{ResourceStorage: 100}, used: ResourceList{ResourceStorage: 200}},
			true,
		},
		{
			"hard limit not found",
			args{hardLimits: ResourceList{ResourceStorage: 100}, used: ResourceList{ResourceCount: 1}},
			true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := isSafe(tt.args.hardLimits, tt.args.used); (err != nil) != tt.wantErr {
				t.Errorf("isSafe() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
