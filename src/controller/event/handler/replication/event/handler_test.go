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

package event

import (
	"context"
	"testing"

	"github.com/stretchr/testify/suite"

	"github.com/goharbor/harbor/src/controller/replication"
	repctlmodel "github.com/goharbor/harbor/src/controller/replication/model"
	"github.com/goharbor/harbor/src/pkg/reg/model"
	replicationtesting "github.com/goharbor/harbor/src/testing/controller/replication"
	"github.com/goharbor/harbor/src/testing/mock"
)

type handlerTestSuite struct {
	suite.Suite
	ctl     *replicationtesting.Controller
	origCtl replication.Controller
}

func (h *handlerTestSuite) SetupTest() {
	h.origCtl = replication.Ctl
	h.ctl = &replicationtesting.Controller{}
	replication.Ctl = h.ctl
}

func (h *handlerTestSuite) TearDownTest() {
	replication.Ctl = h.origCtl
}

func (h *handlerTestSuite) pushEvent(tags ...string) *Event {
	return &Event{
		Type: EventTypeArtifactPush,
		Resource: &model.Resource{
			Type: model.ResourceTypeArtifact,
			Metadata: &model.ResourceMetadata{
				Repository: &model.Repository{Name: "library/hello-world"},
				Artifacts: []*model.Artifact{
					{
						Type:   "IMAGE",
						Digest: "sha256:418fb88ec412e340cdbef913b8ca1bbe8f9e8dc705f9617414c1f2c8db980180",
						Tags:   tags,
					},
				},
			},
		},
	}
}

func (h *handlerTestSuite) policy(decoration, pattern string) *repctlmodel.Policy {
	return &repctlmodel.Policy{
		ID:      1,
		Enabled: true,
		Filters: []*model.Filter{
			{
				Type:       model.FilterTypeTag,
				Decoration: decoration,
				Value:      pattern,
			},
		},
	}
}

func (h *handlerTestSuite) startedResource() *model.Resource {
	for _, call := range h.ctl.Calls {
		if call.Method == "Start" {
			res, ok := call.Arguments[2].(*model.Resource)
			h.Require().True(ok)
			return res
		}
	}
	return nil
}

func (h *handlerTestSuite) TestMatchesFilterReplicatesOnlyMatchingTags() {
	h.ctl.On("ListPolicies", mock.Anything, mock.Anything).
		Return([]*repctlmodel.Policy{h.policy(model.Matches, "1.0.*")}, nil)
	h.ctl.On("Start", mock.Anything, mock.Anything, mock.Anything, mock.Anything).
		Return(int64(1), nil)

	err := Handle(context.Background(), h.pushEvent("1.0.3", "1.0.11", "latest"))
	h.Require().NoError(err)

	res := h.startedResource()
	h.Require().NotNil(res)
	h.Equal([]string{"1.0.3", "1.0.11"}, res.Metadata.Artifacts[0].Tags)
}

func (h *handlerTestSuite) TestExcludesFilterDoesNotReplicateExcludedTags() {
	h.ctl.On("ListPolicies", mock.Anything, mock.Anything).
		Return([]*repctlmodel.Policy{h.policy(model.Excludes, "*-internal")}, nil)
	h.ctl.On("Start", mock.Anything, mock.Anything, mock.Anything, mock.Anything).
		Return(int64(1), nil)

	err := Handle(context.Background(), h.pushEvent("public-v1", "v1-internal"))
	h.Require().NoError(err)

	res := h.startedResource()
	h.Require().NotNil(res)
	h.Equal([]string{"public-v1"}, res.Metadata.Artifacts[0].Tags)
}

func (h *handlerTestSuite) TestNoMatchingTagStartsNoExecution() {
	h.ctl.On("ListPolicies", mock.Anything, mock.Anything).
		Return([]*repctlmodel.Policy{h.policy(model.Matches, "2.0.*")}, nil)

	err := Handle(context.Background(), h.pushEvent("1.0.3", "1.0.11"))
	h.Require().NoError(err)

	h.ctl.AssertNotCalled(h.T(), "Start", mock.Anything, mock.Anything, mock.Anything, mock.Anything)
}

func (h *handlerTestSuite) TestNoFilterReplicatesAllTags() {
	h.ctl.On("ListPolicies", mock.Anything, mock.Anything).
		Return([]*repctlmodel.Policy{{ID: 1, Enabled: true}}, nil)
	h.ctl.On("Start", mock.Anything, mock.Anything, mock.Anything, mock.Anything).
		Return(int64(1), nil)

	err := Handle(context.Background(), h.pushEvent("1.0.3", "1.0.11", "latest"))
	h.Require().NoError(err)

	res := h.startedResource()
	h.Require().NotNil(res)
	h.Equal([]string{"1.0.3", "1.0.11", "latest"}, res.Metadata.Artifacts[0].Tags)
}

func TestHandlerTestSuite(t *testing.T) {
	suite.Run(t, &handlerTestSuite{})
}
