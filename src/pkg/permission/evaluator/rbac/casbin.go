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

package rbac

import (
	"github.com/casbin/casbin"
	"github.com/casbin/casbin/model"
	"github.com/goharbor/harbor/src/pkg/permission/types"
)

// Syntax for models see https://casbin.org/docs/en/syntax-for-models
const modelText = `
# Request definition
[request_definition]
r = sub, obj, act

# Policy definition
[policy_definition]
p = sub, obj, act, eft

# Role definition
[role_definition]
g = _, _

# Policy effect
[policy_effect]
e = some(where (p.eft == allow)) && !some(where (p.eft == deny))

# Matchers
[matchers]
m = g(r.sub, p.sub) && keyMatch2(r.obj, p.obj) && (r.act == p.act || p.act == '*')
`

func makeEnforcer(rbacUser types.RBACUser) *casbin.Enforcer {
	m := model.Model{}
	m.LoadModelFromText(modelText)

	e := casbin.NewEnforcer(m, &adapter{rbacUser: rbacUser})
	e.AddFunction("keyMatch2", keyMatch2Func)
	return e
}
