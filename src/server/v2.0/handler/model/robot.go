package model

import (
	"github.com/go-openapi/strfmt"
	"github.com/goharbor/harbor/src/controller/robot"
	"github.com/goharbor/harbor/src/lib"
	"github.com/goharbor/harbor/src/server/v2.0/models"
)

type Robot struct {
	*robot.Robot
}

func (r *Robot) ToSwagger() *models.Robot {
	perms := []*models.Permission{}
	for _, p := range r.Permissions {
		temp := &models.Permission{}
		lib.JSONCopy(temp, p)
		perms = append(perms, temp)
	}

	return &models.Robot{
		ID:           r.ID,
		Name:         r.Name,
		Description:  r.Description,
		Secret:       r.Secret,
		ExpiresAt:    r.ExpiresAt,
		Level:        r.Level,
		Disable:      r.Disabled,
		CreationTime: strfmt.DateTime(r.CreationTime),
		UpdateTime:   strfmt.DateTime(r.UpdateTime),
		Permissions:  perms,
	}
}

// NewRobot ...
func NewRobot(r *robot.Robot) *Robot {
	return &Robot{
		Robot: r,
	}
}
