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

package purge

import "testing"

func TestString(t *testing.T) {
	type args struct {
		extras map[string]interface{}
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{"normal", args{map[string]interface{}{"dry_run": true, "audit_log_retention_hour": 168}}, "{\"audit_log_retention_hour\":168,\"dry_run\":true}"},
		{"empty", args{map[string]interface{}{}}, "{}"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := String(tt.args.extras); got != tt.want {
				t.Errorf("String() = %v, want %v", got, tt.want)
			}
		})
	}
}
