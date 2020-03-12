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

package cnab

import (
	"context"
	ps "github.com/goharbor/harbor/src/api/artifact/processor"
	"github.com/goharbor/harbor/src/api/artifact/processor/base"
	"github.com/goharbor/harbor/src/api/artifact/processor/blob"
	"github.com/goharbor/harbor/src/common/utils/log"
	"github.com/goharbor/harbor/src/pkg/artifact"
)

// const definitions
const (
	ArtifactTypeCNAB = "CNAB"
	mediaType        = "application/vnd.cnab.manifest.v1"
)

func init() {
	pc := &processor{
		blobFetcher:       blob.Fcher,
		manifestProcessor: base.NewManifestProcessor(),
	}
	pc.IndexProcessor = &base.IndexProcessor{}
	if err := ps.Register(pc, mediaType); err != nil {
		log.Errorf("failed to register processor for media type %s: %v", mediaType, err)
		return
	}
}

type processor struct {
	*base.IndexProcessor
	manifestProcessor *base.ManifestProcessor
	blobFetcher       blob.Fetcher
}

func (p *processor) AbstractMetadata(ctx context.Context, manifest []byte, art *artifact.Artifact) error {
	cfgManiDgt := ""
	// try to get the digest of the manifest that the config layer is referenced by
	for _, reference := range art.References {
		if reference.Annotations != nil &&
			reference.Annotations["io.cnab.manifest.type"] == "config" {
			cfgManiDgt = reference.ChildDigest
		}
	}
	if len(cfgManiDgt) == 0 {
		return nil
	}

	// get the manifest that the config layer is referenced by
	_, cfgMani, err := p.blobFetcher.FetchManifest(art.RepositoryName, cfgManiDgt)
	if err != nil {
		return err
	}

	// abstract the metadata from config layer
	return p.manifestProcessor.AbstractMetadata(ctx, cfgMani, art)
}

func (p *processor) GetArtifactType() string {
	return ArtifactTypeCNAB
}
