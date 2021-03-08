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
