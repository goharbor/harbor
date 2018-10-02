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
	"errors"
	"fmt"
	"reflect"

	common_models "github.com/goharbor/harbor/src/common/models"
	"github.com/goharbor/harbor/src/common/utils"
	"github.com/goharbor/harbor/src/common/utils/log"
	"github.com/goharbor/harbor/src/core/notifier"
	"github.com/goharbor/harbor/src/replication"
	"github.com/goharbor/harbor/src/replication/event/notification"
	"github.com/goharbor/harbor/src/replication/event/topic"
	"github.com/goharbor/harbor/src/replication/models"
	"github.com/goharbor/harbor/src/replication/trigger"
)

// OnPushHandler implements the notification handler interface to handle image on push event.
type OnPushHandler struct{}

// Handle implements the same method of notification handler interface
func (oph *OnPushHandler) Handle(value interface{}) error {
	if value == nil {
		return errors.New("OnPushHandler can not handle nil value")
	}

	vType := reflect.TypeOf(value)
	if vType.Kind() != reflect.Struct || vType.String() != "notification.OnPushNotification" {
		return fmt.Errorf("Mismatch value type of OnPushHandler, expect %s but got %s", "notification.OnPushNotification", vType.String())
	}

	notification := value.(notification.OnPushNotification)

	return checkAndTriggerReplication(notification.Image, common_models.RepOpTransfer)
}

// IsStateful implements the same method of notification handler interface
func (oph *OnPushHandler) IsStateful() bool {
	// Statless
	return false
}

// checks whether replication policy is set on the resource, if is, trigger the replication
func checkAndTriggerReplication(image, operation string) error {
	project, _ := utils.ParseRepository(image)
	watchItems, err := trigger.DefaultWatchList.Get(project, operation)
	if err != nil {
		return fmt.Errorf("failed to get watch list for resource %s, operation %s: %v",
			image, operation, err)
	}
	if len(watchItems) == 0 {
		log.Debugf("no replication should be triggered for resource %s, operation %s, skip", image, operation)
		return nil
	}

	for _, watchItem := range watchItems {
		item := models.FilterItem{
			Kind:      replication.FilterItemKindTag,
			Value:     image,
			Operation: operation,
		}

		if err := notifier.Publish(topic.StartReplicationTopic, notification.StartReplicationNotification{
			PolicyID: watchItem.PolicyID,
			Metadata: map[string]interface{}{
				"candidates": []models.FilterItem{item},
			},
		}); err != nil {
			return fmt.Errorf("failed to publish replication topic for resource %s, operation %s, policy %d: %v",
				image, operation, watchItem.PolicyID, err)
		}
		log.Infof("replication topic for resource %s, operation %s, policy %d triggered",
			image, operation, watchItem.PolicyID)
	}
	return nil
}
