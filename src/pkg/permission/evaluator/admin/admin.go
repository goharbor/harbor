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

package admin

import (
	"github.com/goharbor/harbor/src/common/utils/log"
	"github.com/goharbor/harbor/src/pkg/permission/evaluator"
	"github.com/goharbor/harbor/src/pkg/permission/types"
)

var _ evaluator.Evaluator = &Evaluator{}

// Evaluator the permission evaluator for the system administrator
type Evaluator struct {
	username string
}

// HasPermission always return true for the system administrator
func (e *Evaluator) HasPermission(resource types.Resource, action types.Action) bool {
	log.Debugf("system administrator %s require %s action for resource %s", e.username, action, resource)
	return true
}

// New returns evaluator.Evaluator for the system administrator
func New(username string) *Evaluator {
	return &Evaluator{username: username}
}
