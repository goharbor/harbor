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
	"testing"

	specs "github.com/opencontainers/image-spec/specs-go"
	v1 "github.com/opencontainers/image-spec/specs-go/v1"

	"github.com/goharbor/harbor/src/pkg/artifact"
	"github.com/goharbor/harbor/src/pkg/registry"
)

// The testify mocks record every call, so their bookkeeping would dominate the
// measurement and grow without bound across iterations.
type benchArtMgr struct {
	artifact.Manager
	art *artifact.Artifact
}

func (m *benchArtMgr) GetByDigest(_ context.Context, _, _ string) (*artifact.Artifact, error) {
	return m.art, nil
}

type benchRegCli struct {
	registry.Client
}

// The attestations resolve from their annotation, so no registry round trip is
// needed and the measurement isolates the in-memory cost of walking the children.
func benchmarkIndexAbstract(b *testing.B, platforms, attestations int) {
	content, err := json.Marshal(v1.Index{
		Versioned: specs.Versioned{SchemaVersion: 2},
		MediaType: v1.MediaTypeImageIndex,
		Manifests: syntheticManifests(platforms, attestations, true),
	})
	if err != nil {
		b.Fatal(err)
	}

	abstractor := NewIndex(&benchArtMgr{art: &artifact.Artifact{ID: 2, Size: 10}}, &benchRegCli{})
	ctx := context.Background()

	b.ReportAllocs()
	b.ResetTimer()
	for range b.N {
		art := &artifact.Artifact{ID: 1, RepositoryName: "library/test", ManifestMediaType: v1.MediaTypeImageIndex}
		if err := abstractor.Abstract(ctx, art, content); err != nil {
			b.Fatal(err)
		}
	}
}

// What a BuildKit multi-arch build actually produces.
func BenchmarkIndexAbstractRealistic(b *testing.B) { benchmarkIndexAbstract(b, 2, 2) }

// A 4MiB index body holds roughly this many descriptors, which is the shape an
// attacker controls for free.
func BenchmarkIndexAbstractWide(b *testing.B) { benchmarkIndexAbstract(b, 1, 512) }
