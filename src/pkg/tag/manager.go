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

package tag

import (
	"context"
	"github.com/goharbor/harbor/src/pkg/q"
	"github.com/goharbor/harbor/src/pkg/tag/dao"
	"github.com/goharbor/harbor/src/pkg/tag/model/tag"
)

var (
	// Mgr is a global instance of tag manager
	Mgr = NewManager()
)

// Manager manages the tags
type Manager interface {
	// Count returns the total count of tags according to the query.
	Count(ctx context.Context, query *q.Query) (total int64, err error)
	// List tags according to the query
	List(ctx context.Context, query *q.Query) (tags []*tag.Tag, err error)
	// Get the tag specified by ID
	Get(ctx context.Context, id int64) (tag *tag.Tag, err error)
	// Create the tag and returns the ID
	Create(ctx context.Context, tag *tag.Tag) (id int64, err error)
	// Update the tag. Only the properties specified by "props" will be updated if it is set
	Update(ctx context.Context, tag *tag.Tag, props ...string) (err error)
	// Delete the tag specified by ID
	Delete(ctx context.Context, id int64) (err error)
	// DeleteOfArtifact deletes all tags attached to the artifact
	DeleteOfArtifact(ctx context.Context, artifactID int64) (err error)
}

// NewManager creates an instance of the default tag manager
func NewManager() Manager {
	return &manager{
		dao: dao.New(),
	}
}

type manager struct {
	dao dao.DAO
}

func (m *manager) Count(ctx context.Context, query *q.Query) (int64, error) {
	return m.dao.Count(ctx, query)
}

func (m *manager) List(ctx context.Context, query *q.Query) ([]*tag.Tag, error) {
	return m.dao.List(ctx, query)
}

func (m *manager) Get(ctx context.Context, id int64) (*tag.Tag, error) {
	return m.dao.Get(ctx, id)
}

func (m *manager) Create(ctx context.Context, tag *tag.Tag) (int64, error) {
	return m.dao.Create(ctx, tag)
}

func (m *manager) Update(ctx context.Context, tag *tag.Tag, props ...string) error {
	return m.dao.Update(ctx, tag, props...)
}

func (m *manager) Delete(ctx context.Context, id int64) error {
	return m.dao.Delete(ctx, id)
}

func (m *manager) DeleteOfArtifact(ctx context.Context, artifactID int64) error {
	return m.dao.DeleteOfArtifact(ctx, artifactID)
}
