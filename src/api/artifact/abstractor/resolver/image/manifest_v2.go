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
	"github.com/docker/distribution/manifest/schema2"
	"github.com/goharbor/harbor/src/api/artifact/abstractor/blob"
	"github.com/goharbor/harbor/src/api/artifact/abstractor/resolver"
	"github.com/goharbor/harbor/src/common/utils/log"
	"github.com/goharbor/harbor/src/pkg/artifact"
	"github.com/goharbor/harbor/src/pkg/repository"
	"github.com/opencontainers/image-spec/specs-go/v1"
)

const (
	// ArtifactTypeImage is the artifact type for image
	ArtifactTypeImage = "IMAGE"
)

func init() {
	rslver := &manifestV2Resolver{
		repoMgr:     repository.Mgr,
		blobFetcher: blob.Fcher,
	}
	if err := resolver.Register(rslver, v1.MediaTypeImageConfig, schema2.MediaTypeImageConfig); err != nil {
		log.Errorf("failed to register resolver for artifact %s: %v", rslver.ArtifactType(), err)
		return
	}
}

// manifestV2Resolver resolve artifact with OCI manifest and docker v2 manifest
type manifestV2Resolver struct {
	repoMgr     repository.Manager
	blobFetcher blob.Fetcher
}

func (m *manifestV2Resolver) ArtifactType() string {
	return ArtifactTypeImage
}

func (m *manifestV2Resolver) Resolve(ctx context.Context, content []byte, artifact *artifact.Artifact) error {
	repository, err := m.repoMgr.Get(ctx, artifact.RepositoryID)
	if err != nil {
		return err
	}
	manifest := &v1.Manifest{}
	if err := json.Unmarshal(content, manifest); err != nil {
		return err
	}
	digest := manifest.Config.Digest.String()
	layer, err := m.blobFetcher.FetchLayer(repository.Name, digest)
	if err != nil {
		return err
	}
	image := &v1.Image{}
	if err := json.Unmarshal(layer, image); err != nil {
		return err
	}
	if artifact.ExtraAttrs == nil {
		artifact.ExtraAttrs = map[string]interface{}{}
	}
	artifact.ExtraAttrs["created"] = image.Created
	artifact.ExtraAttrs["author"] = image.Author
	artifact.ExtraAttrs["architecture"] = image.Architecture
	artifact.ExtraAttrs["os"] = image.OS
	return nil
}
