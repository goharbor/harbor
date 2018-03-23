// Copyright (c) 2017 VMware, Inc. All Rights Reserved.
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

	"github.com/vmware/harbor/src/common/models"
	"github.com/vmware/harbor/src/replication/event/notification"
)

//OnDeletionHandler implements the notification handler interface to handle image on push event.
type OnDeletionHandler struct{}

//Handle implements the same method of notification handler interface
func (oph *OnDeletionHandler) Handle(value interface{}) error {
	if value == nil {
		return errors.New("OnDeletionHandler can not handle nil value")
	}

	vType := reflect.TypeOf(value)
	if vType.Kind() != reflect.Struct || vType.String() != "notification.OnDeletionNotification" {
		return fmt.Errorf("Mismatch value type of OnDeletionHandler, expect %s but got %s", "notification.OnDeletionNotification", vType.String())
	}

	notification := value.(notification.OnDeletionNotification)
	return checkAndTriggerReplication(notification.Image, models.RepOpDelete)
}

//IsStateful implements the same method of notification handler interface
func (oph *OnDeletionHandler) IsStateful() bool {
	//Statless
	return false
}
