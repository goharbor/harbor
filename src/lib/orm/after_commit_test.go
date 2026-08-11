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

	cbs := h.close()
	assert.Len(t, cbs, 1)

	cbs[0]()
	assert.True(t, ran)
}

// TestTxHooks_AdoptKeepsRegistrationOrder covers the savepoint-scoping
// primitive used by WithTransaction: a released nested scope hands its
// callbacks to the enclosing sink, which must keep them in registration
// order relative to the enclosing scope's own.
func TestTxHooks_AdoptKeepsRegistrationOrder(t *testing.T) {
	outer, inner := &txHooks{}, &txHooks{}

	var fired []string
	queue := func(h *txHooks, name string) { h.add(func() { fired = append(fired, name) }) }

	queue(outer, "outer-before")
	queue(inner, "inner-1")
	queue(inner, "inner-2")
	outer.adopt(inner.close()) // savepoint released
	queue(outer, "outer-after")

	for _, fn := range outer.close() {
		fn()
	}
	assert.Equal(t, []string{"outer-before", "inner-1", "inner-2", "outer-after"}, fired)
	assert.Empty(t, inner.afterCommit, "a closed sink must not keep its callbacks")
}

// TestTxHooks_DroppedScopeLeavesParentIntact asserts that dropping a
// rolled-back scope's sink cannot touch the enclosing scope's callbacks,
// regardless of when they were registered.
func TestTxHooks_DroppedScopeLeavesParentIntact(t *testing.T) {
	outer, inner := &txHooks{}, &txHooks{}

	outer.add(func() {})
	inner.add(func() {})
	outer.add(func() {}) // registered while the nested scope was still open

	inner.close() // savepoint rolled back: sink discarded, never adopted

	assert.Len(t, outer.afterCommit, 2)
}

// TestTxHooks_AdoptEmpty is a no-op: a scope that registered nothing must
// not disturb the enclosing sink.
func TestTxHooks_AdoptEmpty(t *testing.T) {
	outer := &txHooks{}
	outer.add(func() {})

	outer.adopt(nil)
	outer.adopt((&txHooks{}).close())

	assert.Len(t, outer.afterCommit, 1)
}

// TestTxHooks_ClosedScopeForwardsToParent asserts that registering through a
// context whose scope has already ended is not silently lost: it lands on the
// nearest enclosing scope that is still open, however deep the chain.
func TestTxHooks_ClosedScopeForwardsToParent(t *testing.T) {
	outer := &txHooks{}
	middle := &txHooks{parent: outer}
	inner := &txHooks{parent: middle}

	middle.adopt(inner.close()) // savepoints released, both scopes ended
	outer.adopt(middle.close())

	ran := false
	inner.add(func() { ran = true })

	assert.False(t, ran, "forwarded callback must wait for the still-open scope")
	assert.Len(t, outer.afterCommit, 1)

	for _, fn := range outer.close() {
		fn()
	}
	assert.True(t, ran)
}

// TestTxHooks_ClosedOutermostRunsInline asserts that once the outermost scope
// is done there is no commit left to wait for, so a late registration runs
// straight away rather than being queued into a sink nobody will fire.
func TestTxHooks_ClosedOutermostRunsInline(t *testing.T) {
	outer := &txHooks{}
	inner := &txHooks{parent: outer}

	outer.adopt(inner.close())
	outer.fire()

	ran := false
	inner.add(func() { ran = true })

	assert.True(t, ran)
	assert.Empty(t, outer.afterCommit)
}

// TestTxHooks_FireKeepsOrderOfLateRegistrations asserts that the firing loop
// picks up callbacks registered while it runs instead of letting them execute
// inline out of order, and only then closes the scope.
func TestTxHooks_FireKeepsOrderOfLateRegistrations(t *testing.T) {
	h := &txHooks{}

	var fired []string
	h.add(func() {
		fired = append(fired, "first")
		h.add(func() { fired = append(fired, "registered-while-firing") })
	})
	h.add(func() { fired = append(fired, "second") })

	h.fire()

	assert.Equal(t, []string{"first", "second", "registered-while-firing"}, fired)
	assert.True(t, h.closed, "the scope must close once the queue is empty")

	h.add(func() { fired = append(fired, "after-fire") })
	assert.Equal(t, []string{"first", "second", "registered-while-firing", "after-fire"}, fired)
}
