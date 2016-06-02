package models

import (
	"github.com/astaxie/beego/orm"
)

func init() {
	orm.RegisterModel(new(RepTarget),
		new(RepPolicy),
		new(RepJob),
	        new(User),
		new(Project),
		new(Role),
		new(AccessLog))
}
