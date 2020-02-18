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

package repository

import (
	"context"
	"github.com/goharbor/harbor/src/common/models"
	"github.com/goharbor/harbor/src/pkg/q"
	"github.com/stretchr/testify/mock"
)

// FakeManager is a fake repository manager that implement src/pkg/repository.Manager interface
type FakeManager struct {
	mock.Mock
}

// Count ...
func (f *FakeManager) Count(ctx context.Context, query *q.Query) (int64, error) {
	args := f.Called()
	return int64(args.Int(0)), args.Error(1)
}

// List ...
func (f *FakeManager) List(ctx context.Context, query *q.Query) ([]*models.RepoRecord, error) {
	args := f.Called()
	var repositories []*models.RepoRecord
	if args.Get(0) != nil {
		repositories = args.Get(0).([]*models.RepoRecord)
	}
	return repositories, args.Error(1)
}

// Get ...
func (f *FakeManager) Get(ctx context.Context, id int64) (*models.RepoRecord, error) {
	args := f.Called()
	var repository *models.RepoRecord
	if args.Get(0) != nil {
		repository = args.Get(0).(*models.RepoRecord)
	}
	return repository, args.Error(1)
}

// GetByName ...
func (f *FakeManager) GetByName(ctx context.Context, name string) (*models.RepoRecord, error) {
	args := f.Called()
	var repository *models.RepoRecord
	if args.Get(0) != nil {
		repository = args.Get(0).(*models.RepoRecord)
	}
	return repository, args.Error(1)
}

// Delete ...
func (f *FakeManager) Delete(ctx context.Context, id int64) error {
	args := f.Called()
	return args.Error(0)
}

// Create ...
func (f *FakeManager) Create(ctx context.Context, repository *models.RepoRecord) (int64, error) {
	args := f.Called()
	return int64(args.Int(0)), args.Error(1)
}

// Update ...
func (f *FakeManager) Update(ctx context.Context, repository *models.RepoRecord, props ...string) error {
	args := f.Called()
	return args.Error(0)
}
