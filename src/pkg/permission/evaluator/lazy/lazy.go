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

package lazy

import (
	"context"
	"sync"

	"github.com/goharbor/harbor/src/pkg/permission/evaluator"
	"github.com/goharbor/harbor/src/pkg/permission/types"
)

// EvaluatorFactory the permission evaluator factory
type EvaluatorFactory func() evaluator.Evaluator

var _ evaluator.Evaluator = &Evaluator{}

// Evaluator lazy permission evaluator
type Evaluator struct {
	factory   EvaluatorFactory
	evaluator evaluator.Evaluator
	once      sync.Once
}

// HasPermission returns true when user has action permission for the resource
func (l *Evaluator) HasPermission(ctx context.Context, resource types.Resource, action types.Action) bool {
	l.once.Do(func() {
		l.evaluator = l.factory()
	})

	return l.evaluator != nil && l.evaluator.HasPermission(ctx, resource, action)
}

// New returns lazy evaluator
func New(factory EvaluatorFactory) *Evaluator {
	return &Evaluator{factory: factory}
}
