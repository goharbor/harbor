package models

import (
	"time"

	commonModels "github.com/goharbor/harbor/src/common/models"
	"github.com/goharbor/harbor/src/common/retention"
)

type Policy struct {
	ID      int64  `orm:"pk;auto;column(id)" json:"id"`
	Name    string `orm:"column(name)" json:"name"`
	Enabled bool   `orm:"column(enabled)" json:"enabled"`

	Scope             retention.Scope             `orm:"column(scope)" json:"scope"`
	FallThroughAction retention.FallThroughAction `orm:"column(fall_through_action)" json:"fall_through_action"`

	// The project the policy belongs to. If nil, the policy is a server-wide policy
	Project *commonModels.Project `orm:"column(project_id);null;rel(fk)" json:"project_id"`
	// The repository the policy belongs to. If nil, the policy is a project-wide policy
	Repository *commonModels.RepoRecord `orm:"column(repository_id);null;rel(fk)" json:"repository_id"`

	Filters []*FilterMetadata `orm:"reverse(many)" json:"filters"`

	CreatedAt time.Time `orm:"column(created_at);auto_now_add" json:"created_at"`
	UpdatedAt time.Time `orm:"column(updated_at);auto_now" json:"updated_at"`
}

func (p *Policy) TableName() string {
	return "retention_policy"
}

func (p *Policy) TableUnique() [][]string {
	return [][]string{
		{"Project", "Repository"},
	}
}
