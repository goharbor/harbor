package models

import (
	"time"
)

// OIDCUser ...
type OIDCUser struct {
	ID     int64 `orm:"pk;auto;column(id)" json:"id"`
	UserID int   `orm:"column(user_id)" json:"user_id"`
	// encrypted secret
	Secret string `orm:"column(secret)" json:"-"`
	// secret in plain text
	PlainSecret  string    `orm:"-" json:"secret"`
	SubIss       string    `orm:"column(subiss)" json:"subiss"`
	Token        string    `orm:"column(token)" json:"-"`
	CreationTime time.Time `orm:"column(creation_time);auto_now_add" json:"creation_time"`
	UpdateTime   time.Time `orm:"column(update_time);auto_now" json:"update_time"`
}

// TableName ...
func (o *OIDCUser) TableName() string {
	return "oidc_user"
}
