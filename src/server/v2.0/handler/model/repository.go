package model

import (
	"github.com/go-openapi/strfmt"

	"github.com/goharbor/harbor/src/pkg/repository/model"
	"github.com/goharbor/harbor/src/server/v2.0/models"
)

// RepoRecord model
type RepoRecord struct {
	*model.RepoRecord
}

// ToSwagger converts the repository into the swagger model
func (r *RepoRecord) ToSwagger() *models.Repository {
	var createTime *strfmt.DateTime
	if !r.CreationTime.IsZero() {
		t := strfmt.DateTime(r.CreationTime)
		createTime = &t
	}

	return &models.Repository{
		CreationTime: createTime,
		Description:  r.Description,
		ID:           r.RepositoryID,
		Name:         r.Name,
		ProjectID:    r.ProjectID,
		PullCount:    r.PullCount,
		UpdateTime:   strfmt.DateTime(r.UpdateTime),
	}
}

// NewRepoRecord ...
func NewRepoRecord(r *model.RepoRecord) *RepoRecord {
	return &RepoRecord{RepoRecord: r}
}
