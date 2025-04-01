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

	"github.com/goharbor/harbor/src/common"
	"github.com/goharbor/harbor/src/common/rbac"
	"github.com/goharbor/harbor/src/common/security"
	ctlevent "github.com/goharbor/harbor/src/controller/event"
	"github.com/goharbor/harbor/src/controller/event/metadata/commonevent"
	"github.com/goharbor/harbor/src/controller/event/model"
	"github.com/goharbor/harbor/src/controller/user"
	"github.com/goharbor/harbor/src/lib/config"
	"github.com/goharbor/harbor/src/lib/log"
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

// check if current auth is oidc common user
func isOIDCAuthCommonUser(ctx context.Context, username string) bool {
	authMode, err := config.AuthMode(ctx)
	if err != nil || common.OIDCAuth != authMode {
		return false
	}
	u, err := user.Ctl.GetByName(ctx, username)
	if err != nil {
		log.Errorf("failed to get username %v, error %v", username, err)
		return false
	}
	// for admin user under oidc, it should be handled by /c/log_out, not /c/oidc/logout
	if u.UserID == 1 {
		return false
	}
	return true
}

func (l *logoutResolver) PreCheck(ctx context.Context, _ string, method string) (bool, string) {
	operation := ""
	if method == http.MethodGet {
		operation = opLogout
	}
	if len(operation) == 0 {
		return false, ""
	}
	// current /c/log_out is request is sent twice, ignore the second time
	secCtx, ok := security.FromContext(ctx)
	if !ok {
		return false, ""
	}
	username := secCtx.GetUsername()
	if len(username) == 0 {
		return false, ""
	}
	// for oidc auth common user logout, it is handled by oidclogoutResolver
	if isOIDCAuthCommonUser(ctx, username) {
		return false, ""
	}
	return config.AuditLogEventEnabled(ctx, fmt.Sprintf("%v_%v", operation, rbac.ResourceUser.String())), ""
}
