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
	"github.com/goharbor/harbor/src/pkg/notifier/event"
)

// ReplicationMetaData defines replication related event data
type ReplicationMetaData struct {
	ReplicationTaskID int64
	Status            string
}

// Resolve replication metadata into replication event
func (r *ReplicationMetaData) Resolve(evt *event.Event) error {
	data := &event2.ReplicationEvent{
		ReplicationTaskID: r.ReplicationTaskID,
		EventType:         event2.TopicReplication,
		OccurAt:           time.Now(),
		Status:            r.Status,
	}

	evt.Topic = event2.TopicReplication
	evt.Data = data
	return nil
}
