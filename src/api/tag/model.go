package tag

import (
	"github.com/go-openapi/strfmt"
	"github.com/goharbor/harbor/src/pkg/signature"
	"github.com/goharbor/harbor/src/pkg/tag/model/tag"
	"github.com/goharbor/harbor/src/server/v2.0/models"
)

// Tag is the overall view of tag
type Tag struct {
	tag.Tag
	Immutable bool `json:"immutable"`
	Signed    bool `json:"signed"`
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

// Option is used to specify the properties returned when listing/getting tags
type Option struct {
	WithImmutableStatus bool
	WithSignature       bool
	SignatureChecker    *signature.Checker
}
