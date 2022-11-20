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
	"database/sql"
	"github.com/beego/beego/v2/core/utils"

	"github.com/beego/beego/v2/client/orm"
)

// FakeOrmer ...
type FakeOrmer struct {
}

func (f *FakeOrmer) LoadRelated(md interface{}, name string, args ...utils.KV) (int64, error) {
	return 0, nil
}

func (f *FakeOrmer) QueryM2MWithCtx(ctx context.Context, md interface{}, name string) orm.QueryM2Mer {
	return nil
}

func (f *FakeOrmer) QueryTableWithCtx(ctx context.Context, ptrStructOrTableName interface{}) orm.QuerySeter {
	return nil
}

func (f *FakeOrmer) InsertWithCtx(ctx context.Context, md interface{}) (int64, error) {
	return 0, nil
}

func (f *FakeOrmer) InsertOrUpdateWithCtx(ctx context.Context, md interface{}, colConflitAndArgs ...string) (int64, error) {
	return 0, nil
}

func (f *FakeOrmer) InsertMultiWithCtx(ctx context.Context, bulk int, mds interface{}) (int64, error) {
	return 0, nil
}

func (f *FakeOrmer) UpdateWithCtx(ctx context.Context, md interface{}, cols ...string) (int64, error) {
	return 0, nil
}

func (f *FakeOrmer) DeleteWithCtx(ctx context.Context, md interface{}, cols ...string) (int64, error) {
	return 0, nil
}

func (f *FakeOrmer) RawWithCtx(ctx context.Context, query string, args ...interface{}) orm.RawSeter {
	return nil
}

func (f *FakeOrmer) Begin() (orm.TxOrmer, error) {
	return &FakeTxOrmer{}, nil
}

func (f *FakeOrmer) BeginWithCtx(ctx context.Context) (orm.TxOrmer, error) {
	return nil, nil
}

func (f *FakeOrmer) BeginWithOpts(opts *sql.TxOptions) (orm.TxOrmer, error) {
	return nil, nil
}

func (f *FakeOrmer) BeginWithCtxAndOpts(ctx context.Context, opts *sql.TxOptions) (orm.TxOrmer, error) {
	return nil, nil
}

func (f *FakeOrmer) DoTx(task func(ctx context.Context, txOrm orm.TxOrmer) error) error {
	return nil
}

func (f *FakeOrmer) DoTxWithCtx(ctx context.Context, task func(ctx context.Context, txOrm orm.TxOrmer) error) error {
	return nil
}

func (f *FakeOrmer) DoTxWithOpts(opts *sql.TxOptions, task func(ctx context.Context, txOrm orm.TxOrmer) error) error {
	return nil
}

func (f *FakeOrmer) DoTxWithCtxAndOpts(ctx context.Context, opts *sql.TxOptions, task func(ctx context.Context, txOrm orm.TxOrmer) error) error {
	return nil
}

// Read ...
func (f *FakeOrmer) Read(md interface{}, cols ...string) error {
	return nil
}

func (f *FakeOrmer) ReadWithCtx(ctx context.Context, md interface{}, cols ...string) error {
	return nil
}

func (f *FakeOrmer) ReadForUpdateWithCtx(ctx context.Context, md interface{}, cols ...string) error {
	return nil
}

func (f *FakeOrmer) ReadOrCreateWithCtx(ctx context.Context, md interface{}, col1 string, cols ...string) (bool, int64, error) {
	return false, 0, nil
}

func (f *FakeOrmer) LoadRelatedWithCtx(_ context.Context, md interface{}, name string, args ...utils.KV) (int64, error) {
	return 0, nil
}

// ReadForUpdate ...
func (f *FakeOrmer) ReadForUpdate(md interface{}, cols ...string) error {
	return nil
}

// ReadOrCreate ...
func (f *FakeOrmer) ReadOrCreate(md interface{}, col1 string, cols ...string) (bool, int64, error) {
	return false, 0, nil
}

// Insert ...
func (f *FakeOrmer) Insert(interface{}) (int64, error) {
	return 0, nil
}

// InsertOrUpdate ...
func (f *FakeOrmer) InsertOrUpdate(md interface{}, colConflitAndArgs ...string) (int64, error) {
	return 0, nil
}

// InsertMulti ...
func (f *FakeOrmer) InsertMulti(bulk int, mds interface{}) (int64, error) {
	return 0, nil
}

// Update ...
func (f *FakeOrmer) Update(md interface{}, cols ...string) (int64, error) {
	return 0, nil
}

// Delete ...
func (f *FakeOrmer) Delete(md interface{}, cols ...string) (int64, error) {
	return 0, nil
}

// QueryM2M ...
func (f *FakeOrmer) QueryM2M(md interface{}, name string) orm.QueryM2Mer {
	return nil
}

// QueryTable ...
func (f *FakeOrmer) QueryTable(ptrStructOrTableName interface{}) orm.QuerySeter {
	return nil
}

// Using ...
func (f *FakeOrmer) Using(name string) error {
	return nil
}

// BeginTx ...
func (f *FakeOrmer) BeginTx(ctx context.Context, opts *sql.TxOptions) error {
	return nil
}

// Commit ...
func (f *FakeOrmer) Commit() error {
	return f.Commit()
}

// Rollback ...
func (f *FakeOrmer) Rollback() error {
	return f.Rollback()
}

// Raw ...
func (f *FakeOrmer) Raw(query string, args ...interface{}) orm.RawSeter {
	return &FakeRawSeter{}
}

// Driver ...
func (f *FakeOrmer) Driver() orm.Driver {
	return nil
}

// DBStats ...
func (f *FakeOrmer) DBStats() *sql.DBStats {
	return nil
}

// FakeTxOrmer ...
type FakeTxOrmer struct {
}

func (f *FakeTxOrmer) Read(md interface{}, cols ...string) error {
	return nil
}

func (f *FakeTxOrmer) ReadWithCtx(ctx context.Context, md interface{}, cols ...string) error {
	return nil
}

func (f *FakeTxOrmer) ReadForUpdate(md interface{}, cols ...string) error {
	return nil
}

func (f *FakeTxOrmer) ReadForUpdateWithCtx(ctx context.Context, md interface{}, cols ...string) error {
	return nil
}

func (f *FakeTxOrmer) ReadOrCreate(md interface{}, col1 string, cols ...string) (bool, int64, error) {
	return false, 0, nil
}

func (f *FakeTxOrmer) ReadOrCreateWithCtx(ctx context.Context, md interface{}, col1 string, cols ...string) (bool, int64, error) {
	return false, 0, nil
}

func (f *FakeTxOrmer) LoadRelated(md interface{}, name string, args ...utils.KV) (int64, error) {
	return 0, nil
}

func (f *FakeTxOrmer) LoadRelatedWithCtx(ctx context.Context, md interface{}, name string, args ...utils.KV) (int64, error) {
	return 0, nil
}

func (f *FakeTxOrmer) QueryM2M(md interface{}, name string) orm.QueryM2Mer {
	return nil
}

func (f *FakeTxOrmer) QueryM2MWithCtx(ctx context.Context, md interface{}, name string) orm.QueryM2Mer {
	return nil
}

func (f *FakeTxOrmer) QueryTable(ptrStructOrTableName interface{}) orm.QuerySeter {
	return nil
}

func (f *FakeTxOrmer) QueryTableWithCtx(ctx context.Context, ptrStructOrTableName interface{}) orm.QuerySeter {
	return nil
}

func (f *FakeTxOrmer) DBStats() *sql.DBStats {
	return nil
}

func (f *FakeTxOrmer) Insert(md interface{}) (int64, error) {
	return 0, nil
}

func (f *FakeTxOrmer) InsertWithCtx(ctx context.Context, md interface{}) (int64, error) {
	return 0, nil
}

func (f *FakeTxOrmer) InsertOrUpdate(md interface{}, colConflitAndArgs ...string) (int64, error) {
	return 0, nil
}

func (f *FakeTxOrmer) InsertOrUpdateWithCtx(ctx context.Context, md interface{}, colConflitAndArgs ...string) (int64, error) {
	return 0, nil
}

func (f *FakeTxOrmer) InsertMulti(bulk int, mds interface{}) (int64, error) {
	return 0, nil
}

func (f *FakeTxOrmer) InsertMultiWithCtx(ctx context.Context, bulk int, mds interface{}) (int64, error) {
	return 0, nil
}

func (f *FakeTxOrmer) Update(md interface{}, cols ...string) (int64, error) {
	return 0, nil
}

func (f *FakeTxOrmer) UpdateWithCtx(ctx context.Context, md interface{}, cols ...string) (int64, error) {
	return 0, nil
}

func (f *FakeTxOrmer) Delete(md interface{}, cols ...string) (int64, error) {
	return 0, nil
}

func (f *FakeTxOrmer) DeleteWithCtx(ctx context.Context, md interface{}, cols ...string) (int64, error) {
	return 0, nil
}

func (f *FakeTxOrmer) Raw(query string, args ...interface{}) orm.RawSeter {
	return &FakeRawSeter{}
}

func (f *FakeTxOrmer) RawWithCtx(ctx context.Context, query string, args ...interface{}) orm.RawSeter {
	return nil
}

func (f *FakeTxOrmer) Driver() orm.Driver {
	return nil
}

func (f *FakeTxOrmer) Commit() error {
	return nil
}

func (f *FakeTxOrmer) Rollback() error {
	return nil
}

func (f *FakeTxOrmer) RollbackUnlessCommit() error {
	return nil
}

// FakeTxOrmer ...
type FakeRawSeter struct {
}

func (f FakeRawSeter) Exec() (sql.Result, error) {
	return nil, nil
}

func (f FakeRawSeter) QueryRow(containers ...interface{}) error {
	return nil
}

func (f FakeRawSeter) QueryRows(containers ...interface{}) (int64, error) {
	return 0, nil
}

func (f FakeRawSeter) SetArgs(i ...interface{}) orm.RawSeter {
	return nil
}

func (f FakeRawSeter) Values(container *[]orm.Params, cols ...string) (int64, error) {
	return 0, nil
}

func (f FakeRawSeter) ValuesList(container *[]orm.ParamsList, cols ...string) (int64, error) {
	return 0, nil
}

func (f FakeRawSeter) ValuesFlat(container *orm.ParamsList, cols ...string) (int64, error) {
	return 0, nil
}

func (f FakeRawSeter) RowsToMap(result *orm.Params, keyCol, valueCol string) (int64, error) {
	return 0, nil
}

func (f FakeRawSeter) RowsToStruct(ptrStruct interface{}, keyCol, valueCol string) (int64, error) {
	return 0, nil
}

func (f FakeRawSeter) Prepare() (orm.RawPreparer, error) {
	return nil, nil
}
