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

package manifest

import (
	"context"
	"fmt"
	"io"
	"strings"
	"testing"

	"github.com/docker/distribution"
	v1 "github.com/opencontainers/image-spec/specs-go/v1"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/goharbor/harbor/src/pkg/accessory/model"
	"github.com/goharbor/harbor/src/pkg/artifact"
	"github.com/goharbor/harbor/src/testing/mock"
	tart "github.com/goharbor/harbor/src/testing/pkg/artifact"
	tregistry "github.com/goharbor/harbor/src/testing/pkg/registry"
)

const (
	amd64Digest       = "sha256:cad250bb95ea402adf4f687cc7d6747ecf0de875e6d6117f74437893964903df"
	attestationDigest = "sha256:44401ce7f2bf39029d0d56f095374b7f344e1986c8b4970ef4f4fdb98e3f7220"
	inTotoLayerDigest = "sha256:bbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb"

	// the attestation's reference digest is not a child of the index, so the
	// subject has to come from the in-toto payload
	attestationIndex = `{
  "schemaVersion": 2,
  "mediaType": "application/vnd.oci.image.index.v1+json",
  "manifests": [
    {
      "mediaType": "application/vnd.oci.image.manifest.v1+json",
      "size": 7143,
      "digest": "sha256:cad250bb95ea402adf4f687cc7d6747ecf0de875e6d6117f74437893964903df",
      "platform": {"architecture": "amd64", "os": "linux"}
    },
    {
      "mediaType": "application/vnd.oci.image.manifest.v1+json",
      "size": 1024,
      "digest": "sha256:44401ce7f2bf39029d0d56f095374b7f344e1986c8b4970ef4f4fdb98e3f7220",
      "annotations": {
        "vnd.docker.reference.digest": "sha256:480b518ed0138eacf2d070de80cb8eb019fb0b3565e2598ed654a541c31061a0",
        "vnd.docker.reference.type": "attestation-manifest"
      },
      "platform": {"architecture": "unknown", "os": "unknown"}
    }
  ],
  "annotations": {"com.example.key1": "value1"}
}`

	// here the reference digest points straight at the platform child, so the
	// in-toto payload is never needed
	attestationIndexAnnotationOnly = `{
  "schemaVersion": 2,
  "mediaType": "application/vnd.oci.image.index.v1+json",
  "manifests": [
    {
      "mediaType": "application/vnd.oci.image.manifest.v1+json",
      "size": 7143,
      "digest": "sha256:cad250bb95ea402adf4f687cc7d6747ecf0de875e6d6117f74437893964903df",
      "platform": {"architecture": "amd64", "os": "linux"}
    },
    {
      "mediaType": "application/vnd.oci.image.manifest.v1+json",
      "size": 1024,
      "digest": "sha256:44401ce7f2bf39029d0d56f095374b7f344e1986c8b4970ef4f4fdb98e3f7220",
      "annotations": {
        "vnd.docker.reference.digest": "sha256:cad250bb95ea402adf4f687cc7d6747ecf0de875e6d6117f74437893964903df",
        "vnd.docker.reference.type": "attestation-manifest"
      },
      "platform": {"architecture": "unknown", "os": "unknown"}
    }
  ]
}`

	attestationManifestFixture = `{
  "schemaVersion": 2,
  "mediaType": "application/vnd.oci.image.manifest.v1+json",
  "config": {
    "mediaType": "application/vnd.oci.image.config.v1+json",
    "size": 167,
    "digest": "sha256:aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"
  },
  "layers": [
    {
      "mediaType": "application/vnd.in-toto+json",
      "size": 2156,
      "digest": "sha256:bbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb"
    }
  ]
}`

	attestationStatement = `{
  "_type": "https://in-toto.io/Statement/v0.1",
  "subject": [
    {
      "name": "amd64",
      "digest": {"sha256": "480b518ed0138eacf2d070de80cb8eb019fb0b3565e2598ed654a541c31061a0"}
    }
  ],
  "predicateType": "https://slsa.dev/provenance/v1"
}`
)

func attestationIndexMocks(t *testing.T) (*tart.Manager, *tregistry.Client) {
	t.Helper()
	artMgr := &tart.Manager{}
	regCli := &tregistry.Client{}
	artMgr.On("GetByDigest", mock.Anything, mock.Anything, amd64Digest).
		Return(&artifact.Artifact{ID: 2, Digest: amd64Digest, Size: 10}, nil)
	artMgr.On("GetByDigest", mock.Anything, mock.Anything, attestationDigest).
		Return(&artifact.Artifact{ID: 3, Digest: attestationDigest, Size: 3}, nil)
	return artMgr, regCli
}

// The attestation must become an accessory of the platform image rather than an
// unknown/unknown child of the index.
func TestIndexAbstractorClassifiesAttestationAsAccessory(t *testing.T) {
	artMgr, regCli := attestationIndexMocks(t)

	attestationManifest, _, err := distribution.UnmarshalManifest(v1.MediaTypeImageManifest, []byte(attestationManifestFixture))
	require.NoError(t, err)
	regCli.On("PullManifest", mock.Anything, attestationDigest).Return(attestationManifest, "", nil).Once()
	regCli.On("PullBlob", mock.Anything, inTotoLayerDigest).Return(
		int64(len(attestationStatement)),
		io.NopCloser(strings.NewReader(attestationStatement)),
		nil,
	).Once()

	art := &artifact.Artifact{ID: 1, RepositoryName: "library/test", ManifestMediaType: v1.MediaTypeImageIndex}
	require.NoError(t, NewIndex(artMgr, regCli).Abstract(context.Background(), art, []byte(attestationIndex)))

	require.Len(t, art.References, 1, "the attestation must not be listed as a child of the index")
	require.Len(t, art.AccessoryCandidates, 1)
	assert.Equal(t, int64(3), art.AccessoryCandidates[0].ArtifactID)
	assert.Equal(t, int64(2), art.AccessoryCandidates[0].SubArtifactID)
	assert.Equal(t, amd64Digest, art.AccessoryCandidates[0].SubArtifactDigest)
	assert.Equal(t, model.TypeInTotoAttestation, art.AccessoryCandidates[0].Type)
	assert.Equal(t, int64(len(attestationIndex)+13), art.Size)
}

// Resolution falls back to the reference annotation when the in-toto payload
// cannot be read.
func TestIndexAbstractorFallsBackToAnnotation(t *testing.T) {
	artMgr, regCli := attestationIndexMocks(t)
	regCli.On("PullManifest", mock.Anything, attestationDigest).Return(nil, "", fmt.Errorf("no in-toto payload")).Once()

	art := &artifact.Artifact{ID: 1, RepositoryName: "library/test", ManifestMediaType: v1.MediaTypeImageIndex}
	require.NoError(t, NewIndex(artMgr, regCli).Abstract(context.Background(), art, []byte(attestationIndexAnnotationOnly)))

	require.Len(t, art.References, 1)
	require.Len(t, art.AccessoryCandidates, 1)
	assert.Equal(t, amd64Digest, art.AccessoryCandidates[0].SubArtifactDigest)
	assert.Equal(t, model.TypeInTotoAttestation, art.AccessoryCandidates[0].Type)
}

// An attestation whose subject cannot be resolved stays an ordinary child rather
// than being dropped from the index.
func TestIndexAbstractorKeepsUnresolvableAttestationAsChild(t *testing.T) {
	artMgr, regCli := attestationIndexMocks(t)
	regCli.On("PullManifest", mock.Anything, attestationDigest).Return(nil, "", fmt.Errorf("boom")).Once()

	// no platform children for the annotation to point at
	index := `{
  "schemaVersion": 2,
  "mediaType": "application/vnd.oci.image.index.v1+json",
  "manifests": [
    {
      "mediaType": "application/vnd.oci.image.manifest.v1+json",
      "size": 1024,
      "digest": "sha256:44401ce7f2bf39029d0d56f095374b7f344e1986c8b4970ef4f4fdb98e3f7220",
      "annotations": {
        "vnd.docker.reference.digest": "sha256:480b518ed0138eacf2d070de80cb8eb019fb0b3565e2598ed654a541c31061a0",
        "vnd.docker.reference.type": "attestation-manifest"
      },
      "platform": {"architecture": "unknown", "os": "unknown"}
    }
  ]
}`
	art := &artifact.Artifact{ID: 1, RepositoryName: "library/test", ManifestMediaType: v1.MediaTypeImageIndex}
	require.NoError(t, NewIndex(artMgr, regCli).Abstract(context.Background(), art, []byte(index)))

	assert.Empty(t, art.AccessoryCandidates)
	assert.Len(t, art.References, 1)
}
