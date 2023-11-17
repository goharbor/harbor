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

package immutable

import (
	"context"
	"fmt"

	"github.com/goharbor/harbor/src/lib/q"
	"github.com/goharbor/harbor/src/pkg/immutable"
	"github.com/goharbor/harbor/src/pkg/immutable/model"
)

var (
	// Ctr is a global variable for the default immutable controller implementation
	Ctr = NewAPIController(immutable.NewDefaultRuleManager())
)

// Controller to handle the requests related with immutable
type Controller interface {
	// GetImmutableRule ...
	GetImmutableRule(ctx context.Context, id int64) (*model.Metadata, error)

	// CreateImmutableRule ...
	CreateImmutableRule(ctx context.Context, m *model.Metadata) (int64, error)

	// DeleteImmutableRule ...
	DeleteImmutableRule(ctx context.Context, id int64) error

	// UpdateImmutableRule ...
	UpdateImmutableRule(ctx context.Context, projectID int64, m *model.Metadata) error

	// ListImmutableRules ...
	ListImmutableRules(ctx context.Context, query *q.Query) ([]*model.Metadata, error)

	// Count count the immutable rules
	Count(ctx context.Context, query *q.Query) (int64, error)

	// DeleteImmutableRuleByProject delete immuatable rules with project id
	DeleteImmutableRuleByProject(ctx context.Context, projectID int64) error
}

// DefaultAPIController ...
type DefaultAPIController struct {
	manager immutable.Manager
}

func (r *DefaultAPIController) DeleteImmutableRuleByProject(ctx context.Context, projectID int64) error {
	rules, err := r.ListImmutableRules(ctx, q.New(q.KeyWords{"ProjectID": projectID}))
	if err != nil {
		return err
	}
	for _, rule := range rules {
		if err = r.DeleteImmutableRule(ctx, rule.ID); err != nil {
			return err
		}
	}
	return nil
}

// GetImmutableRule ...
func (r *DefaultAPIController) GetImmutableRule(ctx context.Context, id int64) (*model.Metadata, error) {
	return r.manager.GetImmutableRule(ctx, id)
}

// DeleteImmutableRule ...
func (r *DefaultAPIController) DeleteImmutableRule(ctx context.Context, id int64) error {
	return r.manager.DeleteImmutableRule(ctx, id)
}

// CreateImmutableRule ...
func (r *DefaultAPIController) CreateImmutableRule(ctx context.Context, m *model.Metadata) (int64, error) {
	return r.manager.CreateImmutableRule(ctx, m)
}

// UpdateImmutableRule ...
func (r *DefaultAPIController) UpdateImmutableRule(ctx context.Context, projectID int64, m *model.Metadata) error {
	m0, err := r.manager.GetImmutableRule(ctx, m.ID)
	if err != nil {
		return err
	}
	if m0 == nil {
		return fmt.Errorf("the immutable tag rule is not found id:%v", m.ID)
	}
	if m0.Disabled != m.Disabled {
		return r.manager.EnableImmutableRule(ctx, m.ID, m.Disabled)
	}
	return r.manager.UpdateImmutableRule(ctx, projectID, m)
}

// ListImmutableRules ...
func (r *DefaultAPIController) ListImmutableRules(ctx context.Context, query *q.Query) ([]*model.Metadata, error) {
	return r.manager.ListImmutableRules(ctx, query)
}

// Count count the immutable rules
func (r *DefaultAPIController) Count(ctx context.Context, query *q.Query) (int64, error) {
	return r.manager.Count(ctx, query)
}

// NewAPIController ...
func NewAPIController(immutableMgr immutable.Manager) Controller {
	return &DefaultAPIController{
		manager: immutableMgr,
	}
}
