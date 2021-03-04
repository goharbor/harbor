package model

import (
	"github.com/go-openapi/strfmt"
	"github.com/goharbor/harbor/src/controller/robot"
	"github.com/goharbor/harbor/src/lib"
	"github.com/goharbor/harbor/src/server/v2.0/models"
)

// Robot ...
type Robot struct {
	*robot.Robot
}

// ToSwagger ...
func (r *Robot) ToSwagger() *models.Robot {
	perms := []*models.RobotPermission{}
	for _, p := range r.Permissions {
		temp := &models.RobotPermission{}
		lib.JSONCopy(temp, p)
		perms = append(perms, temp)
	}

	return &models.Robot{
		ID:           r.ID,
		Name:         r.Name,
		Description:  r.Description,
		ExpiresAt:    r.ExpiresAt,
		Duration:     r.Duration,
		Level:        r.Level,
		Disable:      r.Disabled,
		Editable:     r.Editable,
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
