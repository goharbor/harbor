package event

import (
	"time"

	"github.com/goharbor/harbor/src/common/models"
)

// ImageEvent ...
type ImageEvent struct {
	Project  *models.Project
	Resource []*Resource
	OccurAt  time.Time
	Operator string
	RepoName string
}

// Resource include digest and tag
type Resource struct {
	Digest string
	Tag    string
}
