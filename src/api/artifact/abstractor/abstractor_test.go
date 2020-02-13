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

package abstractor

import (
	"github.com/docker/distribution/manifest/schema1"
	"github.com/docker/distribution/manifest/schema2"
	"github.com/goharbor/harbor/src/api/artifact/abstractor/resolver"
	"github.com/goharbor/harbor/src/common/models"
	ierror "github.com/goharbor/harbor/src/internal/error"
	"github.com/goharbor/harbor/src/pkg/artifact"
	"github.com/goharbor/harbor/src/testing/api/artifact/abstractor/blob"
	tresolver "github.com/goharbor/harbor/src/testing/api/artifact/abstractor/resolver"
	"github.com/goharbor/harbor/src/testing/pkg/repository"
	v1 "github.com/opencontainers/image-spec/specs-go/v1"
	"github.com/stretchr/testify/suite"
	"testing"
)

var (
	fakeArtifactType = "FAKE_ARTIFACT"
	v1Manifest       = `{
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
      },
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
	v2Config = `{
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
	abstractor Abstractor
	fetcher    *blob.FakeFetcher
	repoMgr    *repository.FakeManager
	resolver   *tresolver.FakeResolver
}

func (a *abstractorTestSuite) SetupTest() {
	a.fetcher = &blob.FakeFetcher{}
	a.repoMgr = &repository.FakeManager{}
	a.resolver = &tresolver.FakeResolver{}
	a.abstractor = &abstractor{
		repoMgr:     a.repoMgr,
		blobFetcher: a.fetcher,
	}
}

// docker manifest v1
func (a *abstractorTestSuite) TestAbstractMetadataOfV1Manifest() {
	resolver.Register(a.resolver, schema1.MediaTypeSignedManifest)
	a.repoMgr.On("Get").Return(&models.RepoRecord{}, nil)
	a.fetcher.On("FetchManifest").Return(schema1.MediaTypeSignedManifest, []byte(v1Manifest), nil)
	a.resolver.On("ArtifactType").Return(fakeArtifactType)
	a.resolver.On("ResolveMetadata").Return(nil)
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
	resolver.Register(a.resolver, schema2.MediaTypeImageConfig)
	a.repoMgr.On("Get").Return(&models.RepoRecord{}, nil)
	a.fetcher.On("FetchManifest").Return(schema2.MediaTypeManifest, []byte(v2Manifest), nil)
	a.resolver.On("ArtifactType").Return(fakeArtifactType)
	a.resolver.On("ResolveMetadata").Return(nil)
	artifact := &artifact.Artifact{
		ID: 1,
	}

	err := a.abstractor.AbstractMetadata(nil, artifact)
	a.Require().Nil(err)
	a.Assert().Equal(int64(1), artifact.ID)
	a.Assert().Equal(schema2.MediaTypeManifest, artifact.ManifestMediaType)
	a.Assert().Equal(schema2.MediaTypeImageConfig, artifact.MediaType)
	a.Assert().Equal(int64(3043), artifact.Size)
}

// OCI index
func (a *abstractorTestSuite) TestAbstractMetadataOfIndex() {
	resolver.Register(a.resolver, v1.MediaTypeImageIndex)
	a.repoMgr.On("Get").Return(&models.RepoRecord{}, nil)
	a.fetcher.On("FetchManifest").Return(v1.MediaTypeImageIndex, []byte(index), nil)
	a.resolver.On("ArtifactType").Return(fakeArtifactType)
	a.resolver.On("ResolveMetadata").Return(nil)
	artifact := &artifact.Artifact{
		ID: 1,
	}
	err := a.abstractor.AbstractMetadata(nil, artifact)
	a.Require().Nil(err)
	a.Assert().Equal(int64(1), artifact.ID)
	a.Assert().Equal(v1.MediaTypeImageIndex, artifact.ManifestMediaType)
	a.Assert().Equal(v1.MediaTypeImageIndex, artifact.MediaType)
	a.Assert().Equal(int64(0), artifact.Size)
	a.Assert().Equal("value1", artifact.Annotations["com.example.key1"])
}

// OCI index
func (a *abstractorTestSuite) TestAbstractMetadataOfUnsupported() {
	a.repoMgr.On("Get").Return(&models.RepoRecord{}, nil)
	a.fetcher.On("FetchManifest").Return("unsupported-manifest", []byte{}, nil)
	artifact := &artifact.Artifact{
		ID: 1,
	}
	err := a.abstractor.AbstractMetadata(nil, artifact)
	a.Require().NotNil(err)
}

func (a *abstractorTestSuite) TestParseArtifactType() {
	mediaType := ""
	typee := parseArtifactType(mediaType)
	a.Equal(ArtifactTypeUnknown, typee)

	mediaType = "unknown"
	typee = parseArtifactType(mediaType)
	a.Equal(ArtifactTypeUnknown, typee)

	mediaType = "application/vnd.oci.image.config.v1+json"
	typee = parseArtifactType(mediaType)
	a.Equal("IMAGE", typee)

	mediaType = "application/vnd.cncf.helm.chart.config.v1+json"
	typee = parseArtifactType(mediaType)
	a.Equal("HELM.CHART", typee)

	mediaType = "application/vnd.sylabs.sif.config.v1+json"
	typee = parseArtifactType(mediaType)
	a.Equal("SIF", typee)
}

func (a *abstractorTestSuite) TestAbstractAddition() {
	resolver.Register(a.resolver, v1.MediaTypeImageConfig)
	// cannot get the resolver
	art := &artifact.Artifact{
		MediaType: "unknown",
	}
	_, err := a.abstractor.AbstractAddition(nil, art, "addition")
	a.True(ierror.IsErr(err, ierror.BadRequestCode))

	// get the resolver
	art = &artifact.Artifact{
		MediaType: v1.MediaTypeImageConfig,
	}
	a.resolver.On("ResolveAddition").Return(nil, nil)
	_, err = a.abstractor.AbstractAddition(nil, art, "addition")
	a.Require().Nil(err)
}

func TestAbstractorTestSuite(t *testing.T) {
	suite.Run(t, &abstractorTestSuite{})
}
