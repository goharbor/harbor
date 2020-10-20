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

package models

import (
	"github.com/goharbor/harbor/src/common/models"
)

// Project ...
type Project = models.Project

// Projects the connection for Project
type Projects []*models.Project

// OwnerIDs returns all the owner ids from the projects
func (projects Projects) OwnerIDs() []int {
	var ownerIDs []int
	for _, project := range projects {
		ownerIDs = append(ownerIDs, project.OwnerID)
	}
	return ownerIDs
}

// Member ...
type Member = models.Member

// MemberQuery ...
type MemberQuery = models.MemberQuery
