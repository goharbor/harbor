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
	"github.com/goharbor/harbor/src/common/security"
	event2 "github.com/goharbor/harbor/src/controller/event"
	"github.com/goharbor/harbor/src/pkg/artifact"
	"github.com/goharbor/harbor/src/pkg/notifier/event"
	"time"
)

// CreateTagEventMetadata is the metadata from which the create tag event can be resolved
type CreateTagEventMetadata struct {
	Ctx              context.Context
	Tag              string
	AttachedArtifact *artifact.Artifact
}

// Resolve to the event from the metadata
func (c *CreateTagEventMetadata) Resolve(event *event.Event) error {
	data := &event2.CreateTagEvent{
		EventType:        event2.TopicCreateTag,
		Repository:       c.AttachedArtifact.RepositoryName,
		Tag:              c.Tag,
		AttachedArtifact: c.AttachedArtifact,
		OccurAt:          time.Now(),
	}
	cx, exist := security.FromContext(c.Ctx)
	if exist {
		data.Operator = cx.GetUsername()
	}
	event.Topic = event2.TopicCreateTag
	event.Data = data
	return nil
}

// DeleteTagEventMetadata is the metadata from which the delete tag event can be resolved
type DeleteTagEventMetadata struct {
	Ctx              context.Context
	Tag              string
	AttachedArtifact *artifact.Artifact
}

// Resolve to the event from the metadata
func (d *DeleteTagEventMetadata) Resolve(event *event.Event) error {
	data := &event2.DeleteTagEvent{
		EventType:        event2.TopicDeleteTag,
		Repository:       d.AttachedArtifact.RepositoryName,
		Tag:              d.Tag,
		AttachedArtifact: d.AttachedArtifact,
		OccurAt:          time.Now(),
	}
	ctx, exist := security.FromContext(d.Ctx)
	if exist {
		data.Operator = ctx.GetUsername()
	}
	event.Topic = event2.TopicDeleteTag
	event.Data = data
	return nil
}
