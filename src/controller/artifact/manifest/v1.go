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
	"encoding/json"

	"github.com/docker/distribution/manifest/schema1"

	"github.com/goharbor/harbor/src/lib/log"
	"github.com/goharbor/harbor/src/lib/q"
	"github.com/goharbor/harbor/src/pkg/artifact"
	"github.com/goharbor/harbor/src/pkg/blob"
)

func init() {
	mediaTypes := []string{"", "application/json", schema1.MediaTypeSignedManifest}
	if err := Register(NewV1(blob.Mgr), mediaTypes...); err != nil {
		panic(err)
	}
}

// v1Abstractor abstracts artifacts enveloped by docker manifest v1.
type v1Abstractor struct {
	blobMgr blob.Manager
}

func NewV1(blobMgr blob.Manager) Abstractor {
	return &v1Abstractor{blobMgr: blobMgr}
}

func (a *v1Abstractor) Abstract(ctx context.Context, art *artifact.Artifact, content []byte) error {
	// unify the media type of v1 manifest to "schema1.MediaTypeSignedManifest"
	art.ManifestMediaType = schema1.MediaTypeSignedManifest
	// as no config layer in the docker v1 manifest, use the "schema1.MediaTypeSignedManifest"
	// as the media type of artifact
	art.MediaType = schema1.MediaTypeSignedManifest

	mf := &schema1.Manifest{}
	if err := json.Unmarshal(content, mf); err != nil {
		return err
	}

	var ol q.OrList
	for _, fsLayer := range mf.FSLayers {
		ol.Values = append(ol.Values, fsLayer.BlobSum.String())
	}

	// there is no layer size in v1 manifest, compute the artifact size from the blobs
	blobs, err := a.blobMgr.List(ctx, q.New(q.KeyWords{"digest": &ol}))
	if err != nil {
		log.G(ctx).Errorf("failed to get blobs of the artifact %s, error %v", art.Digest, err)
		return err
	}

	art.Size = int64(len(content))
	for _, b := range blobs {
		art.Size += b.Size
	}

	return nil
}
