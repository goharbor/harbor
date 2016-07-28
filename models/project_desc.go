package models

type ProjectDesc struct {
	ProjectId int `orm:"column(project_id)" json:"project_id"`
	Name      string `orm:"column(name)"  json:"name"`
}
