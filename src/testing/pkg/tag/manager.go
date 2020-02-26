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

package tag

import (
	"context"
	"github.com/goharbor/harbor/src/pkg/q"
	"github.com/goharbor/harbor/src/pkg/tag/model/tag"
	"github.com/stretchr/testify/mock"
)

// FakeManager is a fake tag manager that implement the src/pkg/tag.Manager interface
type FakeManager struct {
	mock.Mock
}

// Count ...
func (f *FakeManager) Count(ctx context.Context, query *q.Query) (int64, error) {
	args := f.Called()
	return int64(args.Int(0)), args.Error(1)
}

// List ...
func (f *FakeManager) List(ctx context.Context, query *q.Query) ([]*tag.Tag, error) {
	args := f.Called()
	var tags []*tag.Tag
	if args.Get(0) != nil {
		tags = args.Get(0).([]*tag.Tag)
	}
	return tags, args.Error(1)
}

// Get ...
func (f *FakeManager) Get(ctx context.Context, id int64) (*tag.Tag, error) {
	args := f.Called()
	var tg *tag.Tag
	if args.Get(0) != nil {
		tg = args.Get(0).(*tag.Tag)
	}
	return tg, args.Error(1)
}

// Create ...
func (f *FakeManager) Create(ctx context.Context, tag *tag.Tag) (int64, error) {
	args := f.Called()
	return int64(args.Int(0)), args.Error(1)
}

// GetOrCreate ...
func (f *FakeManager) GetOrCreate(ctx context.Context, tag *tag.Tag) (bool, int64, error) {
	args := f.Called()
	return args.Bool(0), int64(args.Int(1)), args.Error(2)
}

// Update ...
func (f *FakeManager) Update(ctx context.Context, tag *tag.Tag, props ...string) error {
	args := f.Called()
	return args.Error(0)
}

// Delete ...
func (f *FakeManager) Delete(ctx context.Context, id int64) error {
	args := f.Called()
	return args.Error(0)
}

// DeleteOfArtifact ...
func (f *FakeManager) DeleteOfArtifact(ctx context.Context, artifactID int64) error {
	args := f.Called()
	return args.Error(0)
}
