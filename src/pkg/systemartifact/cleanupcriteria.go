package systemartifact

import (
	"context"
	"github.com/goharbor/harbor/src/jobservice/logger"
	"github.com/goharbor/harbor/src/lib/q"
	"github.com/goharbor/harbor/src/pkg/systemartifact/dao"
	"github.com/goharbor/harbor/src/pkg/systemartifact/model"
	"time"
)

var (
	DefaultCleanupWindowSeconds = 86400
)

// CleanupCriteria provides an interface that can be implemented
// by consumers of the system artifact management framework to
// provide a custom clean-up criteria. This allows producers of the
// system artifact data to control the lifespan of the generated artifact
// records and data.
// Every system data artifact produces must register a cleanup criteria.

type CleanupCriteria interface {
	List(ctx context.Context) ([]*model.SystemArtifact, error)
}

var DefaultCleanupCriteria = NewCleanupCriteria()

func NewCleanupCriteria() CleanupCriteria {
	return &defaultCleanupCriteria{daoLayer: dao.NewSystemArtifactDao()}
}

type defaultCleanupCriteria struct {
	daoLayer dao.DAO
}

func (cleanupCriteria *defaultCleanupCriteria) List(ctx context.Context) ([]*model.SystemArtifact, error) {

	currentTime := time.Now()
	duration := time.Duration(DefaultCleanupWindowSeconds) * time.Second
	timeRange := q.Range{Max: currentTime.Add(-duration).Format(time.RFC3339)}
	logger.Infof("Cleaning up system artifacts with range: %v", timeRange)
	query := q.Query{Keywords: map[string]interface{}{"create_time": &timeRange}}
	return cleanupCriteria.daoLayer.List(ctx, &query)
}
