package model

import (
	"github.com/astaxie/beego/orm"
)

func init() {
	orm.RegisterModel(&ImmutableRule{})
}

// ImmutableRule - rule which filter image tags should be immutable.
type ImmutableRule struct {
	ID        int64  `orm:"pk;auto;column(id)" json:"id,omitempty"`
	ProjectID int64  `orm:"column(project_id)" json:"project_id,omitempty"`
	TagFilter string `orm:"column(tag_filter)" json:"tag_filter,omitempty"`
	Disabled  bool   `orm:"column(disabled)" json:"disabled,omitempty"`
}

// TableName ...
func (c *ImmutableRule) TableName() string {
	return "immutable_tag_rule"
}
