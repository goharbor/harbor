package model

import (
	"github.com/astaxie/beego/orm"
	"github.com/astaxie/beego/validation"
	"github.com/goharbor/harbor/src/common/rbac"
	"github.com/goharbor/harbor/src/common/utils"
	"time"
)

// RobotTable is the name of table in DB that holds the robot object
const RobotTable = "robot"

func init() {
	orm.RegisterModel(&Robot{})
}

// Robot holds the details of a robot.
type Robot struct {
	ID           int64     `orm:"pk;auto;column(id)" json:"id"`
	Name         string    `orm:"column(name)" json:"name"`
	Token        string    `orm:"-" json:"token"`
	Description  string    `orm:"column(description)" json:"description"`
	ProjectID    int64     `orm:"column(project_id)" json:"project_id"`
	ExpiresAt    int64     `orm:"column(expiresat)" json:"expires_at"`
	Disabled     bool      `orm:"column(disabled)" json:"disabled"`
	Visible      bool      `orm:"column(visible)" json:"-"`
	CreationTime time.Time `orm:"column(creation_time);auto_now_add" json:"creation_time"`
	UpdateTime   time.Time `orm:"column(update_time);auto_now" json:"update_time"`
}

// TableName ...
func (r *Robot) TableName() string {
	return RobotTable
}

// RobotQuery ...
type RobotQuery struct {
	Name           string
	ProjectID      int64
	Disabled       bool
	FuzzyMatchName bool
	Pagination
}

// RobotCreate ...
type RobotCreate struct {
	Name        string         `json:"name"`
	ProjectID   int64          `json:"pid"`
	Description string         `json:"description"`
	Disabled    bool           `json:"disabled"`
	Visible     bool           `json:"-"`
	PolicyCheck bool           `json:"-"`
	Access      []*rbac.Policy `json:"access"`
}

// Pagination ...
type Pagination struct {
	Page int64
	Size int64
}

// Valid ...
func (rq *RobotCreate) Valid(v *validation.Validation) {
	if utils.IsIllegalLength(rq.Name, 1, 255) {
		v.SetError("name", "robot name with illegal length")
	}
	if utils.IsContainIllegalChar(rq.Name, []string{",", "~", "#", "$", "%"}) {
		v.SetError("name", "robot name contains illegal characters")
	}
}

// RobotRep ...
type RobotRep struct {
	Name  string `json:"name"`
	Token string `json:"token"`
}
