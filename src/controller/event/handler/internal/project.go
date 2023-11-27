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

package internal

import (
	"context"

	"github.com/goharbor/harbor/src/controller/event"
	"github.com/goharbor/harbor/src/controller/immutable"
	"github.com/goharbor/harbor/src/controller/retention"
	"github.com/goharbor/harbor/src/lib/log"
	"github.com/goharbor/harbor/src/pkg/member"
)

// ProjectEventHandler process project event data
type ProjectEventHandler struct {
}

// Name return the name of this handler
func (a *ProjectEventHandler) Name() string {
	return "InternalProject"
}

// IsStateful return false
func (a *ProjectEventHandler) IsStateful() bool {
	return false
}

func (a *ProjectEventHandler) onProjectDelete(ctx context.Context, event *event.DeleteProjectEvent) error {
	log.Infof("delete project id: %d", event.ProjectID)
	if err := immutable.Ctr.DeleteImmutableRuleByProject(ctx, event.ProjectID); err != nil {
		log.Errorf("failed to delete immutable rule, error %v", err)
	}
	if err := retention.Ctl.DeleteRetentionByProject(ctx, event.ProjectID); err != nil {
		log.Errorf("failed to delete retention rule, error %v", err)
	}
	if err := member.Mgr.DeleteMemberByProjectID(ctx, event.ProjectID); err != nil {
		log.Errorf("failed to delete project member, error %v", err)
	}
	return nil
}

// Handle handle project event
func (a *ProjectEventHandler) Handle(ctx context.Context, value interface{}) error {
	switch v := value.(type) {
	case *event.DeleteProjectEvent:
		return a.onProjectDelete(ctx, v)
	default:
		log.Errorf("Can not handler this event type! %#v", v)
	}
	return nil
}
