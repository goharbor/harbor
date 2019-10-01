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

package registry

import (
	"testing"

	"github.com/docker/distribution/manifest/schema2"
)

func TestUnMarshal(t *testing.T) {
	b := []byte(`{
   "schemaVersion":2,
   "mediaType":"application/vnd.docker.distribution.manifest.v2+json",
   "config":{
      "mediaType":"application/vnd.docker.container.image.v1+json",
      "size":1473,
      "digest":"sha256:c54a2cc56cbb2f04003c1cd4507e118af7c0d340fe7e2720f70976c4b75237dc"
   },
   "layers":[
      {
         "mediaType":"application/vnd.docker.image.rootfs.diff.tar.gzip",
         "size":974,
         "digest":"sha256:c04b14da8d1441880ed3fe6106fb2cc6fa1c9661846ac0266b8a5ec8edf37b7c"
      }
   ]
}`)

	manifest, _, err := UnMarshal(schema2.MediaTypeManifest, b)
	if err != nil {
		t.Fatalf("failed to parse manifest: %v", err)
	}

	refs := manifest.References()
	if len(refs) != 2 {
		t.Fatalf("unexpected length of reference: %d != %d", len(refs), 2)
	}

	digest := "sha256:c54a2cc56cbb2f04003c1cd4507e118af7c0d340fe7e2720f70976c4b75237dc"
	if refs[0].Digest.String() != digest {
		t.Errorf("unexpected digest: %s != %s", refs[0].Digest.String(), digest)
	}

	digest = "sha256:c04b14da8d1441880ed3fe6106fb2cc6fa1c9661846ac0266b8a5ec8edf37b7c"
	if refs[1].Digest.String() != digest {
		t.Errorf("unexpected digest: %s != %s", refs[1].Digest.String(), digest)
	}
}
