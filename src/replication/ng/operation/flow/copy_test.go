// You may obtain a copy of the License at
//
//    http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package flow

import (
	"testing"

	"github.com/goharbor/harbor/src/replication/ng/model"
	"github.com/stretchr/testify/require"
)

func TestRunOfCopyFlow(t *testing.T) {
	scheduler := &fakedScheduler{}
	executionMgr := &fakedExecutionManager{}
	registryMgr := &fakedRegistryManager{}
	policy := &model.Policy{}
	flow := NewCopyFlow(executionMgr, registryMgr, scheduler, 1, policy)
	err := flow.Run(nil)
	require.Nil(t, err)
}
