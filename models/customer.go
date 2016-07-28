package models

type Customer struct {
	Id 		int    `orm:"column(id);auto" json:"id"`
	Name  string `orm:"column(name);size(32)" json:"name"`
	Tag  	string `orm:"column(tag);size(32)" json:"tag"`
}
