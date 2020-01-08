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
	"github.com/goharbor/harbor/src/common/models"
	"github.com/goharbor/harbor/src/pkg/q"
	"github.com/stretchr/testify/mock"
)

// FakeRepositoryManager is a fake repository manager that implement src/pkg/repository.Manager interface
type FakeRepositoryManager struct {
	mock.Mock
}

// List ...
func (f *FakeRepositoryManager) List(ctx context.Context, query *q.Query) (int64, []*models.RepoRecord, error) {
	args := f.Called()
	return int64(args.Int(0)), args.Get(1).([]*models.RepoRecord), args.Error(2)
}

// Get ...
func (f *FakeRepositoryManager) Get(ctx context.Context, id int64) (*models.RepoRecord, error) {
	args := f.Called()
	return args.Get(0).(*models.RepoRecord), args.Error(1)
}

// Delete ...
func (f *FakeRepositoryManager) Delete(ctx context.Context, id int64) error {
	args := f.Called()
	return args.Error(0)
}

// Create ...
func (f *FakeRepositoryManager) Create(ctx context.Context, repository *models.RepoRecord) (int64, error) {
	args := f.Called()
	return int64(args.Int(0)), args.Error(1)
}

// Update ...
func (f *FakeRepositoryManager) Update(ctx context.Context, repository *models.RepoRecord, props ...string) error {
	args := f.Called()
	return args.Error(0)
}
