package models

import "github.com/astaxie/beego/orm"

func init() {
	orm.RegisterModel(
		&Policy{},
		&FilterMetadata{},
	)
}
