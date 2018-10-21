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

package models

import (
	"github.com/astaxie/beego/validation"
)

// Replication defines the properties of model used in replication API
type Replication struct {
	PolicyID int64 `json:"policy_id"`
}

// ReplicationResponse describes response of a replication request, it gives
type ReplicationResponse struct {
	UUID string `json:"uuid"`
}

// Valid ...
func (r *Replication) Valid(v *validation.Validation) {
	if r.PolicyID <= 0 {
		v.SetError("policy_id", "invalid value")
	}
}
