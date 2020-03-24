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

// PushArtifactEventMetadata is the metadata from which the push artifact event can be resolved
type PushArtifactEventMetadata struct {
	Ctx      context.Context
	Artifact *artifact.Artifact
	Tag      string
}

// Resolve to the event from the metadata
func (p *PushArtifactEventMetadata) Resolve(event *event.Event) error {
	ae := &event2.ArtifactEvent{
		EventType:  event2.TopicPushArtifact,
		Repository: p.Artifact.RepositoryName,
		Artifact:   p.Artifact,
		OccurAt:    time.Now(),
	}
	if p.Tag != "" {
		ae.Tags = []string{p.Tag}
	}
	data := &event2.PushArtifactEvent{
		ArtifactEvent: ae,
	}
	ctx, exist := security.FromContext(p.Ctx)
	if exist {
		data.Operator = ctx.GetUsername()
	}
	event.Topic = event2.TopicPushArtifact
	event.Data = data
	return nil
}

// PullArtifactEventMetadata is the metadata from which the pull artifact event can be resolved
type PullArtifactEventMetadata struct {
	Ctx      context.Context
	Artifact *artifact.Artifact
	Tag      string
}

// Resolve to the event from the metadata
func (p *PullArtifactEventMetadata) Resolve(event *event.Event) error {
	ae := &event2.ArtifactEvent{
		EventType:  event2.TopicPullArtifact,
		Repository: p.Artifact.RepositoryName,
		Artifact:   p.Artifact,
		OccurAt:    time.Now(),
	}
	if p.Tag != "" {
		ae.Tags = []string{p.Tag}
	}
	data := &event2.PullArtifactEvent{
		ArtifactEvent: ae,
	}
	ctx, exist := security.FromContext(p.Ctx)
	if exist {
		data.Operator = ctx.GetUsername()
	}
	event.Topic = event2.TopicPullArtifact
	event.Data = data
	return nil
}

// DeleteArtifactEventMetadata is the metadata from which the delete artifact event can be resolved
type DeleteArtifactEventMetadata struct {
	Ctx      context.Context
	Artifact *artifact.Artifact
	Tags     []string
}

// Resolve to the event from the metadata
func (d *DeleteArtifactEventMetadata) Resolve(event *event.Event) error {
	data := &event2.DeleteArtifactEvent{
		ArtifactEvent: &event2.ArtifactEvent{
			EventType:  event2.TopicDeleteArtifact,
			Repository: d.Artifact.RepositoryName,
			Artifact:   d.Artifact,
			Tags:       d.Tags,
			OccurAt:    time.Now(),
		},
	}
	ctx, exist := security.FromContext(d.Ctx)
	if exist {
		data.Operator = ctx.GetUsername()
	}
	event.Topic = event2.TopicDeleteArtifact
	event.Data = data
	return nil
}
