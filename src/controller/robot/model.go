package robot

import (
	"github.com/goharbor/harbor/src/pkg/permission/types"
	"github.com/goharbor/harbor/src/pkg/robot/model"
)

const (
	// LEVELSYSTEM ...
	LEVELSYSTEM = "system"
	// LEVELPROJECT ...
	LEVELPROJECT = "project"

	// SCOPESYSTEM ...
	SCOPESYSTEM = "/system"
	// SCOPEPROJECT ...
	SCOPEPROJECT = "/project"
	// SCOPEALLPROJECT ...
	SCOPEALLPROJECT = "/project/*"

	// ROBOTTYPE ...
	ROBOTTYPE = "robotaccount"
)

// Robot ...
type Robot struct {
	model.Robot
	ProjectName string
	Level       string
	Editable    bool          `json:"editable"`
	Permissions []*Permission `json:"permissions"`
}

// IsSysLevel, true is a system level robot, others are project level.
func (r *Robot) IsSysLevel() bool {
	return r.Level == LEVELSYSTEM
}

// setLevel, 0 is a system level robot, others are project level.
func (r *Robot) setLevel() {
	if r.ProjectID == 0 {
		r.Level = LEVELSYSTEM
	} else {
		r.Level = LEVELPROJECT
	}
}

// setEditable, no secret and no permissions should be a old format robot, and it's not editable.
func (r *Robot) setEditable() {
	if r.Secret == "" && len(r.Permissions) == 0 {
		return
	}
	r.Editable = true
}

// Permission ...
type Permission struct {
	Kind      string          `json:"kind"`
	Namespace string          `json:"namespace"`
	Access    []*types.Policy `json:"access"`
	Scope     string          `json:"-"`
}

// IsCoverAll ...
func (p *Permission) IsCoverAll() bool {
	return p.Scope == SCOPEALLPROJECT
}

// Option ...
type Option struct {
	WithPermission bool
}
