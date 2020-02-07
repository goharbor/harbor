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

// FakeController is a fake repository controller that implement src/api/repository.Controller interface
type FakeController struct {
	mock.Mock
}

// Ensure ...
func (f *FakeController) Ensure(ctx context.Context, name string) (bool, int64, error) {
	args := f.Called()
	return args.Bool(0), int64(args.Int(1)), args.Error(2)
}

// List ...
func (f *FakeController) List(ctx context.Context, query *q.Query) (int64, []*models.RepoRecord, error) {
	args := f.Called()
	var repositories []*models.RepoRecord
	if args.Get(1) != nil {
		repositories = args.Get(1).([]*models.RepoRecord)
	}
	return int64(args.Int(0)), repositories, args.Error(2)

}

// Get ...
func (f *FakeController) Get(ctx context.Context, id int64) (*models.RepoRecord, error) {
	args := f.Called()
	var repository *models.RepoRecord
	if args.Get(0) != nil {
		repository = args.Get(0).(*models.RepoRecord)
	}
	return repository, args.Error(1)
}

// GetByName ...
func (f *FakeController) GetByName(ctx context.Context, name string) (*models.RepoRecord, error) {
	args := f.Called()
	var repository *models.RepoRecord
	if args.Get(0) != nil {
		repository = args.Get(0).(*models.RepoRecord)
	}
	return repository, args.Error(1)
}

// Delete ...
func (f *FakeController) Delete(ctx context.Context, id int64) error {
	args := f.Called()
	return args.Error(0)
}
