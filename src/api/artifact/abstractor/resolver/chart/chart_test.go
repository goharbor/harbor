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

package chart

import (
	"github.com/goharbor/harbor/src/common/models"
	"github.com/goharbor/harbor/src/pkg/artifact"
	"github.com/goharbor/harbor/src/testing/api/artifact/abstractor/blob"
	"github.com/goharbor/harbor/src/testing/pkg/repository"
	"github.com/stretchr/testify/suite"
	"testing"
)

type resolverTestSuite struct {
	suite.Suite
	resolver    *resolver
	repoMgr     *repository.FakeManager
	blobFetcher *blob.FakeFetcher
}

func (r *resolverTestSuite) SetupTest() {
	r.repoMgr = &repository.FakeManager{}
	r.blobFetcher = &blob.FakeFetcher{}
	r.resolver = &resolver{
		repoMgr:     r.repoMgr,
		blobFetcher: r.blobFetcher,
	}

}

func (r *resolverTestSuite) TestArtifactType() {
	r.Assert().Equal(ArtifactTypeChart, r.resolver.ArtifactType())
}

func (r *resolverTestSuite) TestResolve() {
	content := `{
  "schemaVersion": 2,
  "config": {
    "mediaType": "application/vnd.cncf.helm.config.v1+json",
    "digest": "sha256:c87983b066bd08616c6135832363ed42784d66386814694b237f5608213be325",
    "size": 542
  },
  "layers": [
    {
      "mediaType": "application/vnd.cncf.helm.chart.content.layer.v1+tar",
      "digest": "sha256:0f8c0650d55f5e00d11d7462381c340454a3b9e517e15a0187011dc305690541",
      "size": 28776
    }
  ]
}`
	config := `{
  "name": "harbor",
  "home": "https://goharbor.io",
  "sources": [
    "https://github.com/goharbor/harbor",
    "https://github.com/goharbor/harbor-helm"
  ],
  "version": "1.1.2",
  "description": "An open source trusted cloud native registry that stores, signs, and scans content",
  "keywords": [
    "docker",
    "registry",
    "harbor"
  ],
  "maintainers": [
    {
      "name": "Jesse Hu",
      "email": "huh@vmware.com"
    },
    {
      "name": "paulczar",
      "email": "username.taken@gmail.com"
    }
  ],
  "icon": "https://raw.githubusercontent.com/goharbor/harbor/master/docs/img/harbor_logo.png",
  "apiVersion": "v1",
  "appVersion": "1.8.2"
}`
	artifact := &artifact.Artifact{}
	r.repoMgr.On("Get").Return(&models.RepoRecord{}, nil)
	r.blobFetcher.On("FetchLayer").Return([]byte(config), nil)
	err := r.resolver.Resolve(nil, []byte(content), artifact)
	r.Require().Nil(err)
	r.repoMgr.AssertExpectations(r.T())
	r.blobFetcher.AssertExpectations(r.T())
	r.Assert().Equal("1.1.2", artifact.ExtraAttrs["version"].(string))
	r.Assert().Equal("1.8.2", artifact.ExtraAttrs["appVersion"].(string))
}

func TestResolverTestSuite(t *testing.T) {
	suite.Run(t, &resolverTestSuite{})
}
