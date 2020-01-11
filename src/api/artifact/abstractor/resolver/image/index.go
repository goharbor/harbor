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
	"fmt"
	"github.com/docker/distribution/manifest/manifestlist"
	"github.com/goharbor/harbor/src/api/artifact/abstractor/resolver"
	"github.com/goharbor/harbor/src/common/utils/log"
	"github.com/goharbor/harbor/src/pkg/artifact"
	"github.com/goharbor/harbor/src/pkg/q"
	v1 "github.com/opencontainers/image-spec/specs-go/v1"
)

func init() {
	rslver := &indexResolver{
		artMgr: artifact.Mgr,
	}
	if err := resolver.Register(rslver, v1.MediaTypeImageIndex, manifestlist.MediaTypeManifestList); err != nil {
		log.Errorf("failed to register resolver for artifact %s: %v", rslver.ArtifactType(), err)
		return
	}
}

// indexResolver resolves artifact with OCI index and docker manifest list
type indexResolver struct {
	artMgr artifact.Manager
}

func (i *indexResolver) ArtifactType() string {
	return ArtifactTypeImage
}

func (i *indexResolver) Resolve(ctx context.Context, manifest []byte, art *artifact.Artifact) error {
	index := &v1.Index{}
	if err := json.Unmarshal(manifest, index); err != nil {
		return err
	}
	// populate the referenced artifacts
	for _, mani := range index.Manifests {
		digest := mani.Digest.String()
		_, arts, err := i.artMgr.List(ctx, &q.Query{
			Keywords: map[string]interface{}{
				"RepositoryID": art.RepositoryID,
				"Digest":       digest,
			},
		})
		if err != nil {
			return err
		}
		// make sure the child artifact exist
		if len(arts) == 0 {
			return fmt.Errorf("the referenced artifact with digest %s not found under repository %d",
				digest, art.RepositoryID)
		}
		art.References = append(art.References, &artifact.Reference{
			ChildID:  arts[0].ID,
			Platform: mani.Platform,
		})
	}
	return nil
}
