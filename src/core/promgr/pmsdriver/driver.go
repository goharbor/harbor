// Copyright 2018 Project Harbor Authors
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

package pmsdriver

import (
	"github.com/goharbor/harbor/src/common/models"
)

// PMSDriver defines the operations that a project management service driver
// should implement
type PMSDriver interface {
	// Get a project by ID or name
	Get(projectIDOrName interface{}) (*models.Project, error)
	// Create a project
	Create(*models.Project) (int64, error)
	// Delete a project by ID or name
	Delete(projectIDOrName interface{}) error
	// Update the properties of a project
	Update(projectIDOrName interface{}, project *models.Project) error
	// List lists projects according to the query conditions
	List(query *models.ProjectQueryParam) (*models.ProjectQueryResult, error)
}
