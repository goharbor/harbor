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

package system

import (
	"context"

	"github.com/goharbor/harbor/src/pkg/permission/evaluator"
	"github.com/goharbor/harbor/src/pkg/permission/evaluator/namespace"
	"github.com/goharbor/harbor/src/pkg/permission/evaluator/rbac"
	"github.com/goharbor/harbor/src/pkg/permission/types"
)

// NewEvaluator create evaluator for the system
func NewEvaluator(username string, policies []*types.Policy) evaluator.Evaluator {
	return namespace.New(NamespaceKind, func(_ context.Context, _ types.Namespace) evaluator.Evaluator {
		return rbac.New(&rbacUser{
			username: username,
			policies: policies,
		})
	})
}
