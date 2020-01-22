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

package project

import (
	"github.com/goharbor/harbor/src/common/models"
	"github.com/stretchr/testify/mock"
)

// FakeManager is a fake project manager that implement src/pkg/project.Manager interface
type FakeManager struct {
	mock.Mock
}

// List ...
func (f *FakeManager) List(query ...*models.ProjectQueryParam) ([]*models.Project, error) {
	args := f.Called()
	var projects []*models.Project
	if args.Get(0) != nil {
		projects = args.Get(0).([]*models.Project)
	}
	return projects, args.Error(1)
}

// Get ...
func (f *FakeManager) Get(interface{}) (*models.Project, error) {
	args := f.Called()
	var project *models.Project
	if args.Get(0) != nil {
		project = args.Get(0).(*models.Project)
	}
	return project, args.Error(1)
}
