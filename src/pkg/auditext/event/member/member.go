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

package member

import (
	"context"
	"fmt"
	"net/http"
	"regexp"
	"strconv"
	"strings"
	"time"

	beegoorm "github.com/beego/beego/v2/client/orm"
	"github.com/goharbor/harbor/src/common"
	"github.com/goharbor/harbor/src/common/rbac"
	ctlevent "github.com/goharbor/harbor/src/controller/event"
	"github.com/goharbor/harbor/src/controller/event/metadata/commonevent"
	"github.com/goharbor/harbor/src/controller/event/model"
	"github.com/goharbor/harbor/src/lib/config"
	"github.com/goharbor/harbor/src/lib/log"
	"github.com/goharbor/harbor/src/lib/orm"
	"github.com/goharbor/harbor/src/pkg"
	ext "github.com/goharbor/harbor/src/pkg/auditext/event"
	pkgMember "github.com/goharbor/harbor/src/pkg/member"
	"github.com/goharbor/harbor/src/pkg/notifier/event"
)

const (
	memberCreatePattern = `/api/v2\.0/projects/[^/]+/members$`
	memberActionPattern = `/api/v2\.0/projects/[^/]+/members/\d+$`
	extractPattern      = `/api/v2\.0/projects/([^/]+)/members(?:/(\d+))?$`
)

var extractRe = regexp.MustCompile(extractPattern)

// overridable for testing
var (
	lookupMemberFn      = lookupMember
	resolveProjectFn    = resolveProject
	auditEventEnabledFn = auditLogMemberEventEnabled
)

func auditLogMemberEventEnabled(ctx context.Context, operation string) bool {
	if len(operation) == 0 {
		return false
	}
	return config.AuditLogEventEnabled(ctx, fmt.Sprintf("%v_%v", operation, rbac.ResourceMember.String()))
}

// custom resolver for project member events, extracts project id and user/group ids
func init() {
	r := &resolver{}
	commonevent.RegisterResolver(memberCreatePattern, r)
	commonevent.RegisterResolver(memberActionPattern, r)
}

type resolver struct{}

func (r *resolver) PreCheck(ctx context.Context, url string, method string) (bool, string) {
	ormCtx := ensureORMContext(ctx)
	operation := ext.MethodToOperation(method)
	if len(operation) == 0 {
		return false, ""
	}
	if !auditEventEnabledFn(ormCtx, operation) {
		return false, ""
	}
	m := extractRe.FindStringSubmatch(url)
	if len(m) < 2 {
		return false, ""
	}
	// for DELETE, resolve member info before the resource is deleted
	if method == http.MethodDelete {
		if len(m) >= 3 && len(m[2]) > 0 {
			name, typ := lookupMemberFn(ormCtx, m[1], m[2])
			if len(typ) > 0 {
				return true, typ + ":" + name
			}
			return true, name
		}
	}
	return true, ""
}

func (r *resolver) Resolve(ce *commonevent.Metadata, evt *event.Event) error {
	if ce == nil || evt == nil {
		return fmt.Errorf("metadata or event is nil")
	}
	ormCtx := ensureORMContext(ce.Ctx)
	operation := ext.MethodToOperation(ce.RequestMethod)
	if len(operation) == 0 {
		return nil
	}
	matches := extractRe.FindStringSubmatch(ce.RequestURL)
	if len(matches) < 2 {
		return nil
	}

	projectID, projectName := resolveProjectFn(ce.Ctx, matches[1])

	e := &model.CommonEvent{
		Operator:     ce.Username,
		ResourceType: rbac.ResourceMember.String(),
		Operation:    operation,
		ProjectID:    projectID,
		OcurrAt:      time.Now(),
		IsSuccessful: true,
	}

	var entityName, entityType string
	switch operation {
	case "create":
		e.IsSuccessful = ce.ResponseCode == http.StatusCreated
		if m := extractRe.FindStringSubmatch(ce.ResponseLocation); len(m) >= 3 && len(m[2]) > 0 {
			entityName, entityType = lookupMemberFn(ormCtx, m[1], m[2])
		}
	case "delete":
		e.IsSuccessful = ce.ResponseCode == http.StatusOK
		entityName, entityType = parsePreResolved(ce.ResourceName)
	case "update":
		e.IsSuccessful = ce.ResponseCode == http.StatusOK
		if e.IsSuccessful && len(matches) >= 3 && len(matches[2]) > 0 {
			entityName, entityType = lookupMemberFn(ormCtx, matches[1], matches[2])
		} else if len(matches) >= 3 && len(matches[2]) > 0 {
			entityName = matches[2]
		}
	}

	e.ResourceName = entityName
	label := "member"
	noun := ""
	if entityType == common.GroupMember {
		label = "group"
		noun = " member"
	} else if entityType == common.UserMember {
		label = "user"
		noun = " member"
	}
	preposition := "in"
	if operation == "delete" {
		preposition = "from"
	}
	resourceTarget := fmt.Sprintf("%s%s", label, noun)
	if len(entityName) > 0 {
		resourceTarget = fmt.Sprintf("%s %s", resourceTarget, entityName)
	} else if len(matches) >= 3 && len(matches[2]) > 0 {
		resourceTarget = fmt.Sprintf("%s ID %s", resourceTarget, matches[2])
	}
	e.OperationDescription = fmt.Sprintf("%s %s %s project %s",
		operation, resourceTarget, preposition, projectName)

	evt.Topic = ctlevent.TopicCommonEvent
	evt.Data = e
	return nil
}

func lookupMember(ctx context.Context, projectNameOrID, memberIDStr string) (string, string) {
	ormCtx := ensureORMContext(ctx)
	projectID, _ := resolveProjectFn(ormCtx, projectNameOrID)
	if projectID == 0 {
		return memberIDStr, ""
	}
	memberID, err := strconv.Atoi(memberIDStr)
	if err != nil {
		return memberIDStr, ""
	}
	m, err := pkgMember.Mgr.Get(ormCtx, projectID, memberID)
	if err != nil {
		log.Errorf("failed to get member %d in project %d: %v", memberID, projectID, err)
		return memberIDStr, ""
	}
	return m.Entityname, m.EntityType
}

// resolveProject resolves a project name or ID string to (projectID, projectName).
func resolveProject(ctx context.Context, projectNameOrID string) (int64, string) {
	ormCtx := ensureORMContext(ctx)
	if id, err := strconv.ParseInt(projectNameOrID, 10, 64); err == nil {
		if p, err := pkg.ProjectMgr.Get(ormCtx, id); err == nil && p != nil {
			return p.ProjectID, p.Name
		} else if err != nil {
			log.Errorf("failed to resolve project %d: %v", id, err)
		}
	}
	p, err := pkg.ProjectMgr.Get(ormCtx, projectNameOrID)
	if err != nil {
		log.Errorf("failed to resolve project %s: %v", projectNameOrID, err)
		return 0, projectNameOrID
	}
	if p == nil {
		return 0, projectNameOrID
	}
	return p.ProjectID, p.Name
}

func parsePreResolved(info string) (string, string) {
	if parts := strings.SplitN(info, ":", 2); len(parts) == 2 {
		return parts[1], parts[0]
	}
	return info, ""
}

func ensureORMContext(ctx context.Context) context.Context {
	if ctx == nil {
		return orm.Context()
	}
	if _, err := orm.FromContext(ctx); err == nil {
		return ctx
	}
	return orm.NewContext(ctx, beegoorm.NewOrm())
}
