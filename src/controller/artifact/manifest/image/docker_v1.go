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
	"context"
	"encoding/json"

	"github.com/docker/distribution/manifest/schema1"

	"github.com/goharbor/harbor/src/controller/artifact/manifest"
	"github.com/goharbor/harbor/src/lib/log"
	"github.com/goharbor/harbor/src/lib/q"
	"github.com/goharbor/harbor/src/pkg/artifact"
	"github.com/goharbor/harbor/src/pkg/blob"
)

func init() {
	v1 := &v1ManifestAbstractor{
		blobMgr: blob.Mgr,
	}
	if err := manifest.Register(v1,
		"",
		"application/json",
		schema1.MediaTypeSignedManifest,
	); err != nil {
		log.Errorf("failed to register v1 manifest abstractor: %v", err)
	}
}

// v1ManifestAbstractor handles Docker V1/schema1 manifests
type v1ManifestAbstractor struct {
	blobMgr blob.Manager
}

func (a *v1ManifestAbstractor) AbstractManifestMetadata(ctx context.Context, artifact *artifact.Artifact, content []byte) error {
	// unify the media type of v1 manifest to "schema1.MediaTypeSignedManifest"
	artifact.ManifestMediaType = schema1.MediaTypeSignedManifest
	// as no config layer in the docker v1 manifest, use the "schema1.MediaTypeSignedManifest"
	// as the media type of artifact
	artifact.MediaType = schema1.MediaTypeSignedManifest

	manifest := &schema1.Manifest{}
	if err := json.Unmarshal(content, manifest); err != nil {
		return err
	}

	var ol q.OrList
	for _, fsLayer := range manifest.FSLayers {
		ol.Values = append(ol.Values, fsLayer.BlobSum.String())
	}

	// there is no layer size in v1 manifest, compute the artifact size from the blobs
	blobs, err := a.blobMgr.List(ctx, q.New(q.KeyWords{"digest": &ol}))
	if err != nil {
		log.G(ctx).Errorf("failed to get blobs of the artifact %s, error %v", artifact.Digest, err)
		return err
	}

	artifact.Size = int64(len(content))
	for _, blob := range blobs {
		artifact.Size += blob.Size
	}

	return nil
}
