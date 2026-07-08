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

package evaluator

import (
	"context"

	"github.com/goharbor/harbor/src/pkg/permission/types"
)

// Evaluator the permission evaluator
type Evaluator interface {
	// HasPermission returns true when user has action permission for the resource
	HasPermission(ctx context.Context, resource types.Resource, action types.Action) bool
}

// Evaluators evaluator set
type Evaluators []Evaluator

// Add adds an evaluator to a given slice of evaluators
func (evaluators Evaluators) Add(newEvaluators ...Evaluator) Evaluators {
	for _, newEvaluator := range newEvaluators {
		if newEvaluator == nil {
			continue
		}

		if items, ok := newEvaluator.(Evaluators); ok {
			evaluators = evaluators.Add(items...)
		} else {
			exists := false
			for _, evaluator := range evaluators {
				if evaluator == newEvaluator {
					exists = true
				}
			}
			if !exists {
				evaluators = append(evaluators, newEvaluator)
			}
		}
	}

	return evaluators
}

// HasPermission returns true when one of evaluator has action permission for the resource
func (evaluators Evaluators) HasPermission(ctx context.Context, resource types.Resource, action types.Action) bool {
	for _, evaluator := range evaluators {
		if evaluator != nil && evaluator.HasPermission(ctx, resource, action) {
			return true
		}
	}

	return false
}
