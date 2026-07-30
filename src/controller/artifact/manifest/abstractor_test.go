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
	"testing"

	"github.com/docker/distribution/manifest/schema1"
	v1 "github.com/opencontainers/image-spec/specs-go/v1"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/goharbor/harbor/src/pkg/artifact"
	"github.com/goharbor/harbor/src/pkg/blob"
	"github.com/goharbor/harbor/src/testing/mock"
	tart "github.com/goharbor/harbor/src/testing/pkg/artifact"
	tblob "github.com/goharbor/harbor/src/testing/pkg/blob"
)

// These tests exercise each abstractor directly with injected mocks, which is
// what the exported constructors are for. Registration wires the same types up
// from the package-level managers.

const (
	v1ManifestContent = `{
   "schemaVersion": 1,
   "name": "library/hello-world",
   "tag": "latest",
   "architecture": "amd64",
   "fsLayers": [
      {"blobSum": "sha256:a3ed95caeb02ffe68cdd9fd84406680ae93d633cb16422d00e8a7c22955b46d4"},
      {"blobSum": "sha256:1b930d010525941c1d56ec53b97bd057a67ae1865eebf042686d2a2d18271ced"}
   ],
   "history": []
}`

	v2ManifestContent = `{
   "schemaVersion": 2,
   "mediaType": "application/vnd.oci.image.manifest.v1+json",
   "config": {
      "mediaType": "application/vnd.oci.image.config.v1+json",
      "size": 7023,
      "digest": "sha256:b5b2b2c507a0944348e0303114d8d93aaaa081732b86451d9bce1f432a537bc7"
   },
   "layers": [
      {
         "mediaType": "application/vnd.oci.image.layer.v1.tar+gzip",
         "size": 32654,
         "digest": "sha256:9834876dcfb05cb167a5c24953eba58c4ac89b1adf57f28f2f9d09af107ee8f0"
      }
   ],
   "annotations": {"com.example.key1": "value1"}
}`

	indexContent = `{
   "schemaVersion": 2,
   "mediaType": "application/vnd.oci.image.index.v1+json",
   "manifests": [
      {
         "mediaType": "application/vnd.oci.image.manifest.v1+json",
         "size": 7143,
         "digest": "sha256:e692418e4cbaf90ca69d05a66403747baa33ee08806650b51fab815ad7fc331f",
         "platform": {"architecture": "ppc64le", "os": "linux"}
      }
   ]
}`
)

func TestV1AbstractorUsesInjectedBlobManager(t *testing.T) {
	blobMgr := &tblob.Manager{}
	mock.OnAnything(blobMgr, "List").Return([]*blob.Blob{{Size: 10}, {Size: 20}}, nil)

	art := &artifact.Artifact{ID: 1}
	require.NoError(t, NewV1(blobMgr).Abstract(context.Background(), art, []byte(v1ManifestContent)))

	assert.Equal(t, schema1.MediaTypeSignedManifest, art.ManifestMediaType)
	assert.Equal(t, schema1.MediaTypeSignedManifest, art.MediaType)
	// there is no layer size in a v1 manifest, so it is summed from the blobs
	assert.Equal(t, int64(30+len(v1ManifestContent)), art.Size)
}

func TestV2Abstractor(t *testing.T) {
	art := &artifact.Artifact{ID: 1}
	require.NoError(t, NewV2().Abstract(context.Background(), art, []byte(v2ManifestContent)))

	assert.Equal(t, v1.MediaTypeImageConfig, art.MediaType)
	// artifactType is absent, so it falls back to the config media type
	assert.Equal(t, v1.MediaTypeImageConfig, art.ArtifactType)
	assert.Equal(t, int64(7023+32654+len(v2ManifestContent)), art.Size)
	assert.Equal(t, "value1", art.Annotations["com.example.key1"])
}

func TestIndexAbstractorUsesInjectedArtifactManager(t *testing.T) {
	artMgr := &tart.Manager{}
	mock.OnAnything(artMgr, "GetByDigest").Return(&artifact.Artifact{ID: 2, Size: 100}, nil)

	art := &artifact.Artifact{ID: 1, ManifestMediaType: v1.MediaTypeImageIndex}
	require.NoError(t, NewIndex(artMgr).Abstract(context.Background(), art, []byte(indexContent)))

	assert.Equal(t, v1.MediaTypeImageIndex, art.MediaType)
	assert.Empty(t, art.ArtifactType)
	require.Len(t, art.References, 1)
	assert.Equal(t, int64(2), art.References[0].ChildID)
	assert.Equal(t, int64(100+len(indexContent)), art.Size)
}

func TestIndexAbstractorPropagatesChildLookupError(t *testing.T) {
	artMgr := &tart.Manager{}
	mock.OnAnything(artMgr, "GetByDigest").Return(nil, assert.AnError)

	art := &artifact.Artifact{ID: 1, ManifestMediaType: v1.MediaTypeImageIndex}
	assert.Error(t, NewIndex(artMgr).Abstract(context.Background(), art, []byte(indexContent)))
}

// CNAB carries its media type in an index annotation rather than in the
// descriptor.
func TestIndexAbstractorReadsMediaTypeFromAnnotations(t *testing.T) {
	artMgr := &tart.Manager{}
	art := &artifact.Artifact{ID: 1, ManifestMediaType: v1.MediaTypeImageIndex}
	content := `{
   "schemaVersion": 2,
   "mediaType": "application/vnd.oci.image.index.v1+json",
   "manifests": [],
   "annotations": {"org.opencontainers.artifactType": "application/vnd.cnab.manifest.v1"}
}`
	require.NoError(t, NewIndex(artMgr).Abstract(context.Background(), art, []byte(content)))
	assert.Equal(t, "application/vnd.cnab.manifest.v1", art.MediaType)
}

func TestIndexAbstractorReadsArtifactType(t *testing.T) {
	artMgr := &tart.Manager{}
	art := &artifact.Artifact{ID: 1, ManifestMediaType: v1.MediaTypeImageIndex}
	content := `{
   "schemaVersion": 2,
   "mediaType": "application/vnd.oci.image.index.v1+json",
   "artifactType": "application/vnd.example.sbom",
   "manifests": []
}`
	require.NoError(t, NewIndex(artMgr).Abstract(context.Background(), art, []byte(content)))
	assert.Equal(t, "application/vnd.example.sbom", art.ArtifactType)
}

func TestV2AbstractorReadsArtifactType(t *testing.T) {
	art := &artifact.Artifact{ID: 1}
	content := `{
   "schemaVersion": 2,
   "mediaType": "application/vnd.oci.image.manifest.v1+json",
   "artifactType": "application/vnd.example.sbom",
   "config": {"mediaType": "application/vnd.oci.image.config.v1+json", "size": 1, "digest": "sha256:b5b2b2c507a0944348e0303114d8d93aaaa081732b86451d9bce1f432a537bc7"},
   "layers": []
}`
	require.NoError(t, NewV2().Abstract(context.Background(), art, []byte(content)))
	assert.Equal(t, "application/vnd.example.sbom", art.ArtifactType)
}

func TestV1AbstractorPropagatesBlobListError(t *testing.T) {
	blobMgr := &tblob.Manager{}
	mock.OnAnything(blobMgr, "List").Return(nil, assert.AnError)

	art := &artifact.Artifact{ID: 1}
	assert.Error(t, NewV1(blobMgr).Abstract(context.Background(), art, []byte(v1ManifestContent)))
}

func TestAbstractorsRejectMalformedContent(t *testing.T) {
	malformed := []byte("{not json")

	assert.Error(t, NewV1(&tblob.Manager{}).Abstract(context.Background(), &artifact.Artifact{}, malformed))
	assert.Error(t, NewV2().Abstract(context.Background(), &artifact.Artifact{}, malformed))
	assert.Error(t, NewIndex(&tart.Manager{}).Abstract(context.Background(), &artifact.Artifact{}, malformed))
}
