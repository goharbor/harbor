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

	"github.com/goharbor/harbor/src/pkg/accessory/model"
	"github.com/goharbor/harbor/src/server/v2.0/models"
)

// Accessory model
type Accessory struct {
	model.AccessoryData
}

// ToSwagger converts the label to the swagger model
func (a *Accessory) ToSwagger() *models.Accessory {
	return &models.Accessory{
		ID:                    a.ID,
		ArtifactID:            a.ArtifactID,
		SubjectArtifactID:     a.SubArtifactID,
		SubjectArtifactRepo:   a.SubArtifactRepo,
		SubjectArtifactDigest: a.SubArtifactDigest,
		Size:                  a.Size,
		Digest:                a.Digest,
		Type:                  a.Type,
		Icon:                  a.Icon,
		CreationTime:          strfmt.DateTime(a.CreatTime),
	}
}

// NewAccessory ...
func NewAccessory(a model.AccessoryData) *Accessory {
	return &Accessory{AccessoryData: a}
}
