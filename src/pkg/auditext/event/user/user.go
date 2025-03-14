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

package user

import (
	"net/http"
	"strconv"
	"strings"

	"github.com/goharbor/harbor/src/common/rbac"
	"github.com/goharbor/harbor/src/controller/event/metadata/commonevent"
	"github.com/goharbor/harbor/src/controller/event/model"
	"github.com/goharbor/harbor/src/lib/log"
	"github.com/goharbor/harbor/src/lib/orm"
	"github.com/goharbor/harbor/src/pkg/auditext/event"
	notifierevent "github.com/goharbor/harbor/src/pkg/notifier/event"
	pkgUser "github.com/goharbor/harbor/src/pkg/user"
)

const urlPattern = `^/api/v2.0/users/(\d+)(/password|/sysadmin)?$`

func init() {
	var userResolver = &userEventResolver{
		Resolver: event.Resolver{
			ResourceType:        rbac.ResourceUser.String(),
			SucceedCodes:        []int{http.StatusCreated, http.StatusOK},
			SensitiveAttributes: []string{"password"},
			ShouldResolveName:   true,
			IDToNameFunc:        userIDToName,
			ResourceIDPattern:   urlPattern,
		},
	}
	commonevent.RegisterResolver(`/api/v2.0/users$`, userResolver)
	commonevent.RegisterResolver(urlPattern, userResolver)
}

type userEventResolver struct {
	event.Resolver
}

// userIDToName convert user id to user name
func userIDToName(userID string) string {
	id, err := strconv.ParseInt(userID, 10, 32)
	if err != nil {
		log.Errorf("failed to parse userID: %v to int", userID)
		return ""
	}
	// use different context to so that the user is visible before the transaction is committed
	user, err := pkgUser.Mgr.Get(orm.Context(), int(id))
	if err != nil {
		log.Errorf("failed to parse userID: %v to int, err %v", userID, err)
		return ""
	}
	return user.Username
}

func (u *userEventResolver) Resolve(ce *commonevent.Metadata, event *notifierevent.Event) error {
	if err := u.Resolver.Resolve(ce, event); err != nil {
		return err
	}
	if ce.RequestMethod != http.MethodPut {
		return nil
	}
	// update operation description for update user password and add/remove user as system administrator
	origin := event.Data.(*model.CommonEvent).OperationDescription
	if strings.HasSuffix(ce.RequestURL, "/sysadmin") {
		event.Data.(*model.CommonEvent).OperationDescription = origin + ", add/remove user as system administrator"
	} else if strings.HasSuffix(ce.RequestURL, "/password") {
		event.Data.(*model.CommonEvent).OperationDescription = origin + ", change user password"
	}
	return nil
}
