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
	"github.com/goharbor/harbor/src/pkg/q"
	"time"
)

var (
	// Mgr is a global artifact manager instance
	Mgr = NewManager()
)

// Manager is the only interface of artifact module to provide the management functions for artifacts
type Manager interface {
	// List artifacts according to the query, returns all artifacts if query is nil
	List(ctx context.Context, query *q.Query) (total int64, artifacts []*Artifact, err error)
	// Get the artifact specified by the ID
	Get(ctx context.Context, id int64) (*Artifact, error)
	// Create the artifact. If the artifact is an index, make sure all the artifacts it references
	// already exist
	Create(ctx context.Context, artifact *Artifact) (id int64, err error)
	// Delete just deletes the artifact record. The underlying data of registry will be
	// removed during garbage collection
	Delete(ctx context.Context, id int64) error
	// UpdatePullTime updates the pull time of the artifact
	UpdatePullTime(ctx context.Context, artifactID int64, time time.Time) error
}

// NewManager returns an instance of the default manager
func NewManager() Manager {
	return &manager{}
}

var _ Manager = &manager{}

type manager struct {
}

func (m *manager) List(ctx context.Context, query *q.Query) (total int64, artifacts []*Artifact, err error) {
	// TODO implement
	return 0, nil, nil
}
func (m *manager) Get(ctx context.Context, id int64) (*Artifact, error) {
	// TODO implement
	return nil, nil
}
func (m *manager) Create(ctx context.Context, artifact *Artifact) (id int64, err error) {
	// TODO implement
	return 0, nil
}
func (m *manager) Delete(ctx context.Context, id int64) error {
	// TODO implement
	return nil
}
func (m *manager) UpdatePullTime(ctx context.Context, artifactID int64, time time.Time) error {
	// TODO implement
	return nil
}
