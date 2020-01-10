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

package lib

import "github.com/astaxie/beego/orm"

// FakeOrmer is a fake ormer that implement github.com/astaxie/beego/orm.Ormer interface
type FakeOrmer struct{}

// Read ...
func (f *FakeOrmer) Read(md interface{}, cols ...string) error {
	return nil
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

// LoadRelated ...
func (f *FakeOrmer) LoadRelated(md interface{}, name string, args ...interface{}) (int64, error) {
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

// Begin ...
func (f *FakeOrmer) Begin() error {
	return nil
}

// Commit ...
func (f *FakeOrmer) Commit() error {
	return nil
}

// Rollback ...
func (f *FakeOrmer) Rollback() error {
	return nil
}

// Raw ...
func (f *FakeOrmer) Raw(query string, args ...interface{}) orm.RawSeter {
	return nil
}

// Driver ...
func (f *FakeOrmer) Driver() orm.Driver {
	return nil
}
