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
	"github.com/docker/distribution/manifest/schema1"
	"github.com/docker/distribution/manifest/schema2"
	"github.com/goharbor/harbor/src/api/artifact/processor"
	"github.com/goharbor/harbor/src/pkg/artifact"
	"github.com/goharbor/harbor/src/testing/api/artifact/processor/blob"
	tart "github.com/goharbor/harbor/src/testing/pkg/artifact"
	v1 "github.com/opencontainers/image-spec/specs-go/v1"
	"github.com/stretchr/testify/suite"
	"testing"
)

var (
	v1Manifest = `{
  "name": "hello-world",
  "tag": "latest",
  "architecture": "amd64",
  "fsLayers": [
    {
      "blobSum": "sha256:5f70bf18a086007016e948b04aed3b82103a36bea41755b6cddfaf10ace3c6ef"
    },
    {
      "blobSum": "sha256:5f70bf18a086007016e948b04aed3b82103a36bea41755b6cddfaf10ace3c6ef"
    },
    {
      "blobSum": "sha256:cc8567d70002e957612902a8e985ea129d831ebe04057d88fb644857caa45d11"
    },
    {
      "blobSum": "sha256:5f70bf18a086007016e948b04aed3b82103a36bea41755b6cddfaf10ace3c6ef"
    }
  ],
  "history": [
    {
      "v1Compatibility": "{\"id\":\"e45a5af57b00862e5ef5782a9925979a02ba2b12dff832fd0991335f4a11e5c5\",\"parent\":\"31cbccb51277105ba3ae35ce33c22b69c9e3f1002e76e4c736a2e8ebff9d7b5d\",\"created\":\"2014-12-31T22:57:59.178729048Z\",\"container\":\"27b45f8fb11795b52e9605b686159729b0d9ca92f76d40fb4f05a62e19c46b4f\",\"container_config\":{\"Hostname\":\"8ce6509d66e2\",\"Domainname\":\"\",\"User\":\"\",\"Memory\":0,\"MemorySwap\":0,\"CpuShares\":0,\"Cpuset\":\"\",\"AttachStdin\":false,\"AttachStdout\":false,\"AttachStderr\":false,\"PortSpecs\":null,\"ExposedPorts\":null,\"Tty\":false,\"OpenStdin\":false,\"StdinOnce\":false,\"Env\":[\"PATH=/usr/local/sbin:/usr/local/bin:/usr/sbin:/usr/bin:/sbin:/bin\"],\"Cmd\":[\"/bin/sh\",\"-c\",\"#(nop) CMD [/hello]\"],\"Image\":\"31cbccb51277105ba3ae35ce33c22b69c9e3f1002e76e4c736a2e8ebff9d7b5d\",\"Volumes\":null,\"WorkingDir\":\"\",\"Entrypoint\":null,\"NetworkDisabled\":false,\"MacAddress\":\"\",\"OnBuild\":[],\"SecurityOpt\":null,\"Labels\":null},\"docker_version\":\"1.4.1\",\"config\":{\"Hostname\":\"8ce6509d66e2\",\"Domainname\":\"\",\"User\":\"\",\"Memory\":0,\"MemorySwap\":0,\"CpuShares\":0,\"Cpuset\":\"\",\"AttachStdin\":false,\"AttachStdout\":false,\"AttachStderr\":false,\"PortSpecs\":null,\"ExposedPorts\":null,\"Tty\":false,\"OpenStdin\":false,\"StdinOnce\":false,\"Env\":[\"PATH=/usr/local/sbin:/usr/local/bin:/usr/sbin:/usr/bin:/sbin:/bin\"],\"Cmd\":[\"/hello\"],\"Image\":\"31cbccb51277105ba3ae35ce33c22b69c9e3f1002e76e4c736a2e8ebff9d7b5d\",\"Volumes\":null,\"WorkingDir\":\"\",\"Entrypoint\":null,\"NetworkDisabled\":false,\"MacAddress\":\"\",\"OnBuild\":[],\"SecurityOpt\":null,\"Labels\":null},\"architecture\":\"amd64\",\"os\":\"linux\",\"Size\":0}\n"
    },
    {
      "v1Compatibility": "{\"id\":\"e45a5af57b00862e5ef5782a9925979a02ba2b12dff832fd0991335f4a11e5c5\",\"parent\":\"31cbccb51277105ba3ae35ce33c22b69c9e3f1002e76e4c736a2e8ebff9d7b5d\",\"created\":\"2014-12-31T22:57:59.178729048Z\",\"container\":\"27b45f8fb11795b52e9605b686159729b0d9ca92f76d40fb4f05a62e19c46b4f\",\"container_config\":{\"Hostname\":\"8ce6509d66e2\",\"Domainname\":\"\",\"User\":\"\",\"Memory\":0,\"MemorySwap\":0,\"CpuShares\":0,\"Cpuset\":\"\",\"AttachStdin\":false,\"AttachStdout\":false,\"AttachStderr\":false,\"PortSpecs\":null,\"ExposedPorts\":null,\"Tty\":false,\"OpenStdin\":false,\"StdinOnce\":false,\"Env\":[\"PATH=/usr/local/sbin:/usr/local/bin:/usr/sbin:/usr/bin:/sbin:/bin\"],\"Cmd\":[\"/bin/sh\",\"-c\",\"#(nop) CMD [/hello]\"],\"Image\":\"31cbccb51277105ba3ae35ce33c22b69c9e3f1002e76e4c736a2e8ebff9d7b5d\",\"Volumes\":null,\"WorkingDir\":\"\",\"Entrypoint\":null,\"NetworkDisabled\":false,\"MacAddress\":\"\",\"OnBuild\":[],\"SecurityOpt\":null,\"Labels\":null},\"docker_version\":\"1.4.1\",\"config\":{\"Hostname\":\"8ce6509d66e2\",\"Domainname\":\"\",\"User\":\"\",\"Memory\":0,\"MemorySwap\":0,\"CpuShares\":0,\"Cpuset\":\"\",\"AttachStdin\":false,\"AttachStdout\":false,\"AttachStderr\":false,\"PortSpecs\":null,\"ExposedPorts\":null,\"Tty\":false,\"OpenStdin\":false,\"StdinOnce\":false,\"Env\":[\"PATH=/usr/local/sbin:/usr/local/bin:/usr/sbin:/usr/bin:/sbin:/bin\"],\"Cmd\":[\"/hello\"],\"Image\":\"31cbccb51277105ba3ae35ce33c22b69c9e3f1002e76e4c736a2e8ebff9d7b5d\",\"Volumes\":null,\"WorkingDir\":\"\",\"Entrypoint\":null,\"NetworkDisabled\":false,\"MacAddress\":\"\",\"OnBuild\":[],\"SecurityOpt\":null,\"Labels\":null},\"architecture\":\"amd64\",\"os\":\"linux\",\"Size\":0}\n"
    }
  ],
  "schemaVersion": 1,
  "signatures": [
    {
      "header": {
        "jwk": {
          "crv": "P-256",
          "kid": "OD6I:6DRK:JXEJ:KBM4:255X:NSAA:MUSF:E4VM:ZI6W:CUN2:L4Z6:LSF4",
          "kty": "EC",
          "x": "3gAwX48IQ5oaYQAYSxor6rYYc_6yjuLCjtQ9LUakg4A",
          "y": "t72ge6kIA1XOjqjVoEOiPPAURltJFBMGDSQvEGVB010"
        },
        "alg": "ES256"
      },
      "signature": "XREm0L8WNn27Ga_iE_vRnTxVMhhYY0Zst_FfkKopg6gWSoTOZTuW4rK0fg_IqnKkEKlbD83tD46LKEGi5aIVFg",
      "protected": "eyJmb3JtYXRMZW5ndGgiOjY2MjgsImZvcm1hdFRhaWwiOiJDbjAiLCJ0aW1lIjoiMjAxNS0wNC0wOFQxODo1Mjo1OVoifQ"
    }
  ]
}`
	v2Manifest = `{
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
  ],
  "annotations": {
    "com.example.key1": "value1"
  }
}`

	index = `{
  "schemaVersion": 2,
  "manifests": [
    {
      "mediaType": "application/vnd.oci.image.manifest.v1+json",
      "size": 7143,
      "digest": "sha256:e692418e4cbaf90ca69d05a66403747baa33ee08806650b51fab815ad7fc331f",
      "platform": {
        "architecture": "ppc64le",
        "os": "linux"
      }
    },
    {
      "mediaType": "application/vnd.oci.image.manifest.v1+json",
      "size": 7682,
      "digest": "sha256:5b0bcabd1ed22e9fb1310cf6c2dec7cdef19f0ad69efa1f392e94a4333501270",
      "platform": {
        "architecture": "amd64",
        "os": "linux"
      }
    }
  ],
  "annotations": {
    "com.example.key1": "value1"
  }
}`
)

type abstractorTestSuite struct {
	suite.Suite
	argMgr     *tart.FakeManager
	fetcher    *blob.FakeFetcher
	abstractor *abstractor
}

func (a *abstractorTestSuite) SetupTest() {
	a.fetcher = &blob.FakeFetcher{}
	a.argMgr = &tart.FakeManager{}
	a.abstractor = &abstractor{
		artMgr:      a.argMgr,
		blobFetcher: a.fetcher,
	}
	// clear all registered processors
	processor.Registry = map[string]processor.Processor{}
}

// docker manifest v1
func (a *abstractorTestSuite) TestAbstractMetadataOfV1Manifest() {
	a.fetcher.On("FetchManifest").Return(schema1.MediaTypeSignedManifest, []byte(v1Manifest), nil)
	artifact := &artifact.Artifact{
		ID: 1,
	}
	err := a.abstractor.AbstractMetadata(nil, artifact)
	a.Require().Nil(err)
	a.Assert().Equal(int64(1), artifact.ID)
	a.Assert().Equal(schema1.MediaTypeSignedManifest, artifact.ManifestMediaType)
	a.Assert().Equal(schema1.MediaTypeSignedManifest, artifact.MediaType)
	a.Assert().Equal(int64(0), artifact.Size)
}

// docker manifest v2
func (a *abstractorTestSuite) TestAbstractMetadataOfV2Manifest() {
	a.fetcher.On("FetchManifest").Return(schema2.MediaTypeManifest, []byte(v2Manifest), nil)
	artifact := &artifact.Artifact{
		ID: 1,
	}
	err := a.abstractor.AbstractMetadata(nil, artifact)
	a.Require().Nil(err)
	a.Assert().Equal(int64(1), artifact.ID)
	a.Assert().Equal(schema2.MediaTypeManifest, artifact.ManifestMediaType)
	a.Assert().Equal(schema2.MediaTypeImageConfig, artifact.MediaType)
	a.Assert().Equal(int64(3043), artifact.Size)
	a.Require().Len(artifact.Annotations, 1)
	a.Equal("value1", artifact.Annotations["com.example.key1"])
}

// OCI index
func (a *abstractorTestSuite) TestAbstractMetadataOfIndex() {
	a.fetcher.On("FetchManifest").Return(v1.MediaTypeImageIndex, []byte(index), nil)
	a.argMgr.On("GetByDigest").Return(&artifact.Artifact{
		ID:   2,
		Size: 10,
	}, nil)
	artifact := &artifact.Artifact{
		ID: 1,
	}
	err := a.abstractor.AbstractMetadata(nil, artifact)
	a.Require().Nil(err)
	a.Assert().Equal(int64(1), artifact.ID)
	a.Assert().Equal(v1.MediaTypeImageIndex, artifact.ManifestMediaType)
	a.Assert().Equal(v1.MediaTypeImageIndex, artifact.MediaType)
	a.Assert().Equal(int64(668), artifact.Size)
	a.Require().Len(artifact.Annotations, 1)
	a.Assert().Equal("value1", artifact.Annotations["com.example.key1"])
	a.Len(artifact.References, 2)
}

// OCI index
func (a *abstractorTestSuite) TestAbstractMetadataOfUnsupported() {
	a.fetcher.On("FetchManifest").Return("unsupported-manifest", []byte{}, nil)
	artifact := &artifact.Artifact{
		ID: 1,
	}
	err := a.abstractor.AbstractMetadata(nil, artifact)
	a.Require().NotNil(err)
}

func TestAbstractorTestSuite(t *testing.T) {
	suite.Run(t, &abstractorTestSuite{})
}
