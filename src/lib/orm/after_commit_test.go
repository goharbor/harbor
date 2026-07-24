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

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
)

// TestAfterCommit_NoTransaction covers the non-tx path: without an
// enclosing WithTransaction scope, AfterCommit must run the callback
// immediately on the caller's goroutine so no cleanup is ever lost.
func TestAfterCommit_NoTransaction(t *testing.T) {
	ran := false
	AfterCommit(context.Background(), func() { ran = true })
	assert.True(t, ran, "AfterCommit must run fn immediately when no tx hooks sink is on the ctx")
}

// TestAfterCommit_NilFn is a no-op and must not panic.
func TestAfterCommit_NilFn(t *testing.T) {
	assert.NotPanics(t, func() {
		AfterCommit(context.Background(), nil)
	})
}

// TestAfterCommit_RecoversPanic verifies hook panics are contained so
// one broken hook cannot take out an entire commit path.
func TestAfterCommit_RecoversPanic(t *testing.T) {
	assert.NotPanics(t, func() {
		AfterCommit(context.Background(), func() { panic("boom") })
	})
}

// TestAfterCommit_QueuesWhenHooksPresent asserts that when a hooks sink
// is attached to the context, AfterCommit queues the callback rather
// than running it inline. This is the in-tx path; WithTransaction tests
// live in lib/orm/test for the real-DB commit/rollback semantics.
func TestAfterCommit_QueuesWhenHooksPresent(t *testing.T) {
	h := &txHooks{}
	ctx := context.WithValue(context.Background(), hooksKey{}, h)

	ran := false
	AfterCommit(ctx, func() { ran = true })

	assert.False(t, ran, "hooks must not fire before commit")

	cbs := h.drain()
	assert.Len(t, cbs, 1)

	cbs[0]()
	assert.True(t, ran)
}
