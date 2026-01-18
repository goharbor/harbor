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

package scan

import "context"

// ScopeHeader is the HTTP header key used to carry scan-all scope JSON.
const ScopeHeader = "X-Scan-All-Scope"

// ScanAllScope defines optional filters for scan-all.
// Currently supports filtering by project IDs or repository names.
// Leave all fields empty to scan everything (default behavior).
type ScanAllScope struct {
	ProjectIDs   []int64  `json:"project_ids,omitempty"`
	Repositories []string `json:"repositories,omitempty"`
}

// scopeCtxKey is the context key type for storing scope in context
// to avoid collisions.
type scopeCtxKey struct{}

// WithScanAllScope returns a new context with the given scope.
func WithScanAllScope(ctx context.Context, scope *ScanAllScope) context.Context {
	if scope == nil {
		return ctx
	}
	return context.WithValue(ctx, scopeCtxKey{}, scope)
}

// FromContextScope returns the ScanAllScope from context if present.
func FromContextScope(ctx context.Context) *ScanAllScope {
	v := ctx.Value(scopeCtxKey{})
	if v == nil {
		return nil
	}
	if s, ok := v.(*ScanAllScope); ok {
		return s
	}
	return nil
}
