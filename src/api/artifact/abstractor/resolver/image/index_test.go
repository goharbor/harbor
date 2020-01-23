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

package image

import (
	"github.com/goharbor/harbor/src/pkg/artifact"
	htesting "github.com/goharbor/harbor/src/testing"
	"github.com/stretchr/testify/suite"
	"testing"
)

type indexResolverTestSuite struct {
	suite.Suite
	resolver *indexResolver
	artMgr   *htesting.FakeArtifactManager
}

func (i *indexResolverTestSuite) SetupTest() {
	i.artMgr = &htesting.FakeArtifactManager{}
	i.resolver = &indexResolver{
		artMgr: i.artMgr,
	}

}

func (i *indexResolverTestSuite) TestArtifactType() {
	i.Assert().Equal(ArtifactTypeImage, i.resolver.ArtifactType())
}

func (i *indexResolverTestSuite) TestResolve() {
	manifest := `{
  "manifests": [
    {
      "digest": "sha256:92c7f9c92844bbbb5d0a101b22f7c2a7949e40f8ea90c8b3bc396879d95e899a",
      "mediaType": "application/vnd.docker.distribution.manifest.v2+json",
      "platform": {
        "architecture": "amd64",
        "os": "linux"
      },
      "size": 524
    },
    {
      "digest": "sha256:e5785cb0c62cebbed4965129bae371f0589cadd6d84798fb58c2c5f9e237efd9",
      "mediaType": "application/vnd.docker.distribution.manifest.v2+json",
      "platform": {
        "architecture": "arm",
        "os": "linux",
        "variant": "v5"
      },
      "size": 525
    },
    {
      "digest": "sha256:50b8560ad574c779908da71f7ce370c0a2471c098d44d1c8f6b513c5a55eeeb1",
      "mediaType": "application/vnd.docker.distribution.manifest.v2+json",
      "platform": {
        "architecture": "arm",
        "os": "linux",
        "variant": "v7"
      },
      "size": 525
    },
    {
      "digest": "sha256:963612c5503f3f1674f315c67089dee577d8cc6afc18565e0b4183ae355fb343",
      "mediaType": "application/vnd.docker.distribution.manifest.v2+json",
      "platform": {
        "architecture": "arm64",
        "os": "linux",
        "variant": "v8"
      },
      "size": 525
    },
    {
      "digest": "sha256:85dc5fbe16214366748ebe9d7cc73bc42d61d19d61fe05f01e317d278c2287ed",
      "mediaType": "application/vnd.docker.distribution.manifest.v2+json",
      "platform": {
        "architecture": "386",
        "os": "linux"
      },
      "size": 527
    },
    {
      "digest": "sha256:8aaea2a718a29334caeaf225716284ce29dc17418edba98dbe6dafea5afcda16",
      "mediaType": "application/vnd.docker.distribution.manifest.v2+json",
      "platform": {
        "architecture": "ppc64le",
        "os": "linux"
      },
      "size": 525
    },
    {
      "digest": "sha256:577ad4331d4fac91807308da99ecc107dcc6b2254bc4c7166325fd01113bea2a",
      "mediaType": "application/vnd.docker.distribution.manifest.v2+json",
      "platform": {
        "architecture": "s390x",
        "os": "linux"
      },
      "size": 525
    },
    {
      "digest": "sha256:351e40a9ab7ca6818dfbf9c967d1dd15599438edc41189e3d4d87eeffba5b8bf",
      "mediaType": "application/vnd.docker.distribution.manifest.v2+json",
      "platform": {
        "architecture": "amd64",
        "os": "windows",
        "os.version": "10.0.17763.914"
      },
      "size": 1124
    }
  ],
  "mediaType": "application/vnd.docker.distribution.manifest.list.v2+json",
  "schemaVersion": 2
}`
	art := &artifact.Artifact{}
	i.artMgr.On("List").Return(1, []*artifact.Artifact{
		{
			ID: 1,
		},
	}, nil)
	err := i.resolver.Resolve(nil, []byte(manifest), art)
	i.Require().Nil(err)
	i.artMgr.AssertExpectations(i.T())
	i.Assert().Len(art.References, 8)
	i.Assert().Equal(int64(1), art.References[0].ChildID)
	i.Assert().Equal("amd64", art.References[0].Platform.Architecture)
}

func TestIndexResolverTestSuite(t *testing.T) {
	suite.Run(t, &indexResolverTestSuite{})
}
