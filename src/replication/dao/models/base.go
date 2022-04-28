package models

import (
	"github.com/beego/beego/orm"
)

func init() {
	orm.RegisterModel(
		new(Registry),
		new(RepPolicy),
		new(Execution),
		new(Task),
		new(ScheduleJob))
}

// Pagination ...
type Pagination struct {
	Page int64
	Size int64
}
