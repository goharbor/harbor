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

	"github.com/beego/beego/v2/client/orm"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// fakeTxOrmer is a minimal orm.TxOrmer that only records Commit/Rollback
// calls; WithTransaction never touches any other method in these tests.
type fakeTxOrmer struct {
	orm.TxOrmer
	commits   int
	rollbacks int
}

func (f *fakeTxOrmer) Commit() error {
	f.commits++
	return nil
}

func (f *fakeTxOrmer) Rollback() error {
	f.rollbacks++
	return nil
}

// fakeOrmer is a minimal orm.Ormer that hands out fakeTxOrmer from
// BeginWithCtx and records the context it was called with.
type fakeOrmer struct {
	orm.Ormer
	tx       *fakeTxOrmer
	beginCtx context.Context
}

func (f *fakeOrmer) BeginWithCtx(ctx context.Context) (orm.TxOrmer, error) {
	f.beginCtx = ctx
	return f.tx, nil
}

func TestWithTransaction_PanicRollsBackInsteadOfLeaking(t *testing.T) {
	tx := &fakeTxOrmer{}
	ctx := NewContext(context.Background(), &fakeOrmer{tx: tx})

	wrapped := WithTransaction(func(context.Context) error {
		panic("boom")
	})

	require.Panics(t, func() {
		_ = wrapped(ctx)
	})

	assert.Equal(t, 1, tx.rollbacks, "a panic inside the wrapped function must still roll back the transaction")
	assert.Equal(t, 0, tx.commits)
}

func TestWithTransaction_BeginUsesRequestContext(t *testing.T) {
	tx := &fakeTxOrmer{}
	o := &fakeOrmer{tx: tx}

	type markerKey struct{}
	ctx := context.WithValue(context.Background(), markerKey{}, "request-ctx")
	ctx = NewContext(ctx, o)

	wrapped := WithTransaction(func(context.Context) error {
		return nil
	})
	require.NoError(t, wrapped(ctx))

	require.NotNil(t, o.beginCtx)
	assert.Equal(t, "request-ctx", o.beginCtx.Value(markerKey{}), "transaction must begin with the request context so cancellation can roll it back")
	assert.Equal(t, 1, tx.commits)
}
