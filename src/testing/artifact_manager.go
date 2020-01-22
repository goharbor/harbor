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

package testing

import (
	"context"
	"github.com/goharbor/harbor/src/pkg/artifact"
	"github.com/goharbor/harbor/src/pkg/q"
	"github.com/stretchr/testify/mock"
	"time"
)

// FakeArtifactManager is a fake artifact manager that implement src/pkg/artifact.Manager interface
type FakeArtifactManager struct {
	mock.Mock
}

// List ...
func (f *FakeArtifactManager) List(ctx context.Context, query *q.Query) (int64, []*artifact.Artifact, error) {
	args := f.Called()
	var artifacts []*artifact.Artifact
	if args.Get(1) != nil {
		artifacts = args.Get(1).([]*artifact.Artifact)
	}
	return int64(args.Int(0)), artifacts, args.Error(2)
}

// Get ...
func (f *FakeArtifactManager) Get(ctx context.Context, id int64) (*artifact.Artifact, error) {
	args := f.Called()
	var art *artifact.Artifact
	if args.Get(0) != nil {
		art = args.Get(0).(*artifact.Artifact)
	}
	return art, args.Error(1)
}

// Create ...
func (f *FakeArtifactManager) Create(ctx context.Context, artifact *artifact.Artifact) (int64, error) {
	args := f.Called()
	return int64(args.Int(0)), args.Error(1)
}

// Delete ...
func (f *FakeArtifactManager) Delete(ctx context.Context, id int64) error {
	args := f.Called()
	return args.Error(0)
}

// UpdatePullTime ...
func (f *FakeArtifactManager) UpdatePullTime(ctx context.Context, artifactID int64, time time.Time) error {
	args := f.Called()
	return args.Error(0)
}
