package models

type UserToken struct {
	UserId int    `orm:"column(user_id);null"`
	Token  string `orm:"column(token);size(128)"`
	Md5Token  string `orm:"column(md5_token);size(32)"`
}
