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

package namespace

import (
	"context"
	"fmt"
	"sync"

	"github.com/goharbor/harbor/src/pkg/permission/evaluator"
	"github.com/goharbor/harbor/src/pkg/permission/types"
)

// EvaluatorFactory returns the evaluator.Evaluator of the namespace
type EvaluatorFactory func(context.Context, types.Namespace) evaluator.Evaluator

var _ evaluator.Evaluator = &Evaluator{}

// Evaluator evaluator for the namespace
type Evaluator struct {
	factory       EvaluatorFactory
	namespaceKind string
	cache         sync.Map
}

// HasPermission returns true when user has action permission for the resource
func (e *Evaluator) HasPermission(ctx context.Context, resource types.Resource, action types.Action) bool {
	ns, ok := types.NamespaceFromResource(resource)
	if ok && ns.Kind() == e.namespaceKind {
		var eva evaluator.Evaluator

		key := fmt.Sprintf("%p:%s:%v", ctx, ns.Kind(), ns.Identity())
		value, ok := e.cache.Load(key)
		if !ok {
			eva = e.factory(ctx, ns)
			e.cache.Store(key, eva)
		} else {
			eva, _ = value.(evaluator.Evaluator) // maybe value is nil
		}

		return eva != nil && eva.HasPermission(ctx, resource, action)
	}

	return false
}

// New returns permission evaluator for which support namespace
func New(namespaceKind string, factory EvaluatorFactory) *Evaluator {
	return &Evaluator{
		namespaceKind: namespaceKind,
		factory:       factory,
	}
}
