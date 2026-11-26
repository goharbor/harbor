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
	taskMgr.On("Create", mock.Anything, mock.Anything, mock.MatchedBy(func(j *task.Job) bool {
		return j.Name == job.ReplicationVendorType
	}), mock.Anything).Return(int64(1), nil).Once()
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

func (c *copyFlowTestSuite) TestRunEventTriggered() {
	adp := &mockAdapter{}
	factory := &mockFactory{}
	factory.On("AdapterPattern").Return(nil)
	factory.On("Create", mock.Anything).Return(adp, nil)
	adapter.RegisterFactory("TEST_FOR_COPY_FLOW_EVENT", factory)

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
	execMgr.On("Get", mock.Anything, int64(2)).Return(&task.Execution{
		ID:      2,
		Status:  job.RunningStatus.String(),
		Trigger: task.ExecutionTriggerEvent,
	}, nil)

	taskMgr := &testingTask.Manager{}
	taskMgr.On("Create", mock.Anything, int64(2), mock.MatchedBy(func(j *task.Job) bool {
		return j.Name == job.EventReplicationVendorType
	}), mock.Anything).Return(int64(1), nil).Once()
	policy := &repctlmodel.Policy{
		SrcRegistry: &model.Registry{
			Type: "TEST_FOR_COPY_FLOW_EVENT",
		},
		DestRegistry: &model.Registry{
			Type: "TEST_FOR_COPY_FLOW_EVENT",
		},
	}
	flow := &copyFlow{
		executionID:  2,
		policy:       policy,
		executionMgr: execMgr,
		taskMgr:      taskMgr,
	}
	err := flow.Run(context.Background())
	c.Require().Nil(err)
}

func (c *copyFlowTestSuite) TestRunStopped() {
	adp := &mockAdapter{}
	factory := &mockFactory{}
	factory.On("AdapterPattern").Return(nil)
	factory.On("Create", mock.Anything).Return(adp, nil)
	adapter.RegisterFactory("TEST_FOR_COPY_FLOW_STOPPED", factory)

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

	execMgr := &testingTask.ExecutionManager{}
	execMgr.On("Get", mock.Anything, int64(3)).Return(&task.Execution{
		ID:     3,
		Status: job.StoppedStatus.String(),
	}, nil)

	policy := &repctlmodel.Policy{
		SrcRegistry: &model.Registry{
			Type: "TEST_FOR_COPY_FLOW_STOPPED",
		},
		DestRegistry: &model.Registry{
			Type: "TEST_FOR_COPY_FLOW_STOPPED",
		},
	}
	flow := &copyFlow{
		executionID:  3,
		policy:       policy,
		executionMgr: execMgr,
	}
	err := flow.Run(context.Background())
	c.Require().Nil(err)
}

func TestCopyFlowTestSuite(t *testing.T) {
	suite.Run(t, &copyFlowTestSuite{})
}
