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

package testing

import (
	"context"
	"github.com/goharbor/harbor/src/pkg/q"
	"github.com/goharbor/harbor/src/pkg/tag/model/tag"
	"github.com/stretchr/testify/mock"
)

// FakeTagManager is a fake tag manager that implement the src/pkg/tag.Manager interface
type FakeTagManager struct {
	mock.Mock
}

// List ...
func (f *FakeTagManager) List(ctx context.Context, query *q.Query) (int64, []*tag.Tag, error) {
	args := f.Called()
	return int64(args.Int(0)), args.Get(1).([]*tag.Tag), args.Error(2)
}

// Get ...
func (f *FakeTagManager) Get(ctx context.Context, id int64) (*tag.Tag, error) {
	args := f.Called()
	return args.Get(0).(*tag.Tag), args.Error(1)
}

// Create ...
func (f *FakeTagManager) Create(ctx context.Context, tag *tag.Tag) (int64, error) {
	args := f.Called()
	return int64(args.Int(0)), args.Error(1)
}

// Update ...
func (f *FakeTagManager) Update(ctx context.Context, tag *tag.Tag, props ...string) error {
	args := f.Called()
	return args.Error(0)
}

// Delete ...
func (f *FakeTagManager) Delete(ctx context.Context, id int64) error {
	args := f.Called()
	return args.Error(0)
}
