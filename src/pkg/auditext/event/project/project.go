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

package project

import (
	"context"
	"fmt"
	"net/http"
	"regexp"
	"strconv"
	"time"

	ctlevent "github.com/goharbor/harbor/src/controller/event"
	"github.com/goharbor/harbor/src/controller/event/metadata/commonevent"
	"github.com/goharbor/harbor/src/controller/event/model"
	"github.com/goharbor/harbor/src/controller/project"
	"github.com/goharbor/harbor/src/lib/config"
	"github.com/goharbor/harbor/src/lib/orm"
	"github.com/goharbor/harbor/src/pkg/notifier/event"
	"github.com/goharbor/harbor/src/pkg/project/models"
)

// Each URL shape is declared once and used for both resolver registration and
// project-name extraction, so the two can never drift apart.
const (
	projectPattern        = `^/api/v2\.0/projects/([^/?]+)(?:\?.*)?$`
	projectMetaPattern    = `^/api/v2\.0/projects/([^/?]+)/metadatas/(?:\?.*)?$`
	projectMetaKeyPattern = `^/api/v2\.0/projects/([^/?]+)/metadatas/([^/?]+)(?:\?.*)?$`
)

var (
	projectReg        = regexp.MustCompile(projectPattern)
	projectMetaReg    = regexp.MustCompile(projectMetaPattern)
	projectMetaKeyReg = regexp.MustCompile(projectMetaKeyPattern)
)

func init() {
	var projectEventResolver = &resolver{}
	commonevent.RegisterResolver(projectPattern, projectEventResolver)
	commonevent.RegisterResolver(projectMetaPattern, projectEventResolver)
	commonevent.RegisterResolver(projectMetaKeyPattern, projectEventResolver)
}

type resolver struct {
}

func (r *resolver) Resolve(ce *commonevent.Metadata, evt *event.Event) error {
	projStr := getProjectNameOrID(ce.RequestURL)
	if projStr == "" {
		return fmt.Errorf("failed to parse project name or ID from URL: %s", ce.RequestURL)
	}

	ctx := ce.Ctx
	if _, err := orm.FromContext(ctx); err == nil {
		ctx = orm.Clone(ctx)
	}

	proj, err := getProject(ctx, projStr, ce.IsResourceName)
	if err != nil {
		return fmt.Errorf("failed to get project: %v", err)
	}

	e := &model.CommonEvent{}
	e.Operation = "update"
	e.Operator = ce.Username
	e.ResourceType = "project"
	e.ResourceName = proj.Name
	e.ProjectID = proj.ProjectID
	e.OcurrAt = time.Now()
	e.OperationDescription = fmt.Sprintf("update project: %s", proj.Name)

	if ce.ResponseCode == http.StatusOK {
		e.IsSuccessful = true
	}
	evt.Topic = ctlevent.TopicCommonEvent
	evt.Data = e
	return nil
}

func (r *resolver) PreCheck(ctx context.Context, url string, method string) (bool, string) {
	if method != http.MethodPut && method != http.MethodPost && method != http.MethodDelete {
		return false, ""
	}

	if method == http.MethodDelete && projectReg.MatchString(url) {
		return false, ""
	}

	return config.AuditLogEventEnabled(ctx, "update_project"), ""
}

func getProjectNameOrID(url string) string {
	if m := projectMetaKeyReg.FindStringSubmatch(url); len(m) > 1 {
		return m[1]
	}
	if m := projectMetaReg.FindStringSubmatch(url); len(m) > 1 {
		return m[1]
	}
	if m := projectReg.FindStringSubmatch(url); len(m) > 1 {
		return m[1]
	}
	return ""
}

// getProject resolves the path segment to a project. isName comes from the
// X-Is-Resource-Name header and forces the segment to be treated as a project
// name even when it is all digits — mirroring parseProjectNameOrID in the API
// handlers.
func getProject(ctx context.Context, str string, isName bool) (*models.Project, error) {
	var projectNameOrID any = str
	if !isName {
		if v, err := strconv.ParseInt(str, 10, 64); err == nil {
			projectNameOrID = v
		}
	}
	return project.Ctl.Get(ctx, projectNameOrID)
}
