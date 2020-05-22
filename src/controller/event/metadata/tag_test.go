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
	event2 "github.com/goharbor/harbor/src/controller/event"
	"github.com/goharbor/harbor/src/pkg/artifact"
	"github.com/goharbor/harbor/src/pkg/notifier/event"
	"github.com/stretchr/testify/suite"
	"testing"
)

type tagEventTestSuite struct {
	suite.Suite
}

func (t *tagEventTestSuite) TestResolveOfCreateTagEventMetadata() {
	e := &event.Event{}
	metadata := &CreateTagEventMetadata{
		Ctx:              context.Background(),
		Tag:              "latest",
		AttachedArtifact: &artifact.Artifact{ID: 1},
	}
	err := metadata.Resolve(e)
	t.Require().Nil(err)
	t.Equal(event2.TopicCreateTag, e.Topic)
	t.Require().NotNil(e.Data)
	data, ok := e.Data.(*event2.CreateTagEvent)
	t.Require().True(ok)
	t.Equal(int64(1), data.AttachedArtifact.ID)
	t.Equal("latest", data.Tag)
}

func (t *tagEventTestSuite) TestResolveOfDeleteTagEventMetadata() {
	e := &event.Event{}
	metadata := &DeleteTagEventMetadata{
		Ctx:              context.Background(),
		Tag:              "latest",
		AttachedArtifact: &artifact.Artifact{ID: 1},
	}
	err := metadata.Resolve(e)
	t.Require().Nil(err)
	t.Equal(event2.TopicDeleteTag, e.Topic)
	t.Require().NotNil(e.Data)
	data, ok := e.Data.(*event2.DeleteTagEvent)
	t.Require().True(ok)
	t.Equal(int64(1), data.AttachedArtifact.ID)
	t.Equal("latest", data.Tag)
}

func TestTagEventTestSuite(t *testing.T) {
	suite.Run(t, &tagEventTestSuite{})
}
