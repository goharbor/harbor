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
	"github.com/goharbor/harbor/src/api/artifact"
	"github.com/goharbor/harbor/src/api/artifact/abstractor/resolver"
	"github.com/goharbor/harbor/src/pkg/q"
	"github.com/stretchr/testify/mock"
	"time"
)

// FakeController is a fake artifact controller that implement src/api/artifact.Controller interface
type FakeController struct {
	mock.Mock
}

// Ensure ...
func (f *FakeController) Ensure(ctx context.Context, repository, digest string, tags ...string) (bool, int64, error) {
	args := f.Called()
	return args.Bool(0), int64(args.Int(1)), args.Error(2)
}

// Count ...
func (f *FakeController) Count(ctx context.Context, query *q.Query) (int64, error) {
	args := f.Called()
	return int64(args.Int(0)), args.Error(1)
}

// List ...
func (f *FakeController) List(ctx context.Context, query *q.Query, option *artifact.Option) ([]*artifact.Artifact, error) {
	args := f.Called()
	var artifacts []*artifact.Artifact
	if args.Get(0) != nil {
		artifacts = args.Get(0).([]*artifact.Artifact)
	}
	return artifacts, args.Error(1)
}

// Get ...
func (f *FakeController) Get(ctx context.Context, id int64, option *artifact.Option) (*artifact.Artifact, error) {
	args := f.Called()
	var art *artifact.Artifact
	if args.Get(0) != nil {
		art = args.Get(0).(*artifact.Artifact)
	}
	return art, args.Error(1)
}

// GetByReference ...
func (f *FakeController) GetByReference(ctx context.Context, repository, reference string, option *artifact.Option) (*artifact.Artifact, error) {
	args := f.Called()
	var art *artifact.Artifact
	if args.Get(0) != nil {
		art = args.Get(0).(*artifact.Artifact)
	}
	return art, args.Error(1)
}

// Delete ...
func (f *FakeController) Delete(ctx context.Context, id int64) (err error) {
	args := f.Called()
	return args.Error(0)
}

// Copy ...
func (f *FakeController) Copy(ctx context.Context, srcRepo, ref, dstRepo string) (int64, error) {
	args := f.Called()
	return int64(args.Int(0)), args.Error(1)
}

// UpdatePullTime ...
func (f *FakeController) UpdatePullTime(ctx context.Context, artifactID int64, tagID int64, time time.Time) error {
	args := f.Called()
	return args.Error(0)
}

// GetAddition ...
func (f *FakeController) GetAddition(ctx context.Context, artifactID int64, addition string) (*resolver.Addition, error) {
	args := f.Called()
	var res *resolver.Addition
	if args.Get(0) != nil {
		res = args.Get(0).(*resolver.Addition)
	}
	return res, args.Error(1)
}

// AddLabel ...
func (f *FakeController) AddLabel(ctx context.Context, artifactID int64, labelID int64) error {
	args := f.Called()
	return args.Error(0)
}

// RemoveLabel ...
func (f *FakeController) RemoveLabel(ctx context.Context, artifactID int64, labelID int64) error {
	args := f.Called()
	return args.Error(0)
}
