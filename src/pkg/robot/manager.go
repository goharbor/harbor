package robot

import (
	"github.com/goharbor/harbor/src/lib/q"
	"github.com/goharbor/harbor/src/pkg/robot/dao"
	"github.com/goharbor/harbor/src/pkg/robot/model"
)

var (
	// Mgr is a global variable for the default robot account manager implementation
	Mgr = NewDefaultRobotAccountManager()
)

// Manager ...
type Manager interface {
	// GetRobotAccount ...
	GetRobotAccount(id int64) (*model.Robot, error)

	// CreateRobotAccount ...
	CreateRobotAccount(m *model.Robot) (int64, error)

	// DeleteRobotAccount ...
	DeleteRobotAccount(id int64) error

	// UpdateRobotAccount ...
	UpdateRobotAccount(m *model.Robot) error

	// ListRobotAccount ...
	ListRobotAccount(query *q.Query) ([]*model.Robot, error)
}

type defaultRobotManager struct {
	dao dao.RobotAccountDao
}

// NewDefaultRobotAccountManager return a new instance of defaultRobotManager
func NewDefaultRobotAccountManager() Manager {
	return &defaultRobotManager{
		dao: dao.New(),
	}
}

// GetRobotAccount ...
func (drm *defaultRobotManager) GetRobotAccount(id int64) (*model.Robot, error) {
	return drm.dao.GetRobotAccount(id)
}

// CreateRobotAccount ...
func (drm *defaultRobotManager) CreateRobotAccount(r *model.Robot) (int64, error) {
	return drm.dao.CreateRobotAccount(r)
}

// DeleteRobotAccount ...
func (drm *defaultRobotManager) DeleteRobotAccount(id int64) error {
	return drm.dao.DeleteRobotAccount(id)
}

// UpdateRobotAccount ...
func (drm *defaultRobotManager) UpdateRobotAccount(r *model.Robot) error {
	return drm.dao.UpdateRobotAccount(r)
}

// ListRobotAccount ...
func (drm *defaultRobotManager) ListRobotAccount(query *q.Query) ([]*model.Robot, error) {
	return drm.dao.ListRobotAccounts(query)
}
