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
	ierror "github.com/goharbor/harbor/src/internal/error"
	"github.com/goharbor/harbor/src/pkg/artifact"
	chartserver "github.com/goharbor/harbor/src/pkg/chart"
	"github.com/goharbor/harbor/src/testing/api/artifact/abstractor/blob"
	"github.com/goharbor/harbor/src/testing/pkg/chart"
	"github.com/stretchr/testify/suite"
	"k8s.io/helm/pkg/chartutil"
	"testing"
)

type resolverTestSuite struct {
	suite.Suite
	resolver    *resolver
	blobFetcher *blob.FakeFetcher
	chartOptr   *chart.FakeOpertaor
}

func (r *resolverTestSuite) SetupTest() {
	r.blobFetcher = &blob.FakeFetcher{}
	r.chartOptr = &chart.FakeOpertaor{}
	r.resolver = &resolver{
		blobFetcher:   r.blobFetcher,
		chartOperator: r.chartOptr,
	}

}

func (r *resolverTestSuite) TestResolveMetadata() {
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
	r.blobFetcher.On("FetchLayer").Return([]byte(config), nil)
	err := r.resolver.ResolveMetadata(nil, []byte(content), artifact)
	r.Require().Nil(err)
	r.blobFetcher.AssertExpectations(r.T())
	r.Assert().Equal("1.1.2", artifact.ExtraAttrs["version"].(string))
	r.Assert().Equal("1.8.2", artifact.ExtraAttrs["appVersion"].(string))
}

func (r *resolverTestSuite) TestResolveAddition() {
	// unknown addition
	_, err := r.resolver.ResolveAddition(nil, nil, "unknown_addition")
	r.True(ierror.IsErr(err, ierror.BadRequestCode))

	chartManifest := `{"schemaVersion":2,"config":{"mediaType":"application/vnd.cncf.helm.config.v1+json","digest":"sha256:76a59ebef39013bf7b57e411629b569a5175590024f31eeaaa577a0f8da9e523","size":528},"layers":[{"mediaType":"application/tar+gzip","digest":"sha256:0bd64cfb958b68c71b46597e22185a41e784dc96e04090bc7d2a480b704c3b65","size":12607}]}`

	chartYaml := `{  
   “name”:“redis”,
   “home”:“http://redis.io/",
   “sources”:[
      “https://github.com/bitnami/bitnami-docker-redis"
   
],
   “version”:“3.2.5",
   “description”:“Open source, advanced key-value store. It is often referred to as a data structure server since keys can contain strings, hashes, lists, sets and sorted sets.“,
   “keywords”:[
      “redis”,
      “keyvalue”,
      “database”
   
],
   “maintainers”:[
      {
         “name”:“bitnami-bot”,
         “email”:“containers@bitnami.com"
      
}
   
],
   “icon”:“https://bitnami.com/assets/stacks/redis/img/redis-stack-220x234.png",
   “apiVersion”:“v1”,
   “appVersion”:“4.0.9”
}`

	chartDetails := &chartserver.VersionDetails{
		Dependencies: []*chartutil.Dependency{
			{
				Name:       "harbor",
				Version:    "v1.10",
				Repository: "github.com/goharbor",
			},
		},
		Values: map[string]interface{}{
			"cluster.enable":                   true,
			"cluster.slaveCount":               1,
			"image.pullPolicy":                 "Always",
			"master.securityContext.runAsUser": 1001,
		},
		Files: map[string]string{
			"README.MD":   "This chart bootstraps a [Redis](https://github.com/bitnami/bitnami-docker-redis) deployment on a [Kubernetes](http://kubernetes.io) cluster using the [Helm](https://helm.sh) package manager.",
			"VALUES.YAML": `image:\n  ## Bitnami MongoDB registry\n  ##\n  registry: docker.io\n  ## Bitnami MongoDB image name\n  ##\n  repository: bitnami/mongodb\n  ## Bitnami MongoDB image tag\n  ## ref: https://hub.docker.com/r/bitnami/mongodb/tags/\n`,
		},
	}

	artifact := &artifact.Artifact{}
	r.blobFetcher.On("FetchManifest").Return("", []byte(chartManifest), nil)
	r.blobFetcher.On("FetchLayer").Return([]byte(chartYaml), nil)
	r.chartOptr.On("GetDetails").Return(chartDetails, nil)

	// values.yaml
	addition, err := r.resolver.ResolveAddition(nil, artifact, AdditionTypeValues)
	r.Require().Nil(err)
	r.Equal("text/plain; charset=utf-8", addition.ContentType)
	r.Equal(`image:\n  ## Bitnami MongoDB registry\n  ##\n  registry: docker.io\n  ## Bitnami MongoDB image name\n  ##\n  repository: bitnami/mongodb\n  ## Bitnami MongoDB image tag\n  ## ref: https://hub.docker.com/r/bitnami/mongodb/tags/\n`, string(addition.Content))

	// README.md
	addition, err = r.resolver.ResolveAddition(nil, artifact, AdditionTypeReadme)
	r.Require().Nil(err)
	r.Equal("text/markdown; charset=utf-8", addition.ContentType)
	r.Equal(`This chart bootstraps a [Redis](https://github.com/bitnami/bitnami-docker-redis) deployment on a [Kubernetes](http://kubernetes.io) cluster using the [Helm](https://helm.sh) package manager.`, string(addition.Content))

	// README.md
	addition, err = r.resolver.ResolveAddition(nil, artifact, AdditionTypeDependencies)
	r.Require().Nil(err)
	r.Equal("application/json; charset=utf-8", addition.ContentType)
	r.Equal(`[{"name":"harbor","version":"v1.10","repository":"github.com/goharbor"}]`, string(addition.Content))
}

func (r *resolverTestSuite) TestGetArtifactType() {
	r.Assert().Equal(ArtifactTypeChart, r.resolver.GetArtifactType())
}

func (r *resolverTestSuite) TestListAdditionTypes() {
	additions := r.resolver.ListAdditionTypes()
	r.EqualValues([]string{AdditionTypeValues, AdditionTypeReadme, AdditionTypeDependencies}, additions)
}

func TestResolverTestSuite(t *testing.T) {
	suite.Run(t, &resolverTestSuite{})
}
