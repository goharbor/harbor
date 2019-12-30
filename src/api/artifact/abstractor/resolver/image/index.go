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
	"github.com/docker/distribution/manifest/manifestlist"
	"github.com/goharbor/harbor/src/api/artifact/abstractor/resolver"
	"github.com/goharbor/harbor/src/common/utils/log"
	"github.com/goharbor/harbor/src/pkg/artifact"
	v1 "github.com/opencontainers/image-spec/specs-go/v1"
)

func init() {
	rslver := &indexResolver{}
	if err := resolver.Register(rslver, v1.MediaTypeImageIndex, manifestlist.MediaTypeManifestList); err != nil {
		log.Errorf("failed to register resolver for artifact %s: %v", rslver.ArtifactType(), err)
		return
	}
}

// indexResolver resolves artifact with OCI index and docker manifest list
type indexResolver struct {
}

func (i *indexResolver) ArtifactType() string {
	return ArtifactTypeImage
}

func (i *indexResolver) Resolve(ctx context.Context, manifest []byte, artifact *artifact.Artifact) error {
	// TODO implement
	// how to make sure the artifact referenced by the index has already been saved in database
	return nil
}
