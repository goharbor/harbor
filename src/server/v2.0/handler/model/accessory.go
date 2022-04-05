package model

import (
	"github.com/go-openapi/strfmt"
	"github.com/goharbor/harbor/src/pkg/accessory/model"
	"github.com/goharbor/harbor/src/server/v2.0/models"
)

// Accessory model
type Accessory struct {
	model.AccessoryData
}

// ToSwagger converts the label to the swagger model
func (a *Accessory) ToSwagger() *models.Accessory {
	return &models.Accessory{
		ID:                a.ID,
		ArtifactID:        a.ArtifactID,
		SubjectArtifactID: a.SubArtifactID,
		Size:              a.Size,
		Digest:            a.Digest,
		Type:              a.Type,
		Icon:              a.Icon,
		CreationTime:      strfmt.DateTime(a.CreatTime),
	}
}

// NewAccessory ...
func NewAccessory(a model.AccessoryData) *Accessory {
	return &Accessory{AccessoryData: a}
}
