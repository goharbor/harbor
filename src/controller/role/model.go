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

package role

import (
	"github.com/goharbor/harbor/src/pkg/permission/types"
	"github.com/goharbor/harbor/src/pkg/role/model"
)

const (
	LEVELROLE = "project-role"

	// SCOPEALLPROJECT ...
	SCOPEALLPROJECT = "/project/*"

	// ROLETYPE ...
	ROLETYPE = "project-role"
)

// Role ...
type Role struct {
	model.Role
	Level       string
	Editable    bool          `json:"editable"`
	Permissions []*Permission `json:"permissions"`
}

// IsSysLevel, true is a system level robot, others are project level.

// setLevel = project-role
func (r *Role) setLevel() {
	r.Level = LEVELROLE
}

// setEditable, no secret and no permissions should be a old format robot, and it's not editable.
func (r *Role) setEditable() {
	r.Editable = true
}

// Permission ...
type Permission struct {
	Kind      string          `json:"kind"`
	Namespace string          `json:"namespace"`
	Access    []*types.Policy `json:"access"`
	Scope     string          `json:"-"`
}

// Option ...
type Option struct {
	WithPermission bool
	Operator       string
}
