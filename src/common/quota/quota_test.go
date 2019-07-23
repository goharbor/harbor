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

	_ "github.com/goharbor/harbor/src/common/quota/driver/project"
	"github.com/goharbor/harbor/src/pkg/types"
)

func TestValidate(t *testing.T) {
	type args struct {
		reference  string
		hardLimits types.ResourceList
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{"valid", args{"project", types.ResourceList{types.ResourceCount: 1, types.ResourceStorage: 1}}, false},
		{"invalid", args{"project", types.ResourceList{types.ResourceCount: 1, types.ResourceStorage: 0}}, true},
		{"not support", args{"not support", types.ResourceList{types.ResourceCount: 1}}, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := Validate(tt.args.reference, tt.args.hardLimits); (err != nil) != tt.wantErr {
				t.Errorf("Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
