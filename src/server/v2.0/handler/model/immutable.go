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
	pkg_model "github.com/goharbor/harbor/src/pkg/immutable/model"
	"github.com/goharbor/harbor/src/server/v2.0/models"
)

// ImmutableRule ...
type ImmutableRule struct {
	*pkg_model.Metadata
}

// ToSwagger ...
func (ir *ImmutableRule) ToSwagger() *models.ImmutableRule {
	return &models.ImmutableRule{
		ID:             ir.ID,
		Disabled:       ir.Disabled,
		Action:         ir.Action,
		Priority:       int64(ir.Priority),
		ScopeSelectors: ir.ToScopeSelectors(),
		TagSelectors:   ir.ToTagSelectors(),
		Template:       ir.Template,
	}
}

// ToTagSelectors ...
func (ir *ImmutableRule) ToTagSelectors() []*models.ImmutableSelector {
	var results []*models.ImmutableSelector
	for _, t := range ir.TagSelectors {
		results = append(results, &models.ImmutableSelector{
			Decoration: t.Decoration,
			Kind:       t.Kind,
			Pattern:    t.Pattern,
		})
	}
	return results
}

// ToScopeSelectors ...
func (ir *ImmutableRule) ToScopeSelectors() map[string][]models.ImmutableSelector {
	results := map[string][]models.ImmutableSelector{}
	for k, v := range ir.ScopeSelectors {
		var scopeSelectors []models.ImmutableSelector
		for _, s := range v {
			scopeSelectors = append(scopeSelectors, models.ImmutableSelector{
				Decoration: s.Decoration,
				Kind:       s.Kind,
				Pattern:    s.Pattern,
			})
		}
		results[k] = scopeSelectors
	}
	return results
}

// NewImmutableRule ...
func NewImmutableRule(meta *pkg_model.Metadata) *ImmutableRule {
	return &ImmutableRule{
		Metadata: meta,
	}
}
