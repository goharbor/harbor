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
	"github.com/goharbor/harbor/src/pkg/artifact"
	"github.com/stretchr/testify/suite"
	"testing"
)

type fakeResolver struct{}

func (f *fakeResolver) ResolveMetadata(ctx context.Context, manifest []byte, artifact *artifact.Artifact) error {
	return nil
}
func (f *fakeResolver) ResolveAddition(ctx context.Context, artifact *artifact.Artifact, additionType string) (*Addition, error) {
	return nil, nil
}

type resolverTestSuite struct {
	suite.Suite
}

func (r *resolverTestSuite) SetupTest() {
	registry = map[string]Resolver{}
}

func (r *resolverTestSuite) TestRegister() {
	// registry a resolver
	mediaType := "fake_media_type"
	err := Register(nil, mediaType)
	r.Assert().Nil(err)

	// try to register a resolver for the existing media type
	err = Register(nil, mediaType)
	r.Assert().NotNil(err)
}

func (r *resolverTestSuite) TestGet() {
	// registry a resolver
	mediaType := "fake_media_type"
	err := Register(&fakeResolver{}, mediaType)
	r.Assert().Nil(err)

	// get the resolver
	resolver := Get(mediaType)
	r.Assert().NotNil(resolver)

	// get the not exist resolver
	resolver = Get("not_existing_media_type")
	r.Assert().Nil(resolver)
}

func TestResolverTestSuite(t *testing.T) {
	suite.Run(t, &resolverTestSuite{})
}
