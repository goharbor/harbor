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

package metadata

import (
	"context"
	event2 "github.com/goharbor/harbor/src/api/event"
	"github.com/goharbor/harbor/src/pkg/artifact"
	"github.com/goharbor/harbor/src/pkg/notifier/event"
	"github.com/stretchr/testify/suite"
	"testing"
)

type artifactEventTestSuite struct {
	suite.Suite
}

func (a *artifactEventTestSuite) TestResolveOfPushArtifactEventMetadata() {
	e := &event.Event{}
	metadata := &PushArtifactEventMetadata{
		Ctx:      context.Background(),
		Artifact: &artifact.Artifact{ID: 1},
		Tag:      "latest",
	}
	err := metadata.Resolve(e)
	a.Require().Nil(err)
	a.Equal(event2.TopicPushArtifact, e.Topic)
	a.Require().NotNil(e.Data)
	data, ok := e.Data.(*event2.PushArtifactEvent)
	a.Require().True(ok)
	a.Equal(int64(1), data.Artifact.ID)
	a.Equal("latest", data.Tags[0])
}

func (a *artifactEventTestSuite) TestResolveOfPullArtifactEventMetadata() {
	e := &event.Event{}
	metadata := &PullArtifactEventMetadata{
		Ctx:      context.Background(),
		Artifact: &artifact.Artifact{ID: 1},
		Tag:      "latest",
	}
	err := metadata.Resolve(e)
	a.Require().Nil(err)
	a.Equal(event2.TopicPullArtifact, e.Topic)
	a.Require().NotNil(e.Data)
	data, ok := e.Data.(*event2.PullArtifactEvent)
	a.Require().True(ok)
	a.Equal(int64(1), data.Artifact.ID)
	a.Equal("latest", data.Tags[0])
}

func (a *artifactEventTestSuite) TestResolveOfDeleteArtifactEventMetadata() {
	e := &event.Event{}
	metadata := &DeleteArtifactEventMetadata{
		Ctx:      context.Background(),
		Artifact: &artifact.Artifact{ID: 1},
		Tags:     []string{"latest"},
	}
	err := metadata.Resolve(e)
	a.Require().Nil(err)
	a.Equal(event2.TopicDeleteArtifact, e.Topic)
	a.Require().NotNil(e.Data)
	data, ok := e.Data.(*event2.DeleteArtifactEvent)
	a.Require().True(ok)
	a.Equal(int64(1), data.Artifact.ID)
	a.Require().Len(data.Tags, 1)
	a.Equal("latest", data.Tags[0])
}

func TestArtifactEventTestSuite(t *testing.T) {
	suite.Run(t, &artifactEventTestSuite{})
}
