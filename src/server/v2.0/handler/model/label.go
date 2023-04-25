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

	"github.com/goharbor/harbor/src/pkg/label/model"
	"github.com/goharbor/harbor/src/server/v2.0/models"
)

// Label model
type Label struct {
	*model.Label
}

// ToSwagger converts the label to the swagger model
func (l *Label) ToSwagger() *models.Label {
	return &models.Label{
		Color:        l.Color,
		CreationTime: strfmt.DateTime(l.CreationTime),
		Description:  l.Description,
		ID:           l.ID,
		Name:         l.Name,
		ProjectID:    l.ProjectID,
		Scope:        l.Scope,
		UpdateTime:   strfmt.DateTime(l.UpdateTime),
	}
}

// NewLabel ...
func NewLabel(l *model.Label) *Label {
	return &Label{Label: l}
}
