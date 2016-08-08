package apilib

import ()

//Project for go test
type Project struct {
	ProjectID    int32  `json:"id,omitempty"`
	OwnerID      int32  `json:"owner_id,omitempty"`
	ProjectName  string `json:"project_name,omitempty"`
	CreationTime string `json:"creation_time,omitempty"`
	Deleted      int32  `json:"deleted,omitempty"`
	UserID       int32  `json:"user_id,omitempty"`
	OwnerName    string `json:"owner_name,omitempty"`
	Public       bool   `json:"public,omitempty"`
	Togglable    bool   `json:"togglable,omitempty"`
}
