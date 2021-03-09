package model

import (
	"github.com/go-openapi/strfmt"
	"github.com/goharbor/harbor/src/pkg/allowlist/models"
	svrmodels "github.com/goharbor/harbor/src/server/v2.0/models"
)

// CVEAllowlist model
type CVEAllowlist struct {
	*models.CVEAllowlist
}

// ToSwagger converts the model to swagger model
func (l *CVEAllowlist) ToSwagger() *svrmodels.CVEAllowlist {
	res := &svrmodels.CVEAllowlist{
		ID:           l.ID,
		Items:        []*svrmodels.CVEAllowlistItem{},
		ProjectID:    l.ProjectID,
		ExpiresAt:    l.ExpiresAt,
		CreationTime: strfmt.DateTime(l.CreationTime),
		UpdateTime:   strfmt.DateTime(l.UpdateTime),
	}
	for _, it := range l.Items {
		cveItem := &svrmodels.CVEAllowlistItem{
			CVEID: it.CVEID,
		}
		res.Items = append(res.Items, cveItem)
	}
	return res
}

// NewCVEAllowlist ...
func NewCVEAllowlist(l *models.CVEAllowlist) *CVEAllowlist {
	return &CVEAllowlist{l}
}
