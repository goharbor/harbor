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

package role // nolint:revive

import (
	"net/http"
	"strconv"

	"github.com/goharbor/harbor/src/common/rbac"
	"github.com/goharbor/harbor/src/controller/event/metadata/commonevent"
	"github.com/goharbor/harbor/src/lib/log"
	"github.com/goharbor/harbor/src/lib/orm"
	"github.com/goharbor/harbor/src/pkg/auditext/event"
	pkgRole "github.com/goharbor/harbor/src/pkg/role"
)

const urlPattern = `^/api/v2.0/roles/(\d+)$`

func init() {
	roleResolver := &event.Resolver{
		ResourceType:      rbac.ResourceRole.String(),
		SucceedCodes:      []int{http.StatusCreated, http.StatusOK},
		ShouldResolveName: true,
		IDToNameFunc:      roleIDToName,
		ResourceIDPattern: urlPattern,
	}
	commonevent.RegisterResolver(`/api/v2.0/roles$`, roleResolver)
	commonevent.RegisterResolver(urlPattern, roleResolver)
}

// roleIDToName resolves a role id to its name. It is used so create/update log
// the current name (including after a rename) and delete can record the name
// before the row is gone.
func roleIDToName(roleID string) string {
	id, err := strconv.ParseInt(roleID, 10, 64)
	if err != nil {
		log.Errorf("failed to parse roleID: %v to int", roleID)
		return ""
	}
	// Use a fresh ORM context so the role is visible before the request
	// transaction commits.
	r, err := pkgRole.Mgr.Get(orm.Context(), id)
	if err != nil {
		log.Errorf("failed to get role by id: %v, err: %v", roleID, err)
		return ""
	}
	return r.Name
}
