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

package retention

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/goharbor/harbor/src/common/rbac"
	rbacProject "github.com/goharbor/harbor/src/common/rbac/project"
	"github.com/goharbor/harbor/src/common/security"
	localSecurity "github.com/goharbor/harbor/src/common/security/local"
	robotSecurity "github.com/goharbor/harbor/src/common/security/robot"
	"github.com/goharbor/harbor/src/controller/event/operator"
	robotCtl "github.com/goharbor/harbor/src/controller/robot"
	"github.com/goharbor/harbor/src/lib/log"
	"github.com/goharbor/harbor/src/pkg/retention/policy"
	"github.com/goharbor/harbor/src/pkg/scheduler"
	"github.com/goharbor/harbor/src/pkg/user"
)

type scheduledRetentionController interface {
	GetRetention(ctx context.Context, id int64) (*policy.Metadata, error)
	TriggerRetentionExec(
		ctx context.Context,
		policyID int64,
		trigger string,
		dryRun bool,
	) (int64, error)
}

type principalResolver func(context.Context, *policy.ExecutionPrincipal) (security.Context, error)

type defaultPrincipalResolver struct {
	userMgr  user.Manager
	robotCtl robotCtl.Controller
}

func init() {
	err := scheduler.RegisterCallbackFunc(SchedulerCallback, retentionCallback)
	if err != nil {
		log.Fatalf("failed to register retention callback, %v", err)
	}
}

func retentionCallback(ctx context.Context, p string) error {
	resolver := &defaultPrincipalResolver{
		userMgr:  user.Mgr,
		robotCtl: robotCtl.Ctl,
	}
	return runRetentionCallback(ctx, p, Ctl, resolver.resolve)
}

func runRetentionCallback(
	ctx context.Context,
	p string,
	ctl scheduledRetentionController,
	resolve principalResolver,
) error {
	param := &TriggerParam{}
	if err := json.Unmarshal([]byte(p), param); err != nil {
		return fmt.Errorf("failed to unmarshal the param: %v", err)
	}

	metadata, err := ctl.GetRetention(ctx, param.PolicyID)
	if err != nil {
		return fmt.Errorf("failed to load retention policy %d: %w", param.PolicyID, err)
	}
	sc, err := authorizeScheduledExecution(ctx, metadata, resolve)
	if err != nil {
		return err
	}

	ctx = security.NewContext(ctx, sc)
	ctx = context.WithValue(ctx, operator.ContextKey{}, sc.GetUsername())
	_, err = ctl.TriggerRetentionExec(ctx, param.PolicyID, param.Trigger, false)
	return err
}

func authorizeScheduledExecution(
	ctx context.Context,
	metadata *policy.Metadata,
	resolve principalResolver,
) (security.Context, error) {
	if metadata == nil || metadata.ExecutionPrincipal == nil {
		return nil, fmt.Errorf("scheduled retention policy has no authorized execution principal")
	}
	if metadata.Scope == nil || metadata.Scope.Level != policy.ScopeLevelProject {
		return nil, fmt.Errorf("scheduled retention policy has invalid project scope")
	}
	if metadata.ExecutionPrincipal.ID <= 0 {
		return nil, fmt.Errorf("scheduled retention policy has invalid execution principal")
	}

	sc, err := resolve(ctx, metadata.ExecutionPrincipal)
	if err != nil {
		return nil, fmt.Errorf("failed to resolve scheduled retention principal: %w", err)
	}
	resource := rbacProject.NewNamespace(metadata.Scope.Reference).Resource(rbac.ResourceArtifact)
	if !sc.Can(ctx, rbac.ActionDelete, resource) {
		return nil, fmt.Errorf("scheduled retention principal cannot delete project artifacts")
	}
	return sc, nil
}

func (r *defaultPrincipalResolver) resolve(
	ctx context.Context,
	principal *policy.ExecutionPrincipal,
) (security.Context, error) {
	switch principal.Type {
	case policy.ExecutionPrincipalTypeLocal:
		account, err := r.userMgr.Get(ctx, int(principal.ID))
		if err != nil {
			return nil, err
		}
		if account == nil || account.Deleted {
			return nil, fmt.Errorf("local user %d is deleted", principal.ID)
		}
		return localSecurity.NewSecurityContext(account), nil
	case policy.ExecutionPrincipalTypeRobot:
		account, err := r.robotCtl.Get(ctx, principal.ID, &robotCtl.Option{WithPermission: true})
		if err != nil {
			return nil, err
		}
		if account == nil {
			return nil, fmt.Errorf("robot %d is unavailable", principal.ID)
		}
		if account.Disabled {
			return nil, fmt.Errorf("robot %d is disabled", principal.ID)
		}
		if account.ExpiresAt != -1 && account.ExpiresAt <= time.Now().Unix() {
			return nil, fmt.Errorf("robot %d is expired", principal.ID)
		}
		return robotSecurity.NewSecurityContext(account), nil
	default:
		return nil, fmt.Errorf("unsupported execution principal type %q", principal.Type)
	}
}
