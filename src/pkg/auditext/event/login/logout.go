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

package login

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/goharbor/harbor/src/common/rbac"
	ctlevent "github.com/goharbor/harbor/src/controller/event"
	"github.com/goharbor/harbor/src/controller/event/metadata/commonevent"
	"github.com/goharbor/harbor/src/controller/event/model"
	"github.com/goharbor/harbor/src/lib/config"
	"github.com/goharbor/harbor/src/pkg/notifier/event"
)

type logoutResolver struct {
}

func (l *logoutResolver) Resolve(ce *commonevent.Metadata, event *event.Event) error {
	e := &model.CommonEvent{
		Operator:             ce.Username,
		ResourceType:         rbac.ResourceUser.String(),
		ResourceName:         ce.Username,
		OcurrAt:              time.Now(),
		Operation:            opLogout,
		OperationDescription: opLogout,
		IsSuccessful:         true,
	}
	if ce.ResponseCode != http.StatusOK {
		e.IsSuccessful = false
	}
	event.Topic = ctlevent.TopicCommonEvent
	event.Data = e
	return nil
}

func (l *logoutResolver) PreCheck(ctx context.Context, _ string, method string) (bool, string) {
	operation := ""
	if method == http.MethodGet {
		operation = opLogout
	}
	if len(operation) == 0 {
		return false, ""
	}
	return config.AuditLogEventEnabled(ctx, fmt.Sprintf("%v_%v", operation, rbac.ResourceUser.String())), ""
}
