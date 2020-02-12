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

package artifact

import (
	"context"
	"github.com/goharbor/harbor/src/pkg/artifact/dao"
	"github.com/goharbor/harbor/src/pkg/q"
	"time"
)

var (
	// Mgr is a global artifact manager instance
	Mgr = NewManager()
)

// Manager is the only interface of artifact module to provide the management functions for artifacts
type Manager interface {
	// List artifacts according to the query. The artifacts that referenced by others and
	// without tags are not returned
	List(ctx context.Context, query *q.Query) (total int64, artifacts []*Artifact, err error)
	// Get the artifact specified by the ID
	Get(ctx context.Context, id int64) (artifact *Artifact, err error)
	// GetByDigest returns the artifact specified by repository ID and digest
	GetByDigest(ctx context.Context, repositoryID int64, digest string) (artifact *Artifact, err error)
	// Create the artifact. If the artifact is an index, make sure all the artifacts it references
	// already exist
	Create(ctx context.Context, artifact *Artifact) (id int64, err error)
	// Delete just deletes the artifact record. The underlying data of registry will be
	// removed during garbage collection
	Delete(ctx context.Context, id int64) (err error)
	// UpdatePullTime updates the pull time of the artifact
	UpdatePullTime(ctx context.Context, artifactID int64, time time.Time) (err error)
}

// NewManager returns an instance of the default manager
func NewManager() Manager {
	return &manager{
		dao.New(),
	}
}

var _ Manager = &manager{}

type manager struct {
	dao dao.DAO
}

func (m *manager) List(ctx context.Context, query *q.Query) (int64, []*Artifact, error) {
	total, err := m.dao.Count(ctx, query)
	if err != nil {
		return 0, nil, err
	}
	arts, err := m.dao.List(ctx, query)
	if err != nil {
		return 0, nil, err
	}
	var artifacts []*Artifact
	for _, art := range arts {
		artifact, err := m.assemble(ctx, art)
		if err != nil {
			return 0, nil, err
		}
		artifacts = append(artifacts, artifact)
	}
	return total, artifacts, nil
}

func (m *manager) Get(ctx context.Context, id int64) (*Artifact, error) {
	art, err := m.dao.Get(ctx, id)
	if err != nil {
		return nil, err
	}
	return m.assemble(ctx, art)
}

func (m *manager) GetByDigest(ctx context.Context, repositoryID int64, digest string) (*Artifact, error) {
	art, err := m.dao.GetByDigest(ctx, repositoryID, digest)
	if err != nil {
		return nil, err
	}
	return m.assemble(ctx, art)
}

func (m *manager) Create(ctx context.Context, artifact *Artifact) (int64, error) {
	id, err := m.dao.Create(ctx, artifact.To())
	if err != nil {
		return 0, err
	}
	for _, reference := range artifact.References {
		reference.ParentID = id
		if _, err = m.dao.CreateReference(ctx, reference.To()); err != nil {
			return 0, err
		}
	}
	return id, nil
}
func (m *manager) Delete(ctx context.Context, id int64) error {
	// delete references
	if err := m.dao.DeleteReferences(ctx, id); err != nil {
		return err
	}
	// delete artifact
	return m.dao.Delete(ctx, id)
}
func (m *manager) UpdatePullTime(ctx context.Context, artifactID int64, time time.Time) error {
	return m.dao.Update(ctx, &dao.Artifact{
		ID:       artifactID,
		PullTime: time,
	}, "PullTime")
}

// assemble the artifact with references populated
func (m *manager) assemble(ctx context.Context, art *dao.Artifact) (*Artifact, error) {
	artifact := &Artifact{}
	// convert from database object
	artifact.From(art)
	// populate the references
	refs, err := m.dao.ListReferences(ctx, &q.Query{
		Keywords: map[string]interface{}{
			"parent_id": artifact.ID,
		},
	})
	if err != nil {
		return nil, err
	}
	for _, ref := range refs {
		reference := &Reference{}
		reference.From(ref)
		art, err := m.dao.Get(ctx, reference.ChildID)
		if err != nil {
			return nil, err
		}
		reference.ChildDigest = art.Digest
		artifact.References = append(artifact.References, reference)
	}
	return artifact, nil
}
