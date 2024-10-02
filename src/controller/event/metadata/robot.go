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
	"fmt"
	"time"

	"github.com/goharbor/harbor/src/common/security"
	event2 "github.com/goharbor/harbor/src/controller/event"
	"github.com/goharbor/harbor/src/lib/config"
	"github.com/goharbor/harbor/src/pkg/notifier/event"
	"github.com/goharbor/harbor/src/pkg/robot/model"
)

// CreateRobotEventMetadata is the metadata from which the create robot event can be resolved
type CreateRobotEventMetadata struct {
	Ctx   context.Context
	Robot *model.Robot
}

// Resolve to the event from the metadata
func (c *CreateRobotEventMetadata) Resolve(event *event.Event) error {
	data := &event2.CreateRobotEvent{
		EventType: event2.TopicCreateRobot,
		Robot:     c.Robot,
		OccurAt:   time.Now(),
	}
	cx, exist := security.FromContext(c.Ctx)
	if exist {
		data.Operator = cx.GetUsername()
	}
	data.Robot.Name = fmt.Sprintf("%s%s", config.RobotPrefix(c.Ctx), data.Robot.Name)
	event.Topic = event2.TopicCreateRobot
	event.Data = data
	return nil
}

// DeleteRobotEventMetadata is the metadata from which the delete robot event can be resolved
type DeleteRobotEventMetadata struct {
	Ctx   context.Context
	Robot *model.Robot
}

// Resolve to the event from the metadata
func (d *DeleteRobotEventMetadata) Resolve(event *event.Event) error {
	data := &event2.DeleteRobotEvent{
		EventType: event2.TopicDeleteRobot,
		Robot:     d.Robot,
		OccurAt:   time.Now(),
	}
	cx, exist := security.FromContext(d.Ctx)
	if exist {
		data.Operator = cx.GetUsername()
	}
	data.Robot.Name = fmt.Sprintf("%s%s", config.RobotPrefix(d.Ctx), data.Robot.Name)
	event.Topic = event2.TopicDeleteRobot
	event.Data = data
	return nil
}
