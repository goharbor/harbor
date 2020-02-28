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
	"time"

	"github.com/goharbor/harbor/src/pkg/artifact/dao"
	"github.com/goharbor/harbor/src/pkg/q"
)

var (
	// Mgr is a global artifact manager instance
	Mgr = NewManager()
)

// Manager is the only interface of artifact module to provide the management functions for artifacts
type Manager interface {
	// Count returns the total count of artifacts according to the query.
	// The artifacts that referenced by others and without tags are not counted
	Count(ctx context.Context, query *q.Query) (total int64, err error)
	// List artifacts according to the query. The artifacts that referenced by others and
	// without tags are not returned
	List(ctx context.Context, query *q.Query) (artifacts []*Artifact, err error)
	// Get the artifact specified by the ID
	Get(ctx context.Context, id int64) (artifact *Artifact, err error)
	// GetByDigest returns the artifact specified by repository and digest
	GetByDigest(ctx context.Context, repository, digest string) (artifact *Artifact, err error)
	// Create the artifact. If the artifact is an index, make sure all the artifacts it references
	// already exist
	Create(ctx context.Context, artifact *Artifact) (id int64, err error)
	// Delete just deletes the artifact record. The underlying data of registry will be
	// removed during garbage collection
	Delete(ctx context.Context, id int64) (err error)
	// UpdatePullTime updates the pull time of the artifact
	UpdatePullTime(ctx context.Context, artifactID int64, time time.Time) (err error)
	// ListReferences according to the query
	ListReferences(ctx context.Context, query *q.Query) (references []*Reference, err error)
	// DeleteReference specified by ID
	DeleteReference(ctx context.Context, id int64) (err error)
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

func (m *manager) Count(ctx context.Context, query *q.Query) (int64, error) {
	return m.dao.Count(ctx, query)
}

func (m *manager) List(ctx context.Context, query *q.Query) ([]*Artifact, error) {
	arts, err := m.dao.List(ctx, query)
	if err != nil {
		return nil, err
	}
	var artifacts []*Artifact
	for _, art := range arts {
		artifact, err := m.assemble(ctx, art)
		if err != nil {
			return nil, err
		}
		artifacts = append(artifacts, artifact)
	}
	return artifacts, nil
}

func (m *manager) Get(ctx context.Context, id int64) (*Artifact, error) {
	art, err := m.dao.Get(ctx, id)
	if err != nil {
		return nil, err
	}
	return m.assemble(ctx, art)
}

func (m *manager) GetByDigest(ctx context.Context, repository, digest string) (*Artifact, error) {
	art, err := m.dao.GetByDigest(ctx, repository, digest)
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

func (m *manager) ListReferences(ctx context.Context, query *q.Query) ([]*Reference, error) {
	references, err := m.dao.ListReferences(ctx, query)
	if err != nil {
		return nil, err
	}
	var refs []*Reference
	for _, reference := range references {
		ref := &Reference{}
		ref.From(reference)
		art, err := m.dao.Get(ctx, reference.ChildID)
		if err != nil {
			return nil, err
		}
		ref.ChildDigest = art.Digest
		refs = append(refs, ref)
	}
	return refs, nil
}

func (m *manager) DeleteReference(ctx context.Context, id int64) error {
	return m.dao.DeleteReference(ctx, id)
}

// assemble the artifact with references populated
func (m *manager) assemble(ctx context.Context, art *dao.Artifact) (*Artifact, error) {
	artifact := &Artifact{}
	// convert from database object
	artifact.From(art)

	// populate the references
	if artifact.HasChildren() {
		references, err := m.ListReferences(ctx, q.New(q.KeyWords{"ParentID": artifact.ID}))
		if err != nil {
			return nil, err
		}
		artifact.References = references
	}

	return artifact, nil
}
