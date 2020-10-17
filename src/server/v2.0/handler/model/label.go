package model

import (
	"github.com/go-openapi/strfmt"
	common_models "github.com/goharbor/harbor/src/common/models"
	"github.com/goharbor/harbor/src/server/v2.0/models"
)

// Label model
type Label struct {
	*common_models.Label
}

// ToSwagger converts the label to the swagger model
func (l *Label) ToSwagger() *models.Label {
	return &models.Label{
		Color:        l.Color,
		CreationTime: strfmt.DateTime(l.CreationTime),
		Description:  l.Description,
		ID:           l.ID,
		Name:         l.Name,
		ProjectID:    l.ProjectID,
		Scope:        l.Scope,
		UpdateTime:   strfmt.DateTime(l.UpdateTime),
	}
}

// NewLabel ...
func NewLabel(l *common_models.Label) *Label {
	return &Label{Label: l}
}
