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

package index

import (
	"sync"

	"github.com/goharbor/harbor/src/pkg/retention/policy/rule"

	"github.com/goharbor/harbor/src/pkg/retention/policy/rule/latestpl"

	"github.com/goharbor/harbor/src/pkg/retention/policy/rule/latestk"

	"github.com/goharbor/harbor/src/pkg/retention/policy/rule/always"

	"github.com/goharbor/harbor/src/pkg/retention/policy/rule/lastx"

	"github.com/goharbor/harbor/src/pkg/retention/policy/rule/latestps"

	"github.com/goharbor/harbor/src/pkg/retention/policy/action"

	"github.com/pkg/errors"
)

// index for keeping the mapping between template ID and evaluator
var index sync.Map

// Metadata defines metadata for rule registration
type Metadata struct {
	TemplateID string `json:"rule_template"`

	// Action of the rule performs
	// "retain"
	Action string `json:"action"`

	Parameters []*IndexedParam `json:"params"`
}

// IndexedParam declares the param info
type IndexedParam struct {
	Name string `json:"name"`

	// Type of the param
	// "int", "string" or "[]string"
	Type string `json:"type"`

	Unit string `json:"unit"`

	Required bool `json:"required"`
}

// indexedItem is the item saved in the sync map
type indexedItem struct {
	Meta *Metadata

	Factory rule.Factory
}

func init() {
	// Register latest pushed
	Register(&Metadata{
		TemplateID: latestps.TemplateID,
		Action:     action.Retain,
		Parameters: []*IndexedParam{
			{
				Name:     latestps.ParameterK,
				Type:     "int",
				Unit:     "count",
				Required: true,
			},
		},
	}, latestps.New)

	// Register latest pulled
	Register(&Metadata{
		TemplateID: latestpl.TemplateID,
		Action:     action.Retain,
		Parameters: []*IndexedParam{
			{
				Name:     latestpl.ParameterN,
				Type:     "int",
				Unit:     "count",
				Required: true,
			},
		},
	}, latestpl.New)

	// Register latest active
	Register(&Metadata{
		TemplateID: latestk.TemplateID,
		Action:     action.Retain,
		Parameters: []*IndexedParam{
			{
				Name:     latestk.ParameterK,
				Type:     "int",
				Unit:     "count",
				Required: true,
			},
		},
	}, latestk.New)

	// Register lastx
	Register(&Metadata{
		TemplateID: lastx.TemplateID,
		Action:     action.Retain,
		Parameters: []*IndexedParam{
			{
				Name:     lastx.ParameterX,
				Type:     "int",
				Unit:     "days",
				Required: true,
			},
		},
	}, lastx.New)

	// Register always
	Register(&Metadata{
		TemplateID: always.TemplateID,
		Action:     action.Retain,
		Parameters: []*IndexedParam{},
	}, always.New)
}

// Register the rule evaluator with the corresponding rule template
func Register(meta *Metadata, factory rule.Factory) {
	if meta == nil || factory == nil || len(meta.TemplateID) == 0 {
		// do nothing
		return
	}

	index.Store(meta.TemplateID, &indexedItem{
		Meta:    meta,
		Factory: factory,
	})
}

// Get rule evaluator with the provided template ID
func Get(templateID string, parameters rule.Parameters) (rule.Evaluator, error) {
	if len(templateID) == 0 {
		return nil, errors.New("empty rule template ID")
	}

	v, ok := index.Load(templateID)
	if !ok {
		return nil, errors.Errorf("rule evaluator %s is not registered", templateID)
	}

	item := v.(*indexedItem)

	// We can check more things if we want to do in the future
	if len(item.Meta.Parameters) > 0 {
		for _, p := range item.Meta.Parameters {
			if p.Required {
				exists := parameters != nil
				if exists {
					_, exists = parameters[p.Name]
				}

				if !exists {
					return nil, errors.Errorf("missing required parameter %s for rule %s", p.Name, templateID)
				}
			}
		}
	}
	factory := item.Factory

	return factory(parameters), nil
}

// Index returns all the metadata of the registered rules
func Index() []*Metadata {
	res := make([]*Metadata, 0)

	index.Range(func(k, v interface{}) bool {
		if item, ok := v.(*indexedItem); ok {
			res = append(res, item.Meta)
			return true
		}

		return false
	})

	return res
}
