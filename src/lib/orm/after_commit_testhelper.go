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

package orm

import "context"

// ContextWithAfterCommitHooksForTest returns a context that carries an
// AfterCommit hooks sink without a real database transaction, plus a
// drain function that runs all queued callbacks (simulating a commit).
// It is intended only for unit tests that want to exercise code paths
// registering AfterCommit callbacks in-transaction, without needing a
// live ORM / Postgres harness.
func ContextWithAfterCommitHooksForTest(ctx context.Context) (context.Context, func()) {
	h := &txHooks{}
	ctx = context.WithValue(ctx, hooksKey{}, h)
	return ctx, func() {
		for _, fn := range h.drain() {
			safeInvoke(fn)
		}
	}
}
