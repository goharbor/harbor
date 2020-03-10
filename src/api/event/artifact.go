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
	"github.com/goharbor/harbor/src/common/security"
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
	data := &PushArtifactEvent{
		Repository: p.Artifact.RepositoryName,
		Artifact:   p.Artifact,
		Tag:        p.Tag,
		OccurAt:    time.Now(),
	}
	ctx, exist := security.FromContext(p.Ctx)
	if exist {
		data.Operator = ctx.GetUsername()
	}
	event.Topic = TopicPushArtifact
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
	data := &PullArtifactEvent{
		Repository: p.Artifact.RepositoryName,
		Artifact:   p.Artifact,
		Tag:        p.Tag,
		OccurAt:    time.Now(),
	}
	ctx, exist := security.FromContext(p.Ctx)
	if exist {
		data.Operator = ctx.GetUsername()
	}
	event.Topic = TopicPullArtifact
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
	data := &DeleteArtifactEvent{
		Repository: d.Artifact.RepositoryName,
		Artifact:   d.Artifact,
		Tags:       d.Tags,
		OccurAt:    time.Now(),
	}
	ctx, exist := security.FromContext(d.Ctx)
	if exist {
		data.Operator = ctx.GetUsername()
	}
	event.Topic = TopicDeleteArtifact
	event.Data = data
	return nil
}
