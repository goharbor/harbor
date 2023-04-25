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

	"github.com/goharbor/harbor/src/pkg/allowlist/models"
	svrmodels "github.com/goharbor/harbor/src/server/v2.0/models"
)

// CVEAllowlist model
type CVEAllowlist struct {
	*models.CVEAllowlist
}

// ToSwagger converts the model to swagger model
func (l *CVEAllowlist) ToSwagger() *svrmodels.CVEAllowlist {
	res := &svrmodels.CVEAllowlist{
		ID:           l.ID,
		Items:        []*svrmodels.CVEAllowlistItem{},
		ProjectID:    l.ProjectID,
		ExpiresAt:    l.ExpiresAt,
		CreationTime: strfmt.DateTime(l.CreationTime),
		UpdateTime:   strfmt.DateTime(l.UpdateTime),
	}
	for _, it := range l.Items {
		cveItem := &svrmodels.CVEAllowlistItem{
			CVEID: it.CVEID,
		}
		res.Items = append(res.Items, cveItem)
	}
	return res
}

// NewCVEAllowlist ...
func NewCVEAllowlist(l *models.CVEAllowlist) *CVEAllowlist {
	return &CVEAllowlist{l}
}
