package apilib

import ()

//Repository4Search ...
type Repository4Search struct {
	ProjectID     int32     `json:"project_id,omitempty"`
	ProjectName   string    `json:"project_name,omitempty"`
	ProjectPublic int32     `json:"project_public,omitempty"`
	RepoName      string    `json:"repository_name,omitempty"`
}

