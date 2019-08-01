package models

import (
	"time"
)

// Artifact holds the details of a artifact.
type Artifact struct {
	ID           int64     `orm:"pk;auto;column(id)" json:"id"`
	PID          int64     `orm:"column(project_id)" json:"project_id"`
	Repo         string    `orm:"column(repo)" json:"repo"`
	Tag          string    `orm:"column(tag)" json:"tag"`
	Digest       string    `orm:"column(digest)" json:"digest"`
	Kind         string    `orm:"column(kind)" json:"kind"`
	PushTime     time.Time `orm:"column(push_time)" json:"push_time"`
	PullTime     time.Time `orm:"column(pull_time)" json:"pull_time"`
	CreationTime time.Time `orm:"column(creation_time);auto_now_add" json:"creation_time"`
}

// TableName ...
func (af *Artifact) TableName() string {
	return "artifact"
}

// ArtifactQuery ...
type ArtifactQuery struct {
	PID    int64
	Repo   string
	Tag    string
	Digest string
	Pagination
}
