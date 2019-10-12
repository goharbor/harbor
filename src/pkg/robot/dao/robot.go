package dao

import (
	"fmt"
	"github.com/astaxie/beego/orm"
	"github.com/goharbor/harbor/src/common/dao"
	"github.com/goharbor/harbor/src/pkg/q"
	"github.com/goharbor/harbor/src/pkg/robot/model"
	"strings"
	"time"
)

// RobotAccountDao defines the interface to access the ImmutableRule data model
type RobotAccountDao interface {
	// CreateRobotAccount ...
	CreateRobotAccount(robot *model.Robot) (int64, error)

	// UpdateRobotAccount ...
	UpdateRobotAccount(robot *model.Robot) error

	// GetRobotAccount ...
	GetRobotAccount(id int64) (*model.Robot, error)

	// ListRobotAccounts ...
	ListRobotAccounts(query *q.Query) ([]*model.Robot, error)

	// DeleteRobotAccount ...
	DeleteRobotAccount(id int64) error
}

// New creates a default implementation for RobotAccountDao
func New() RobotAccountDao {
	return &robotAccountDao{}
}

type robotAccountDao struct{}

// CreateRobotAccount ...
func (r *robotAccountDao) CreateRobotAccount(robot *model.Robot) (int64, error) {
	now := time.Now()
	robot.CreationTime = now
	robot.UpdateTime = now
	id, err := dao.GetOrmer().Insert(robot)
	if err != nil {
		if strings.Contains(err.Error(), "duplicate key value violates unique constraint") {
			return 0, dao.ErrDupRows
		}
		return 0, err
	}
	return id, nil
}

// GetRobotAccount ...
func (r *robotAccountDao) GetRobotAccount(id int64) (*model.Robot, error) {
	robot := &model.Robot{
		ID: id,
	}
	if err := dao.GetOrmer().Read(robot); err != nil {
		if err == orm.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}

	return robot, nil
}

// ListRobotAccounts ...
func (r *robotAccountDao) ListRobotAccounts(query *q.Query) ([]*model.Robot, error) {
	o := dao.GetOrmer()
	qt := o.QueryTable(new(model.Robot))

	if query != nil {
		if len(query.Keywords) > 0 {
			for k, v := range query.Keywords {
				qt = qt.Filter(fmt.Sprintf("%s__icontains", k), v)
			}
		}

		if query.PageNumber > 0 && query.PageSize > 0 {
			qt = qt.Limit(query.PageSize, (query.PageNumber-1)*query.PageSize)
		}
	}

	robots := make([]*model.Robot, 0)
	_, err := qt.All(&robots)
	return robots, err
}

// UpdateRobotAccount ...
func (r *robotAccountDao) UpdateRobotAccount(robot *model.Robot) error {
	robot.UpdateTime = time.Now()
	_, err := dao.GetOrmer().Update(robot)
	return err
}

// DeleteRobotAccount ...
func (r *robotAccountDao) DeleteRobotAccount(id int64) error {
	_, err := dao.GetOrmer().QueryTable(&model.Robot{}).Filter("ID", id).Delete()
	return err
}

func getRobotQuerySetter(query *model.RobotQuery) orm.QuerySeter {
	qs := dao.GetOrmer().QueryTable(&model.Robot{})

	if query == nil {
		return qs
	}

	if len(query.Name) > 0 {
		if query.FuzzyMatchName {
			qs = qs.Filter("Name__icontains", query.Name)
		} else {
			qs = qs.Filter("Name", query.Name)
		}
	}
	if query.ProjectID != 0 {
		qs = qs.Filter("ProjectID", query.ProjectID)
	}
	return qs
}
