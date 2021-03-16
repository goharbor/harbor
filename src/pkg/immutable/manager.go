package immutable

import (
	"context"
	"encoding/json"
	"github.com/goharbor/harbor/src/lib/q"
	"sort"

	"github.com/goharbor/harbor/src/pkg/immutable/dao"
	dao_model "github.com/goharbor/harbor/src/pkg/immutable/dao/model"
	"github.com/goharbor/harbor/src/pkg/immutable/model"
)

var (
	// Mgr is a global variable for the default immutablerule manager implementation
	Mgr = NewDefaultRuleManager()
)

// Manager ...
type Manager interface {
	// CreateImmutableRule creates the Immutable Rule
	CreateImmutableRule(ctx context.Context, m *model.Metadata) (int64, error)
	// UpdateImmutableRule update the immutable rules
	UpdateImmutableRule(ctx context.Context, projectID int64, ir *model.Metadata) error
	// EnableImmutableRule enable/disable immutable rules
	EnableImmutableRule(ctx context.Context, id int64, enabled bool) error
	// GetImmutableRule get immutable rule
	GetImmutableRule(ctx context.Context, id int64) (*model.Metadata, error)
	// Count count the immutable rules
	Count(ctx context.Context, query *q.Query) (int64, error)
	// ListImmutableRules list the immutable rules
	ListImmutableRules(ctx context.Context, query *q.Query) ([]*model.Metadata, error)
	// DeleteImmutableRule delete the immutable rule
	DeleteImmutableRule(ctx context.Context, id int64) error
}

type defaultRuleManager struct {
	dao dao.DAO
}

func (drm *defaultRuleManager) CreateImmutableRule(ctx context.Context, ir *model.Metadata) (int64, error) {
	daoRule := &dao_model.ImmutableRule{}
	daoRule.Disabled = ir.Disabled
	daoRule.ProjectID = ir.ProjectID
	data, _ := json.Marshal(ir)
	daoRule.TagFilter = string(data)
	return drm.dao.CreateImmutableRule(ctx, daoRule)
}

func (drm *defaultRuleManager) UpdateImmutableRule(ctx context.Context, projectID int64, ir *model.Metadata) error {
	daoRule := &dao_model.ImmutableRule{}
	data, _ := json.Marshal(ir)
	daoRule.ID = ir.ID
	daoRule.TagFilter = string(data)
	return drm.dao.UpdateImmutableRule(ctx, projectID, daoRule)
}

func (drm *defaultRuleManager) EnableImmutableRule(ctx context.Context, id int64, enabled bool) error {
	return drm.dao.ToggleImmutableRule(ctx, id, enabled)
}

func (drm *defaultRuleManager) GetImmutableRule(ctx context.Context, id int64) (*model.Metadata, error) {
	daoRule, err := drm.dao.GetImmutableRule(ctx, id)
	if err != nil {
		return nil, err
	}
	rule := &model.Metadata{}
	if daoRule == nil {
		return nil, nil
	}
	if err = json.Unmarshal([]byte(daoRule.TagFilter), rule); err != nil {
		return nil, err
	}
	rule.ID = daoRule.ID
	rule.Disabled = daoRule.Disabled
	return rule, nil
}

func (drm *defaultRuleManager) ListImmutableRules(ctx context.Context, query *q.Query) ([]*model.Metadata, error) {
	daoRules, err := drm.dao.ListImmutableRules(ctx, query)
	if err != nil {
		return nil, err
	}
	rules := make([]*model.Metadata, 0)
	for _, daoRule := range daoRules {
		rule := model.Metadata{}
		if err = json.Unmarshal([]byte(daoRule.TagFilter), &rule); err != nil {
			return nil, err
		}
		rule.ID = daoRule.ID
		rule.Disabled = daoRule.Disabled
		rules = append(rules, &rule)
	}
	sort.Slice(rules, func(i, j int) bool {
		return rules[i].ID < rules[j].ID
	})
	return rules, nil
}

func (drm *defaultRuleManager) Count(ctx context.Context, query *q.Query) (int64, error) {
	return drm.dao.Count(ctx, query)
}

func (drm *defaultRuleManager) DeleteImmutableRule(ctx context.Context, id int64) error {
	return drm.dao.DeleteImmutableRule(ctx, id)
}

// NewDefaultRuleManager return a new instance of defaultRuleManager
func NewDefaultRuleManager() Manager {
	return &defaultRuleManager{
		dao: dao.New(),
	}
}
