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
	"github.com/goharbor/harbor/src/pkg/artifact"
	"github.com/goharbor/harbor/src/pkg/q"
	"github.com/stretchr/testify/mock"
	"time"
)

// FakeManager is a fake artifact manager that implement src/pkg/artifact.Manager interface
type FakeManager struct {
	mock.Mock
}

// List ...
func (f *FakeManager) List(ctx context.Context, query *q.Query) (int64, []*artifact.Artifact, error) {
	args := f.Called()
	var artifacts []*artifact.Artifact
	if args.Get(1) != nil {
		artifacts = args.Get(1).([]*artifact.Artifact)
	}
	return int64(args.Int(0)), artifacts, args.Error(2)
}

// Get ...
func (f *FakeManager) Get(ctx context.Context, id int64) (*artifact.Artifact, error) {
	args := f.Called()
	var art *artifact.Artifact
	if args.Get(0) != nil {
		art = args.Get(0).(*artifact.Artifact)
	}
	return art, args.Error(1)
}

// GetByDigest ...
func (f *FakeManager) GetByDigest(ctx context.Context, repositoryID int64, digest string) (*artifact.Artifact, error) {
	args := f.Called()
	var art *artifact.Artifact
	if args.Get(0) != nil {
		art = args.Get(0).(*artifact.Artifact)
	}
	return art, args.Error(1)
}

// Create ...
func (f *FakeManager) Create(ctx context.Context, artifact *artifact.Artifact) (int64, error) {
	args := f.Called()
	return int64(args.Int(0)), args.Error(1)
}

// Delete ...
func (f *FakeManager) Delete(ctx context.Context, id int64) error {
	args := f.Called()
	return args.Error(0)
}

// UpdatePullTime ...
func (f *FakeManager) UpdatePullTime(ctx context.Context, artifactID int64, time time.Time) error {
	args := f.Called()
	return args.Error(0)
}

// ListReferences ...
func (f *FakeManager) ListReferences(ctx context.Context, query *q.Query) ([]*artifact.Reference, error) {
	args := f.Called()
	var references []*artifact.Reference
	if args.Get(0) != nil {
		references = args.Get(0).([]*artifact.Reference)
	}
	return references, args.Error(1)
}

// DeleteReference ...
func (f *FakeManager) DeleteReference(ctx context.Context, id int64) error {
	args := f.Called()
	return args.Error(0)
}
