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

	"github.com/vmware/harbor/src/replication/core"
	"github.com/vmware/harbor/src/replication/event/notification"
)

//StartReplicationHandler implements the notification handler interface to handle start replication requests.
type StartReplicationHandler struct{}

//Handle implements the same method of notification handler interface
func (srh *StartReplicationHandler) Handle(value interface{}) error {
	if value == nil {
		return errors.New("StartReplicationHandler can not handle nil value")
	}

	vType := reflect.TypeOf(value)
	if vType.Kind() != reflect.Struct || vType.String() != "notification.StartReplicationNotification" {
		return fmt.Errorf("Mismatch value type of StartReplicationHandler, expect %s but got %s", "notification.StartReplicationNotification", vType.String())
	}

	notification := value.(notification.StartReplicationNotification)
	if notification.PolicyID <= 0 {
		return errors.New("Invalid policy")
	}

	//Start replication
	return core.GlobalController.Replicate(notification.PolicyID, notification.Metadata)
}

//IsStateful implements the same method of notification handler interface
func (srh *StartReplicationHandler) IsStateful() bool {
	//Stateless
	return false
}
