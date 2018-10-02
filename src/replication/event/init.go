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
	"github.com/goharbor/harbor/src/common/utils/log"
	"github.com/goharbor/harbor/src/core/notifier"
	"github.com/goharbor/harbor/src/replication/event/topic"
)

// Subscribe related topics
func init() {
	// Listen the related event topics
	handlers := map[string]notifier.NotificationHandler{
		topic.StartReplicationTopic:           &StartReplicationHandler{},
		topic.ReplicationEventTopicOnPush:     &OnPushHandler{},
		topic.ReplicationEventTopicOnDeletion: &OnDeletionHandler{},
	}

	for topic, handler := range handlers {
		if err := notifier.Subscribe(topic, handler); err != nil {
			log.Errorf("failed to subscribe topic %s: %v", topic, err)
			continue
		}
		log.Debugf("topic %s is subscribed", topic)
	}
}
