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
	event2 "github.com/goharbor/harbor/src/api/event"
	"github.com/goharbor/harbor/src/pkg/notifier/event"
	"time"
)

// CreateProjectEventMetadata is the metadata from which the create project event can be resolved
type CreateProjectEventMetadata struct {
	ProjectID int64
	Project   string
	Operator  string
}

// Resolve to the event from the metadata
func (c *CreateProjectEventMetadata) Resolve(event *event.Event) error {
	event.Topic = event2.TopicCreateProject
	event.Data = &event2.CreateProjectEvent{
		EventType: event2.TopicCreateProject,
		ProjectID: c.ProjectID,
		Project:   c.Project,
		Operator:  c.Operator,
		OccurAt:   time.Now(),
	}
	return nil
}

// DeleteProjectEventMetadata is the metadata from which the delete project event can be resolved
type DeleteProjectEventMetadata struct {
	ProjectID int64
	Project   string
	Operator  string
}

// Resolve to the event from the metadata
func (d *DeleteProjectEventMetadata) Resolve(event *event.Event) error {
	event.Topic = event2.TopicDeleteProject
	event.Data = &event2.DeleteProjectEvent{
		EventType: event2.TopicDeleteProject,
		ProjectID: d.ProjectID,
		Project:   d.Project,
		Operator:  d.Operator,
		OccurAt:   time.Now(),
	}
	return nil
}
