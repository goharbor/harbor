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

	"github.com/goharbor/harbor/src/controller/tag"
	"github.com/goharbor/harbor/src/server/v2.0/models"
)

// Tag model
type Tag struct {
	*tag.Tag
}

// ToSwagger converts the tag to the swagger model
func (t *Tag) ToSwagger() *models.Tag {
	return &models.Tag{
		ArtifactID:   t.ArtifactID,
		ID:           t.ID,
		Name:         t.Name,
		PullTime:     strfmt.DateTime(t.PullTime),
		PushTime:     strfmt.DateTime(t.PushTime),
		RepositoryID: t.RepositoryID,
		Immutable:    t.Immutable,
	}
}

// NewTag ...
func NewTag(t *tag.Tag) *Tag {
	return &Tag{Tag: t}
}
