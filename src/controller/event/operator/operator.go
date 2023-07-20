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

package operator

import (
	"context"

	"github.com/goharbor/harbor/src/common/security"
)

// ContextKey is the key for storing operator in the context.
type ContextKey struct{}

// FromContext return the event operator from context
func FromContext(ctx context.Context) string {
	var operator string
	sc, ok := security.FromContext(ctx)
	if ok {
		operator = sc.GetUsername()
	}
	// retrieve from context if not found in security context
	if operator == "" {
		op, ok := ctx.Value(ContextKey{}).(string)
		if ok {
			operator = op
		}
	}

	return operator
}
