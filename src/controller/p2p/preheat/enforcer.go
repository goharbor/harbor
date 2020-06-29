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

package preheat

import (
	"github.com/goharbor/harbor/src/pkg/artifact"
	"github.com/goharbor/harbor/src/pkg/p2p/preheat/models/policy"
)

// Enforcer defines policy enforcement operations.
type Enforcer interface {
	// Enforce the specified policy.
	//
	// Arguments:
	//   p *policy.Schema : the being enforced policy
	//   art ...*artifact.Artifact （optional）: the relevant artifact referred by the happening events
	//   that defined in the event-based policy p.
	//
	// Returns:
	//   - ID of the execution
	//   - non-nil error if any error occurred during the enforcement
	Enforce(p *policy.Schema, art ...*artifact.Artifact) (int64, error)
}
