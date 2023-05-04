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

package v2auth

import (
	"context"
	"fmt"
	"net/http"

	"github.com/goharbor/harbor/src/common/rbac"
	"github.com/goharbor/harbor/src/lib"
	"github.com/goharbor/harbor/src/lib/log"
)

type target int

const (
	login target = iota
	catalog
	repository
)

func (t target) String() string {
	return []string{"login", "catalog", "repository"}[t]
}

type access struct {
	target target
	name   string
	action rbac.Action
}

func (a access) scopeStr(ctx context.Context) string {
	logger := log.G(ctx)
	if a.target != repository {
		// Currently we do not support providing a token to list catalog
		return ""
	}
	act := ""
	if a.action == rbac.ActionPull {
		act = "pull"
	} else if a.action == rbac.ActionPush {
		act = "pull,push"
	} else if a.action == rbac.ActionDelete {
		act = "delete"
	} else {
		logger.Warningf("Invalid action in access: %s, returning empty scope", a.action)
		return ""
	}
	return fmt.Sprintf("repository:%s:%s", a.name, act)
}

func getAction(req *http.Request) rbac.Action {
	actions := map[string]rbac.Action{
		http.MethodPost:   rbac.ActionPush,
		http.MethodPatch:  rbac.ActionPush,
		http.MethodPut:    rbac.ActionPush,
		http.MethodGet:    rbac.ActionPull,
		http.MethodHead:   rbac.ActionPull,
		http.MethodDelete: rbac.ActionDelete,
	}
	if action, ok := actions[req.Method]; ok {
		return action
	}
	return ""
}

func accessList(req *http.Request) []access {
	l := make([]access, 0, 4)
	if req.URL.Path == "/v2/" {
		l = append(l, access{
			target: login,
		})
		return l
	}
	if lib.V2CatalogURLRe.MatchString(req.URL.Path) {
		l = append(l, access{
			target: catalog,
		})
		return l
	}
	none := lib.ArtifactInfo{}
	if a := lib.GetArtifactInfo(req.Context()); a != none {
		action := getAction(req)
		if action == "" {
			return l
		}
		l = append(l, access{
			target: repository,
			name:   a.Repository,
			action: action,
		})
		if req.Method == http.MethodPost && a.BlobMountRepository != "" { // need pull access for blob mount
			l = append(l, access{
				target: repository,
				name:   a.BlobMountRepository,
				action: rbac.ActionPull,
			})
		}
	}
	return l
}
