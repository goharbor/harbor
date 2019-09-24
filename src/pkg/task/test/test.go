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

package test

import (
	"github.com/goharbor/harbor/src/pkg/task/model"
	"github.com/stretchr/testify/mock"
)

// MockTaskManager is a mock implementation for task manager
type MockTaskManager struct {
	mock.Mock
}

// Create ...
func (m *MockTaskManager) Create(task *model.Task) (int64, error) {
	args := m.Called(task)
	return int64(args.Int(0)), args.Error(1)
}

// Get ...
func (m *MockTaskManager) Get(id int64) (*model.Task, error) {
	args := m.Called(id)
	var task *model.Task
	if args.Get(0) != nil {
		task = args.Get(0).(*model.Task)
	}
	return task, args.Error(1)
}

// Update ...
func (m *MockTaskManager) Update(task *model.Task, cols ...string) error {
	args := m.Called(task)
	return args.Error(0)
}

// UpdateStatus ...
func (m *MockTaskManager) UpdateStatus(id int64, status string, statusCode int, statusRevision int64) error {
	args := m.Called()
	return args.Error(0)
}

// Delete ...
func (m *MockTaskManager) Delete(id int64) error {
	args := m.Called(id)
	return args.Error(0)
}

// AppendCheckInData ...
func (m *MockTaskManager) AppendCheckInData(id int64, data string) error {
	args := m.Called()
	return args.Error(0)
}

// CalculateTaskGroupStatus ...
func (m *MockTaskManager) CalculateTaskGroupStatus(groupID int64) (*model.GroupStatus, error) {
	args := m.Called()
	return args.Get(0).(*model.GroupStatus), args.Error(1)
}
