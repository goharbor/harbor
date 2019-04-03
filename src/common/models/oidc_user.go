package models

import (
	"time"
)

// OIDCUser ...
type OIDCUser struct {
	ID           int64     `orm:"pk;auto;column(id)" json:"id"`
	UserID       int       `orm:"column(user_id)" json:"user_id"`
	Secret       string    `orm:"column(secret)" json:"secret"`
	SubIss       string    `orm:"column(subiss)" json:"subiss"`
	CreationTime time.Time `orm:"column(creation_time);auto_now_add" json:"creation_time"`
	UpdateTime   time.Time `orm:"column(update_time);auto_now" json:"update_time"`
}

// TableName ...
func (o *OIDCUser) TableName() string {
	return "oidc_user"
}
