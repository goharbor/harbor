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

package model

import (
	"github.com/go-openapi/strfmt"

	"github.com/goharbor/harbor/src/pkg/repository/model"
	"github.com/goharbor/harbor/src/server/v2.0/models"
)

// RepoRecord model
type RepoRecord struct {
	*model.RepoRecord
}

// ToSwagger converts the repository into the swagger model
func (r *RepoRecord) ToSwagger() *models.Repository {
	var createTime *strfmt.DateTime
	if !r.CreationTime.IsZero() {
		t := strfmt.DateTime(r.CreationTime)
		createTime = &t
	}

	return &models.Repository{
		CreationTime: createTime,
		Description:  r.Description,
		ID:           r.RepositoryID,
		Name:         r.Name,
		ProjectID:    r.ProjectID,
		PullCount:    r.PullCount,
		UpdateTime:   strfmt.DateTime(r.UpdateTime),
	}
}

// NewRepoRecord ...
func NewRepoRecord(r *model.RepoRecord) *RepoRecord {
	return &RepoRecord{RepoRecord: r}
}
