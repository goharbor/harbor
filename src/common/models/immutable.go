package models

// ImmutableRule - rule which filter image tags should be immutable.
type ImmutableRule struct {
	ID        int64  `orm:"pk;auto;column(id)" json:"id,omitempty"`
	ProjectID int64  `orm:"column(project_id)" json:"project_id,omitempty"`
	TagFilter string `orm:"column(tag_filter)" json:"tag_filter,omitempty"`
	Enabled   bool   `orm:"column(enabled)" json:"enabled,omitempty"`
}

// TableName ...
func (c *ImmutableRule) TableName() string {
	return "immutable_tag_rule"
}
