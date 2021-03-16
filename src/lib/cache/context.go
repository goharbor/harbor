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

package cache

import (
	"context"
)

type cacheKey struct{}

// FromContext returns cache from context
func FromContext(ctx context.Context) (Cache, bool) {
	c, ok := ctx.Value(cacheKey{}).(Cache)
	return c, ok
}

// NewContext returns new context with cache
func NewContext(ctx context.Context, c Cache) context.Context {
	if ctx == nil {
		ctx = context.Background()
	}
	return context.WithValue(ctx, cacheKey{}, c)
}
