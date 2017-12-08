package models

//Member : For /api/projects/:pid/members
type Member struct {
	UserName string `json:"username"`
	Roles    []int  `json:"roles"`
}
