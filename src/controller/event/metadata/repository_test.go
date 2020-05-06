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
	"github.com/goharbor/harbor/src/pkg/notifier/event"
	"github.com/stretchr/testify/suite"
	"testing"
)

type repositoryEventTestSuite struct {
	suite.Suite
}

func (r *repositoryEventTestSuite) TestResolveOfDeleteRepositoryEventMetadata() {
	e := &event.Event{}
	metadata := &DeleteRepositoryEventMetadata{
		Ctx:        context.Background(),
		Repository: "library/hello-world",
	}
	err := metadata.Resolve(e)
	r.Require().Nil(err)
	r.Equal(event2.TopicDeleteRepository, e.Topic)
	r.Require().NotNil(e.Data)
	data, ok := e.Data.(*event2.DeleteRepositoryEvent)
	r.Require().True(ok)
	r.Equal("library/hello-world", data.Repository)
}

func TestRepositoryEventTestSuite(t *testing.T) {
	suite.Run(t, &repositoryEventTestSuite{})
}
