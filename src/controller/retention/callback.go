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
	"github.com/goharbor/harbor/src/common/secret"
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
	TriggerRetentionExec(ctx context.Context, policyID int64, trigger string, dryRun bool) (int64, error)
	FailRetentionExec(ctx context.Context, policyID int64, trigger string, message string) (int64, error)
}

type principalResolver func(context.Context, *policy.ExecutionPrincipal) (security.Context, error)

type defaultPrincipalResolver struct {
	userMgr  user.Manager
	robotCtl robotCtl.Controller
}

func init() {
	err := scheduler.RegisterCallbackFunc(SchedulerCallback, retentionCallback)
	if err != nil {
		log.Errorf("failed to register retention callback, %v", err)
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
		// Record the rejection as a failed execution so it shows up on the retention page.
		log.Warningf("scheduled retention policy %d rejected for %s: %v",
			param.PolicyID, describePrincipal(metadata), err)
		if _, failErr := ctl.FailRetentionExec(ctx, param.PolicyID, param.Trigger, err.Error()); failErr != nil {
			log.Errorf("failed to record the rejected retention execution for policy %d: %v",
				param.PolicyID, failErr)
		}
		return err
	}

	if sc == nil {
		ctx = context.WithValue(ctx, operator.ContextKey{}, secret.JobserviceUser)
	} else {
		ctx = security.NewContext(ctx, sc)
		ctx = context.WithValue(ctx, operator.ContextKey{}, sc.GetUsername())
	}
	_, err = ctl.TriggerRetentionExec(ctx, param.PolicyID, param.Trigger, false)
	return err
}

// authorizeScheduledExecution returns the security context the schedule runs under, or nil for a
// policy Harbor created itself.
func authorizeScheduledExecution(
	ctx context.Context,
	metadata *policy.Metadata,
	resolve principalResolver,
) (security.Context, error) {
	if metadata == nil || metadata.ExecutionPrincipal == nil {
		return nil, fmt.Errorf("scheduled retention policy has no recorded execution principal; " +
			"save the policy again as a user who can delete artifacts in the project")
	}
	if metadata.Scope == nil || metadata.Scope.Level != policy.ScopeLevelProject {
		return nil, fmt.Errorf("scheduled retention policy has invalid project scope")
	}
	principal := metadata.ExecutionPrincipal
	if principal.Type == policy.ExecutionPrincipalTypeSystem {
		return nil, nil
	}
	if principal.ID <= 0 {
		return nil, fmt.Errorf("scheduled retention policy has invalid execution principal")
	}

	sc, err := resolve(ctx, principal)
	if err != nil {
		return nil, fmt.Errorf("failed to resolve scheduled retention principal: %w", err)
	}
	resource := rbacProject.NewNamespace(metadata.Scope.Reference).Resource(rbac.ResourceArtifact)
	if !sc.Can(ctx, rbac.ActionDelete, resource) {
		return nil, fmt.Errorf("%s %q can no longer delete artifacts in project %d; "+
			"save the policy again as a user who can",
			principal.Type, sc.GetUsername(), metadata.Scope.Reference)
	}
	return sc, nil
}

func describePrincipal(metadata *policy.Metadata) string {
	if metadata == nil || metadata.ExecutionPrincipal == nil {
		return "no principal"
	}
	return fmt.Sprintf("%s principal %d", metadata.ExecutionPrincipal.Type, metadata.ExecutionPrincipal.ID)
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
		// Authority only the identity provider knows is reapplied from the schedule. The local
		// sysadmin flag can be revoked since the schedule was saved but never regained from the
		// database, so an external identity sharing a local admin's name stays unprivileged.
		account.SysAdminFlag = account.SysAdminFlag && principal.SysAdmin
		account.AdminRoleInAuth = principal.AdminRoleInAuth
		account.GroupIDs = principal.GroupIDs
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
