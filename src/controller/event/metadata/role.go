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
	"github.com/goharbor/harbor/src/pkg/role/model"
)

// CreateRoleEventMetadata is the metadata from which the create role event can be resolved
type CreateRoleEventMetadata struct {
	Ctx  context.Context
	Role *model.Role
}

// Resolve to the event from the metadata
func (c *CreateRoleEventMetadata) Resolve(event *event.Event) error {
	data := &event2.CreateRoleEvent{
		EventType: event2.TopicCreateRole,
		Role:      c.Role,
		OccurAt:   time.Now(),
	}
	cx, exist := security.FromContext(c.Ctx)
	if exist {
		data.Operator = cx.GetUsername()
	}
	data.Role.Name = fmt.Sprintf("%s%s", config.RolePrefix(c.Ctx), data.Role.Name)
	event.Topic = event2.TopicCreateRole
	event.Data = data
	return nil
}

// DeleteRoleEventMetadata is the metadata from which the delete role event can be resolved
type DeleteRoleEventMetadata struct {
	Ctx      context.Context
	Role     *model.Role
	Operator string
}

// Resolve to the event from the metadata
func (d *DeleteRoleEventMetadata) Resolve(event *event.Event) error {
	data := &event2.DeleteRoleEvent{
		EventType: event2.TopicDeleteRole,
		Role:      d.Role,
		OccurAt:   time.Now(),
	}
	if d.Operator != "" {
		data.Operator = d.Operator
	} else {
		cx, exist := security.FromContext(d.Ctx)
		if exist {
			data.Operator = cx.GetUsername()
		}
	}
	data.Role.Name = fmt.Sprintf("%s%s", config.RolePrefix(d.Ctx), data.Role.Name)
	event.Topic = event2.TopicDeleteRole
	event.Data = data
	return nil
}
