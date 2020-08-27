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

	ps "github.com/goharbor/harbor/src/controller/artifact/processor"
	"github.com/goharbor/harbor/src/controller/artifact/processor/base"
	"github.com/goharbor/harbor/src/lib/log"
	"github.com/goharbor/harbor/src/pkg/artifact"
)

// const definitions
const (
	ArtifactTypeCNAB = "CNAB"
	mediaType        = "application/vnd.cnab.manifest.v1"
)

func init() {
	pc := &processor{
		manifestProcessor: base.NewManifestProcessor(),
	}
	pc.IndexProcessor = base.NewIndexProcessor()
	if err := ps.Register(pc, mediaType); err != nil {
		log.Errorf("failed to register processor for media type %s: %v", mediaType, err)
		return
	}
}

type processor struct {
	*base.IndexProcessor
	manifestProcessor *base.ManifestProcessor
}

func (p *processor) AbstractMetadata(ctx context.Context, art *artifact.Artifact, manifest []byte) error {
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
	mani, _, err := p.RegCli.PullManifest(art.RepositoryName, cfgManiDgt)
	if err != nil {
		return err
	}
	_, payload, err := mani.Payload()
	if err != nil {
		return err
	}

	// abstract the metadata from config layer
	return p.manifestProcessor.AbstractMetadata(ctx, art, payload)
}

func (p *processor) GetArtifactType(ctx context.Context, artifact *artifact.Artifact) string {
	return ArtifactTypeCNAB
}
