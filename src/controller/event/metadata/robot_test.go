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
	"testing"

	"github.com/stretchr/testify/suite"

	"github.com/goharbor/harbor/src/common"
	event2 "github.com/goharbor/harbor/src/controller/event"
	"github.com/goharbor/harbor/src/lib/config"
	_ "github.com/goharbor/harbor/src/pkg/config/inmemory"
	"github.com/goharbor/harbor/src/pkg/notifier/event"
	"github.com/goharbor/harbor/src/pkg/robot/model"
)

type robotEventTestSuite struct {
	suite.Suite
}

func (t *tagEventTestSuite) TestResolveOfCreateRobotEventMetadata() {
	cfg := map[string]interface{}{
		common.RobotPrefix: "robot$",
	}
	config.InitWithSettings(cfg)

	e := &event.Event{}
	metadata := &CreateRobotEventMetadata{
		Ctx: context.Background(),
		Robot: &model.Robot{
			ID:   1,
			Name: "test",
		},
	}
	err := metadata.Resolve(e)
	t.Require().Nil(err)
	t.Equal(event2.TopicCreateRobot, e.Topic)
	t.Require().NotNil(e.Data)
	data, ok := e.Data.(*event2.CreateRobotEvent)
	t.Require().True(ok)
	t.Equal(int64(1), data.Robot.ID)
	t.Equal("robot$test", data.Robot.Name)
}

func (t *tagEventTestSuite) TestResolveOfDeleteRobotEventMetadata() {
	cfg := map[string]interface{}{
		common.RobotPrefix: "robot$",
	}
	config.InitWithSettings(cfg)
	
	e := &event.Event{}
	metadata := &DeleteRobotEventMetadata{
		Ctx: context.Background(),
		Robot: &model.Robot{
			ID: 1,
		},
	}
	err := metadata.Resolve(e)
	t.Require().Nil(err)
	t.Equal(event2.TopicDeleteRobot, e.Topic)
	t.Require().NotNil(e.Data)
	data, ok := e.Data.(*event2.DeleteRobotEvent)
	t.Require().True(ok)
	t.Equal(int64(1), data.Robot.ID)
}

func TestRobotEventTestSuite(t *testing.T) {
	suite.Run(t, &robotEventTestSuite{})
}
