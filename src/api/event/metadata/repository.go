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
	"github.com/goharbor/harbor/src/common/security"
	"github.com/goharbor/harbor/src/pkg/notifier/event"
	"time"
)

// DeleteRepositoryEventMetadata is the metadata from which the delete repository event can be resolved
type DeleteRepositoryEventMetadata struct {
	Ctx        context.Context
	Repository string
	ProjectID  int64
}

// Resolve to the event from the metadata
func (d *DeleteRepositoryEventMetadata) Resolve(event *event.Event) error {
	data := &event2.DeleteRepositoryEvent{
		Repository: d.Repository,
		ProjectID:  d.ProjectID,
		OccurAt:    time.Now(),
	}
	cx, exist := security.FromContext(d.Ctx)
	if exist {
		data.Operator = cx.GetUsername()
	}
	event.Topic = event2.TopicDeleteRepository
	event.Data = data
	return nil
}
