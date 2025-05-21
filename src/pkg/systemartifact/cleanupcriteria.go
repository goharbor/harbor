// Copyright Project Harbor Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//    http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package systemartifact

import (
	"context"
	"time"

	"github.com/goharbor/harbor/src/jobservice/logger"
	"github.com/goharbor/harbor/src/lib/q"
	"github.com/goharbor/harbor/src/pkg/systemartifact/dao"
	"github.com/goharbor/harbor/src/pkg/systemartifact/model"
)

var (
	DefaultCleanupWindowSeconds = 86400
)

// Selector provides an interface that can be implemented
// by consumers of the system artifact management framework to
// provide a custom clean-up criteria. This allows producers of the
// system artifact data to control the lifespan of the generated artifact
// records and data.
// Every system data artifact produces must register a cleanup criteria.

type Selector interface {
	// List all system artifacts created greater than 24 hours.
	List(ctx context.Context) ([]*model.SystemArtifact, error)
	// ListWithFilters allows retrieval of system artifact records that match
	// multiple filter and sort criteria that can be specified by the clients
	ListWithFilters(ctx context.Context, query *q.Query) ([]*model.SystemArtifact, error)
}

var DefaultSelector = Default()

func Default() Selector {
	return &defaultSelector{dao: dao.NewSystemArtifactDao()}
}

// defaultSelector is a default implementation of the  Selector  which select system artifacts
// older than 24 hours for clean-up
type defaultSelector struct {
	dao dao.DAO
}

func (cleanupCriteria *defaultSelector) ListWithFilters(ctx context.Context, query *q.Query) ([]*model.SystemArtifact, error) {
	return cleanupCriteria.dao.List(ctx, query)
}

func (cleanupCriteria *defaultSelector) List(ctx context.Context) ([]*model.SystemArtifact, error) {
	currentTime := time.Now()
	duration := time.Duration(DefaultCleanupWindowSeconds) * time.Second
	timeRange := q.Range{Max: currentTime.Add(-duration).Format(time.RFC3339)}
	logger.Debugf("Cleaning up system artifacts with range: %v", timeRange)
	query := q.New(map[string]any{"create_time": &timeRange})
	return cleanupCriteria.dao.List(ctx, query)
}
