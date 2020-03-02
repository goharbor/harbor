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
	ierror "github.com/goharbor/harbor/src/internal/error"
	"github.com/goharbor/harbor/src/pkg/artifact"
	"github.com/goharbor/harbor/src/testing/api/artifact/abstractor/blob"
	"github.com/stretchr/testify/suite"
	"testing"
)

var (
	manifest = `{
   "schemaVersion": 2,
   "mediaType": "application/vnd.docker.distribution.manifest.v2+json",
   "config": {
      "mediaType": "application/vnd.docker.container.image.v1+json",
      "size": 1510,
      "digest": "sha256:fce289e99eb9bca977dae136fbe2a82b6b7d4c372474c9235adc1741675f587e"
   },
   "layers": [
      {
         "mediaType": "application/vnd.docker.image.rootfs.diff.tar.gzip",
         "size": 977,
         "digest": "sha256:1b930d010525941c1d56ec53b97bd057a67ae1865eebf042686d2a2d18271ced"
      }
   ]
}`
	config = `{
  "architecture": "amd64",
  "config": {
    "Hostname": "",
    "Domainname": "",
    "User": "",
    "AttachStdin": false,
    "AttachStdout": false,
    "AttachStderr": false,
    "Tty": false,
    "OpenStdin": false,
    "StdinOnce": false,
    "Env": [
      "PATH=/usr/local/sbin:/usr/local/bin:/usr/sbin:/usr/bin:/sbin:/bin"
    ],
    "Cmd": [
      "/hello"
    ],
    "ArgsEscaped": true,
    "Image": "sha256:a6d1aaad8ca65655449a26146699fe9d61240071f6992975be7e720f1cd42440",
    "Volumes": null,
    "WorkingDir": "",
    "Entrypoint": null,
    "OnBuild": null,
    "Labels": null
  },
  "container": "8e2caa5a514bb6d8b4f2a2553e9067498d261a0fd83a96aeaaf303943dff6ff9",
  "container_config": {
    "Hostname": "8e2caa5a514b",
    "Domainname": "",
    "User": "",
    "AttachStdin": false,
    "AttachStdout": false,
    "AttachStderr": false,
    "Tty": false,
    "OpenStdin": false,
    "StdinOnce": false,
    "Env": [
      "PATH=/usr/local/sbin:/usr/local/bin:/usr/sbin:/usr/bin:/sbin:/bin"
    ],
    "Cmd": [
      "/bin/sh",
      "-c",
      "#(nop) ",
      "CMD [\"/hello\"]"
    ],
    "ArgsEscaped": true,
    "Image": "sha256:a6d1aaad8ca65655449a26146699fe9d61240071f6992975be7e720f1cd42440",
    "Volumes": null,
    "WorkingDir": "",
    "Entrypoint": null,
    "OnBuild": null,
    "Labels": {
      
    }
  },
  "created": "2019-01-01T01:29:27.650294696Z",
  "docker_version": "18.06.1-ce",
  "history": [
    {
      "created": "2019-01-01T01:29:27.416803627Z",
      "created_by": "/bin/sh -c #(nop) COPY file:f77490f70ce51da25bd21bfc30cb5e1a24b2b65eb37d4af0c327ddc24f0986a6 in / "
    },
    {
      "created": "2019-01-01T01:29:27.650294696Z",
      "created_by": "/bin/sh -c #(nop)  CMD [\"/hello\"]",
      "empty_layer": true
    }
  ],
  "os": "linux",
  "rootfs": {
    "type": "layers",
    "diff_ids": [
      "sha256:af0b15c8625bb1938f1d7b17081031f649fd14e6b233688eea3c5483994a66a3"
    ]
  }
}`
)

type manifestV2ResolverTestSuite struct {
	suite.Suite
	resolver    *manifestV2Resolver
	blobFetcher *blob.FakeFetcher
}

func (m *manifestV2ResolverTestSuite) SetupTest() {
	m.blobFetcher = &blob.FakeFetcher{}
	m.resolver = &manifestV2Resolver{
		blobFetcher: m.blobFetcher,
	}

}

func (m *manifestV2ResolverTestSuite) TestResolveMetadata() {
	artifact := &artifact.Artifact{}
	m.blobFetcher.On("FetchLayer").Return([]byte(config), nil)
	err := m.resolver.ResolveMetadata(nil, []byte(manifest), artifact)
	m.Require().Nil(err)
	m.blobFetcher.AssertExpectations(m.T())
	m.Assert().Equal("amd64", artifact.ExtraAttrs["architecture"].(string))
	m.Assert().Equal("linux", artifact.ExtraAttrs["os"].(string))
}

func (m *manifestV2ResolverTestSuite) TestResolveAddition() {
	// unknown addition
	_, err := m.resolver.ResolveAddition(nil, nil, "unknown_addition")
	m.True(ierror.IsErr(err, ierror.BadRequestCode))

	// build history
	artifact := &artifact.Artifact{}
	m.blobFetcher.On("FetchManifest").Return("", []byte(manifest), nil)
	m.blobFetcher.On("FetchLayer").Return([]byte(config), nil)
	addition, err := m.resolver.ResolveAddition(nil, artifact, AdditionTypeBuildHistory)
	m.Require().Nil(err)
	m.Equal("application/json; charset=utf-8", addition.ContentType)
	m.Equal(`[{"created":"2019-01-01T01:29:27.416803627Z","created_by":"/bin/sh -c #(nop) COPY file:f77490f70ce51da25bd21bfc30cb5e1a24b2b65eb37d4af0c327ddc24f0986a6 in / "},{"created":"2019-01-01T01:29:27.650294696Z","created_by":"/bin/sh -c #(nop)  CMD [\"/hello\"]","empty_layer":true}]`, string(addition.Content))
}

func (m *manifestV2ResolverTestSuite) TestGetArtifactType() {
	m.Assert().Equal(ArtifactTypeImage, m.resolver.GetArtifactType())
}

func (m *manifestV2ResolverTestSuite) TestListAdditionTypes() {
	additions := m.resolver.ListAdditionTypes()
	m.EqualValues([]string{AdditionTypeBuildHistory}, additions)
}

func TestManifestV2ResolverTestSuite(t *testing.T) {
	suite.Run(t, &manifestV2ResolverTestSuite{})
}
