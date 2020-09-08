package model

import (
	"github.com/go-openapi/strfmt"
	common_models "github.com/goharbor/harbor/src/common/models"
	"github.com/goharbor/harbor/src/server/v2.0/models"
)

// RepoRecord model
type RepoRecord struct {
	*common_models.RepoRecord
}

// ToSwagger converts the repository into the swagger model
func (r *RepoRecord) ToSwagger() *models.Repository {
	return &models.Repository{
		CreationTime: strfmt.DateTime(r.CreationTime),
		Description:  r.Description,
		ID:           r.RepositoryID,
		Name:         r.Name,
		ProjectID:    r.ProjectID,
		PullCount:    r.PullCount,
		UpdateTime:   strfmt.DateTime(r.UpdateTime),
	}
}

// NewRepoRecord ...
func NewRepoRecord(r *common_models.RepoRecord) *RepoRecord {
	return &RepoRecord{RepoRecord: r}
}
