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
	"time"

	event2 "github.com/goharbor/harbor/src/controller/event"
	"github.com/goharbor/harbor/src/lib/selector"
	"github.com/goharbor/harbor/src/pkg/notifier/event"
)

// RetentionMetaData defines tag retention related event data
type RetentionMetaData struct {
	Total    int
	Retained int
	Deleted  []*selector.Result
	Status   string
	TaskID   int64
}

// Resolve tag retention metadata into tag retention event
func (r *RetentionMetaData) Resolve(evt *event.Event) error {
	data := &event2.RetentionEvent{
		EventType: event2.TopicTagRetention,
		OccurAt:   time.Now(),
		Status:    r.Status,
		Deleted:   r.Deleted,
		TaskID:    r.TaskID,
		Total:     r.Total,
		Retained:  r.Retained,
	}

	evt.Topic = event2.TopicTagRetention
	evt.Data = data
	return nil
}
