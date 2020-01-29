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

package resolver

import (
	"context"
	"github.com/goharbor/harbor/src/api/artifact/abstractor/resolver"
	"github.com/goharbor/harbor/src/pkg/artifact"
	"github.com/stretchr/testify/mock"
)

// FakeResolver is a fake resolver that implement the src/api/artifact/abstractor/resolver.Resolver interface
type FakeResolver struct {
	mock.Mock
}

// ResolveMetadata ...
func (f *FakeResolver) ResolveMetadata(ctx context.Context, manifest []byte, artifact *artifact.Artifact) error {
	args := f.Called()
	return args.Error(0)
}

// ResolveAddition ...
func (f *FakeResolver) ResolveAddition(ctx context.Context, artifact *artifact.Artifact, additionType string) (*resolver.Addition, error) {
	args := f.Called()
	var addition *resolver.Addition
	if args.Get(0) != nil {
		addition = args.Get(0).(*resolver.Addition)
	}
	return addition, args.Error(1)
}
