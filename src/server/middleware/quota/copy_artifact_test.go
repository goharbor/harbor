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

import "testing"

func Test_parseRepositoryName(t *testing.T) {
	type args struct {
		p string
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{"/api/v2.0/projects", args{"/api/v2.0/projects"}, ""},
		{"/api/v2.0/projects/library/repositories/photon/artifacts", args{"/api/v2.0/projects/library/repositories/photon/artifacts"}, "photon"},
		{"/api/v2.0/projects/library/repositories/photon/artifacts/", args{"/api/v2.0/projects/library/repositories/photon/artifacts/"}, "photon"},
		{"/api/v2.0/projects/library/repositories/amd64/photon/artifacts", args{"/api/v2.0/projects/library/repositories/amd64/photon/artifacts"}, "amd64/photon"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := parseRepositoryName(tt.args.p); got != tt.want {
				t.Errorf("parseRepositoryName() = %v, want %v", got, tt.want)
			}
		})
	}
}
