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

	"github.com/goharbor/harbor/src/common/rbac"
	event2 "github.com/goharbor/harbor/src/controller/event"
	evtModel "github.com/goharbor/harbor/src/controller/event/model"
	"github.com/goharbor/harbor/src/pkg/notifier/event"
)

// OIDCOnboardEventMetadata is the metadata from which the OIDC onboard login can be resolved.
type OIDCOnboardEventMetadata struct {
	Username string
}

// Resolve to the event from the metadata
func (o *OIDCOnboardEventMetadata) Resolve(event *event.Event) error {
	data := &evtModel.CommonEvent{
		Operator:             o.Username,
		ResourceType:         rbac.ResourceUser.String(),
		ResourceName:         o.Username,
		OcurrAt:              time.Now(),
		Operation:            "login",
		OperationDescription: "login",
		IsSuccessful:         true,
	}
	event.Topic = event2.TopicCommonEvent
	event.Data = data
	return nil
}
