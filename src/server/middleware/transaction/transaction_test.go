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

package transaction

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	o "github.com/beego/beego/v2/client/orm"
	"github.com/stretchr/testify/assert"

	"github.com/goharbor/harbor/src/lib/errors"
	"github.com/goharbor/harbor/src/lib/orm"
	"github.com/goharbor/harbor/src/server/middleware"
)

type mockOrmer struct {
	o.Ormer
	tx       mockTxOrmer
	beginErr error
}

func (m *mockOrmer) Begin() (o.TxOrmer, error) {
	return &m.tx, m.beginErr
}

func (m *mockOrmer) Reset() {
	m.tx.Reset()
	m.beginErr = nil
}

type mockTxOrmer struct {
	o.TxOrmer
	commitErr error
	records   []any
}

func (m *mockTxOrmer) Insert(i any) (int64, error) {
	m.records = append(m.records, i)
	return int64(len(m.records)), nil
}

func (m *mockTxOrmer) Commit() error {
	return m.commitErr
}

func (m *mockTxOrmer) Rollback() error {
	m.ResetRecords()
	return nil
}

func (m *mockTxOrmer) ResetRecords() {
	m.records = nil
}

func (m *mockTxOrmer) Reset() {
	m.ResetRecords()
	m.commitErr = nil
}

func TestTransaction(t *testing.T) {
	assert := assert.New(t)

	tx := mockTxOrmer{}
	mo := &mockOrmer{
		tx: tx,
	}

	newRequest := func(method, target string, body io.Reader) *http.Request {
		req := httptest.NewRequest(http.MethodGet, "/req1", nil)
		return req.WithContext(orm.NewContext(req.Context(), mo))
	}

	next := func(status int) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			mo.tx.Insert("record1")
			w.WriteHeader(status)
		})
	}

	// test response status code accepted
	req1 := newRequest(http.MethodGet, "/req", nil)
	rec1 := httptest.NewRecorder()
	Middleware()(next(http.StatusOK)).ServeHTTP(rec1, req1)
	assert.Equal(http.StatusOK, rec1.Code)
	assert.NotEmpty(mo.tx.records)

	mo.tx.ResetRecords()
	assert.Empty(mo.tx.records)

	// test response status code not accepted
	req2 := newRequest(http.MethodGet, "/req", nil)
	rec2 := httptest.NewRecorder()
	Middleware()(next(http.StatusBadRequest)).ServeHTTP(rec2, req2)
	assert.Equal(http.StatusBadRequest, rec2.Code)
	assert.Empty(mo.tx.records)

	// test begin transaction failed
	mo.beginErr = errors.New("begin tx failed")
	req3 := newRequest(http.MethodGet, "/req", nil)
	rec3 := httptest.NewRecorder()
	Middleware()(next(http.StatusBadRequest)).ServeHTTP(rec3, req3)
	assert.Equal(http.StatusInternalServerError, rec3.Code)
	assert.Empty(mo.tx.records)

	// test commit transaction failed
	mo.beginErr = nil
	mo.tx.commitErr = errors.New("commit tx failed")
	req4 := newRequest(http.MethodGet, "/req", nil)
	rec4 := httptest.NewRecorder()
	Middleware()(next(http.StatusOK)).ServeHTTP(rec4, req4)
	assert.Equal(http.StatusInternalServerError, rec4.Code)

	// test MustCommit
	mo.Reset()
	assert.Empty(mo.tx.records)

	txMustCommit := func(status int) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			defer MustCommit(r)
			mo.tx.Insert("record1")
			w.WriteHeader(status)
		})
	}

	req5 := newRequest(http.MethodGet, "/req", nil)
	rec5 := httptest.NewRecorder()

	m1 := middleware.New(func(w http.ResponseWriter, r *http.Request, next http.Handler) {
		type key struct{}
		next.ServeHTTP(w, r.WithContext(context.WithValue(r.Context(), key{}, "value")))
	})

	Middleware()(m1((txMustCommit(http.StatusBadRequest)))).ServeHTTP(rec5, req5)
	assert.Equal(http.StatusBadRequest, rec2.Code)
	assert.NotEmpty(mo.tx.records)
}

func TestMustCommit(t *testing.T) {
	newRequest := func(ctx context.Context) *http.Request {
		req := httptest.NewRequest(http.MethodGet, "/req", nil)
		return req.WithContext(ctx)
	}

	ctx := context.Background()
	committableCtx := context.WithValue(ctx, orm.CommittedKey{}, new(bool))

	type args struct {
		r *http.Request
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{"request committable", args{newRequest(committableCtx)}, false},
		{"request not committable", args{newRequest(ctx)}, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := MustCommit(tt.args.r); (err != nil) != tt.wantErr {
				t.Errorf("MustCommit() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
