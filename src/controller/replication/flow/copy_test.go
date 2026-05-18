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
	"context"
	"testing"

	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"

	repctlmodel "github.com/goharbor/harbor/src/controller/replication/model"
	"github.com/goharbor/harbor/src/jobservice/job"
	"github.com/goharbor/harbor/src/pkg/reg/adapter"
	"github.com/goharbor/harbor/src/pkg/reg/model"
	"github.com/goharbor/harbor/src/pkg/task"
	testingTask "github.com/goharbor/harbor/src/testing/pkg/task"
)

type copyFlowTestSuite struct {
	suite.Suite
}

func (c *copyFlowTestSuite) TestRun() {
	adp := &mockAdapter{}
	factory := &mockFactory{}
	factory.On("AdapterPattern").Return(nil)
	factory.On("Create", mock.Anything).Return(adp, nil)
	adapter.RegisterFactory("TEST_FOR_COPY_FLOW", factory)

	adp.On("Info").Return(&model.RegistryInfo{
		SupportedResourceTypes: []string{
			model.ResourceTypeArtifact,
		},
	}, nil)
	adp.On("FetchArtifacts", mock.Anything).Return([]*model.Resource{
		{
			Type: model.ResourceTypeArtifact,
			Metadata: &model.ResourceMetadata{
				Repository: &model.Repository{
					Name: "proxy/hello-world",
				},
				Vtags: []string{"latest"},
			},
			Override: false,
		},
	}, nil)
	adp.On("PrepareForPush", mock.Anything).Return(nil)

	execMgr := &testingTask.ExecutionManager{}
	execMgr.On("Get", mock.Anything, mock.Anything).Return(&task.Execution{
		Status: job.RunningStatus.String(),
	}, nil)

	taskMgr := &testingTask.Manager{}
	taskMgr.On("Create", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(int64(1), nil).Once()
	policy := &repctlmodel.Policy{
		SrcRegistry: &model.Registry{
			Type: "TEST_FOR_COPY_FLOW",
		},
		DestRegistry: &model.Registry{
			Type: "TEST_FOR_COPY_FLOW",
		},
	}
	flow := &copyFlow{
		executionID:  1,
		policy:       policy,
		executionMgr: execMgr,
		taskMgr:      taskMgr,
	}
	err := flow.Run(context.Background())
	c.Require().Nil(err)
}

func (c *copyFlowTestSuite) TestCopyByChunkAutoEnabled() {
	// When the destination adapter supports chunked uploads,
	// copy_by_chunk should be auto-enabled even if the policy doesn't set it.
	adp := &mockAdapter{}
	factory := &mockFactory{}
	factory.On("AdapterPattern").Return(nil)
	factory.On("Create", mock.Anything).Return(adp, nil)
	adapter.RegisterFactory("TEST_FOR_COPY_FLOW_CHUNK", factory)

	adp.On("Info").Return(&model.RegistryInfo{
		SupportedResourceTypes: []string{
			model.ResourceTypeArtifact,
		},
		SupportedCopyByChunk: true,
	}, nil)
	adp.On("FetchArtifacts", mock.Anything).Return([]*model.Resource{
		{
			Type: model.ResourceTypeArtifact,
			Metadata: &model.ResourceMetadata{
				Repository: &model.Repository{
					Name: "library/test",
				},
				Vtags: []string{"v1"},
			},
			Override: false,
		},
	}, nil)
	adp.On("PrepareForPush", mock.Anything).Return(nil)

	execMgr := &testingTask.ExecutionManager{}
	execMgr.On("Get", mock.Anything, mock.Anything).Return(&task.Execution{
		Status: job.RunningStatus.String(),
	}, nil)

	taskMgr := &testingTask.Manager{}
	taskMgr.On("Create", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(int64(1), nil).Once()

	policy := &repctlmodel.Policy{
		CopyByChunk: false, // policy does NOT enable it
		SrcRegistry: &model.Registry{
			Type: "TEST_FOR_COPY_FLOW_CHUNK",
		},
		DestRegistry: &model.Registry{
			Type: "TEST_FOR_COPY_FLOW_CHUNK",
		},
	}
	flow := &copyFlow{
		executionID:  1,
		policy:       policy,
		executionMgr: execMgr,
		taskMgr:      taskMgr,
	}
	err := flow.Run(context.Background())
	c.Require().Nil(err)

	// Verify the task was created with copy_by_chunk=true
	taskMgr.AssertCalled(c.T(), "Create", mock.Anything, mock.Anything,
		mock.MatchedBy(func(j *task.Job) bool {
			return j.Parameters["copy_by_chunk"] == true
		}), mock.Anything)
}

func TestCopyFlowTestSuite(t *testing.T) {
	suite.Run(t, &copyFlowTestSuite{})
}
