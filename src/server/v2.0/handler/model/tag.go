package model

import (
	"github.com/go-openapi/strfmt"
	"github.com/goharbor/harbor/src/controller/tag"
	"github.com/goharbor/harbor/src/server/v2.0/models"
)

// Tag model
type Tag struct {
	*tag.Tag
}

// ToSwagger converts the tag to the swagger model
func (t *Tag) ToSwagger() *models.Tag {
	return &models.Tag{
		ArtifactID:   t.ArtifactID,
		ID:           t.ID,
		Name:         t.Name,
		PullTime:     strfmt.DateTime(t.PullTime),
		PushTime:     strfmt.DateTime(t.PushTime),
		RepositoryID: t.RepositoryID,
		Immutable:    t.Immutable,
		Signed:       t.Signed,
	}
}

// NewTag ...
func NewTag(t *tag.Tag) *Tag {
	return &Tag{Tag: t}
}
