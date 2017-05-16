// Copyright (c) 2017 VMware, Inc. All Rights Reserved.
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

package projectmanager

import (
	"github.com/vmware/harbor/src/common/models"
)

// ProjectManager is the project mamager which abstracts the operations related
// to projects
type ProjectManager interface {
	Get(projectIDOrName interface{}) *models.Project
	IsPublic(projectIDOrName interface{}) bool
	Exist(projectIDOrName interface{}) bool
	GetRoles(username string, projectIDOrName interface{}) []int
	// get all public project
	GetPublic() []*models.Project
	// get projects which the user is a member of
	GetByMember(username string) []*models.Project
	Create(*models.Project) (int64, error)
	Delete(projectIDOrName interface{}) error
	Update(projectIDOrName interface{}, project *models.Project) error
	// GetAll returns a project list and the total count according to
	// the query conditions:
	// owner: username of owner
	// name: name of project
	// public: public or not, can be "true", "false" or ""
	// member: username of the member
	// role: the role of member specified by member parameter
	// page, size: pagination parameters
	GetAll(owner, name, public, member string, role int, page,
		size int64) ([]*models.Project, int64)
}
