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
	"errors"
	"testing"

	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/goharbor/harbor/src/common"
	"github.com/goharbor/harbor/src/common/models"
	"github.com/goharbor/harbor/src/common/secret"
	"github.com/goharbor/harbor/src/common/security"
	"github.com/goharbor/harbor/src/controller/event/operator"
	projectCtl "github.com/goharbor/harbor/src/controller/project"
	robotCtl "github.com/goharbor/harbor/src/controller/robot"
	projectmodels "github.com/goharbor/harbor/src/pkg/project/models"
	"github.com/goharbor/harbor/src/pkg/retention/policy"
	robotmodel "github.com/goharbor/harbor/src/pkg/robot/model"
	"github.com/goharbor/harbor/src/pkg/user"
	projecttesting "github.com/goharbor/harbor/src/testing/controller/project"
	robottesting "github.com/goharbor/harbor/src/testing/controller/robot"
	usertesting "github.com/goharbor/harbor/src/testing/pkg/user"
)

const legacyScheduleParam = `{"PolicyID":2,"Trigger":"SCHEDULE","Operator":"harbor-jobservice"}`

type callbackController struct {
	Controller

	metadata       *policy.Metadata
	triggerCalls   int
	triggerContext context.Context
	failures       []string
}

func (c *callbackController) GetRetention(context.Context, int64) (*policy.Metadata, error) {
	return c.metadata, nil
}

func (c *callbackController) TriggerRetentionExec(ctx context.Context, _ int64, _ string, _ bool) (int64, error) {
	c.triggerCalls++
	c.triggerContext = ctx
	return 1, nil
}

func (c *callbackController) FailRetentionExec(_ context.Context, policyID int64, trigger string, message string) (int64, error) {
	if policyID != 2 || trigger != "SCHEDULE" {
		return 0, errors.New("unexpected execution attributes")
	}
	c.failures = append(c.failures, message)
	return int64(len(c.failures)), nil
}

func scheduledMetadata(principal *policy.ExecutionPrincipal) *policy.Metadata {
	metadata := policy.WithNDaysSinceLastPull(1, 7)
	metadata.ExecutionPrincipal = principal
	return metadata
}

func useCallbackController(t *testing.T, metadata *policy.Metadata) *callbackController {
	t.Helper()
	original := Ctl
	ctl := &callbackController{metadata: metadata}
	Ctl = ctl
	t.Cleanup(func() {
		Ctl = original
	})
	return ctl
}

func useUserManager(t *testing.T) *usertesting.Manager {
	t.Helper()
	original := user.Mgr
	users := &usertesting.Manager{}
	user.Mgr = users
	t.Cleanup(func() {
		user.Mgr = original
	})
	return users
}

func useProjectController(t *testing.T) *projecttesting.Controller {
	t.Helper()
	original := projectCtl.Ctl
	projects := &projecttesting.Controller{}
	projectCtl.Ctl = projects
	t.Cleanup(func() {
		projectCtl.Ctl = original
	})
	projects.On("Get", mock.Anything, int64(1), mock.Anything).
		Return(&projectmodels.Project{ProjectID: 1}, nil)
	return projects
}

func useRobotController(t *testing.T) *robottesting.Controller {
	t.Helper()
	original := robotCtl.Ctl
	robots := &robottesting.Controller{}
	robotCtl.Ctl = robots
	t.Cleanup(func() {
		robotCtl.Ctl = original
	})
	return robots
}

func requireRejected(t *testing.T, ctl *callbackController, err error, message string) {
	t.Helper()
	require.ErrorContains(t, err, message)
	require.Zero(t, ctl.triggerCalls)
	require.Len(t, ctl.failures, 1)
	require.Contains(t, ctl.failures[0], message)
}

func TestRetentionCallbackRejectsLegacySchedule(t *testing.T) {
	ctl := useCallbackController(t, scheduledMetadata(nil))

	err := retentionCallback(context.Background(), legacyScheduleParam)
	requireRejected(t, ctl, err, "no recorded execution principal")
}

func TestRetentionCallbackRunsSystemPolicy(t *testing.T) {
	ctl := useCallbackController(t, policy.WithNDaysSinceLastPull(1, 7))

	require.NoError(t, retentionCallback(context.Background(), legacyScheduleParam))
	require.Equal(t, 1, ctl.triggerCalls)
	require.Equal(t, secret.JobserviceUser, operator.FromContext(ctl.triggerContext))
	_, hasSecurity := security.FromContext(ctl.triggerContext)
	require.False(t, hasSecurity)
}

func TestRetentionCallbackRevalidatesLocalPrincipal(t *testing.T) {
	ctl := useCallbackController(t, scheduledMetadata(&policy.ExecutionPrincipal{
		Type: policy.ExecutionPrincipalTypeLocal, ID: 17, SysAdmin: true,
	}))
	useUserManager(t).On("Get", mock.Anything, 17).Return(&models.User{
		UserID:       17,
		Username:     "retention-admin",
		SysAdminFlag: true,
	}, nil).Once()

	require.NoError(t, retentionCallback(context.Background(), legacyScheduleParam))
	require.Equal(t, 1, ctl.triggerCalls)
	require.Equal(t, "retention-admin", operator.FromContext(ctl.triggerContext))
	resolved, ok := security.FromContext(ctl.triggerContext)
	require.True(t, ok)
	require.Equal(t, "retention-admin", resolved.GetUsername())
}

func TestRetentionCallbackReappliesRecordedGroups(t *testing.T) {
	ctl := useCallbackController(t, scheduledMetadata(&policy.ExecutionPrincipal{
		Type: policy.ExecutionPrincipalTypeLocal, ID: 21, GroupIDs: []int{7},
	}))
	useUserManager(t).On("Get", mock.Anything, 21).
		Return(&models.User{UserID: 21, Username: "ldap-maintainer"}, nil).Once()
	useProjectController(t).On("ListRoles", mock.Anything, int64(1), mock.MatchedBy(func(u *models.User) bool {
		return u.UserID == 21 && len(u.GroupIDs) == 1 && u.GroupIDs[0] == 7
	})).Return([]int{common.RoleMaintainer}, nil).Once()

	require.NoError(t, retentionCallback(context.Background(), legacyScheduleParam))
	require.Equal(t, 1, ctl.triggerCalls)
	require.Equal(t, "ldap-maintainer", operator.FromContext(ctl.triggerContext))
}

func TestRetentionCallbackNeverRegainsLocalSysAdmin(t *testing.T) {
	// An auth-proxy identity sharing the seeded admin's name was saved without the local flag.
	ctl := useCallbackController(t, scheduledMetadata(&policy.ExecutionPrincipal{
		Type: policy.ExecutionPrincipalTypeLocal, ID: 1,
	}))
	useUserManager(t).On("Get", mock.Anything, 1).
		Return(&models.User{UserID: 1, Username: "admin", SysAdminFlag: true}, nil).Once()
	useProjectController(t).On("ListRoles", mock.Anything, int64(1), mock.MatchedBy(func(u *models.User) bool {
		return u.UserID == 1 && !u.SysAdminFlag
	})).Return([]int{}, nil).Once()

	err := retentionCallback(context.Background(), legacyScheduleParam)
	requireRejected(t, ctl, err, `local "admin" can no longer delete artifacts in project 1`)
}

func TestRetentionCallbackRejectsDeveloperPrincipal(t *testing.T) {
	ctl := useCallbackController(t, scheduledMetadata(&policy.ExecutionPrincipal{
		Type: policy.ExecutionPrincipalTypeLocal, ID: 18,
	}))
	developer := &models.User{UserID: 18, Username: "developer"}
	useUserManager(t).On("Get", mock.Anything, 18).Return(developer, nil).Once()
	useProjectController(t).On("ListRoles", mock.Anything, int64(1), developer).
		Return([]int{common.RoleDeveloper}, nil).Once()

	err := retentionCallback(context.Background(), legacyScheduleParam)
	requireRejected(t, ctl, err, "can no longer delete artifacts in project 1")
}

func TestRetentionCallbackRejectsUnavailablePrincipals(t *testing.T) {
	t.Run("invalid principal ID", func(t *testing.T) {
		ctl := useCallbackController(t, scheduledMetadata(&policy.ExecutionPrincipal{
			Type: policy.ExecutionPrincipalTypeLocal, ID: 0,
		}))

		err := retentionCallback(context.Background(), legacyScheduleParam)
		requireRejected(t, ctl, err, "invalid execution principal")
	})

	t.Run("missing local user", func(t *testing.T) {
		ctl := useCallbackController(t, scheduledMetadata(&policy.ExecutionPrincipal{
			Type: policy.ExecutionPrincipalTypeLocal, ID: 19,
		}))
		useUserManager(t).On("Get", mock.Anything, 19).
			Return(nil, errors.New("user not found")).Once()

		err := retentionCallback(context.Background(), legacyScheduleParam)
		requireRejected(t, ctl, err, "user not found")
	})

	t.Run("deleted local account", func(t *testing.T) {
		ctl := useCallbackController(t, scheduledMetadata(&policy.ExecutionPrincipal{
			Type: policy.ExecutionPrincipalTypeLocal, ID: 20,
		}))
		useUserManager(t).On("Get", mock.Anything, 20).Return(&models.User{
			UserID:   20,
			Username: "deleted-user",
			Deleted:  true,
		}, nil).Once()

		err := retentionCallback(context.Background(), legacyScheduleParam)
		requireRejected(t, ctl, err, "local user 20 is deleted")
	})

	t.Run("disabled robot", func(t *testing.T) {
		ctl := useCallbackController(t, scheduledMetadata(&policy.ExecutionPrincipal{
			Type: policy.ExecutionPrincipalTypeRobot, ID: 29,
		}))
		useRobotController(t).On("Get", mock.Anything, int64(29), mock.Anything).
			Return(&robotCtl.Robot{Robot: robotmodel.Robot{ID: 29, Disabled: true}}, nil).Once()

		err := retentionCallback(context.Background(), legacyScheduleParam)
		requireRejected(t, ctl, err, "robot 29 is disabled")
	})

	t.Run("expired robot", func(t *testing.T) {
		ctl := useCallbackController(t, scheduledMetadata(&policy.ExecutionPrincipal{
			Type: policy.ExecutionPrincipalTypeRobot, ID: 30,
		}))
		useRobotController(t).On("Get", mock.Anything, int64(30), mock.Anything).
			Return(&robotCtl.Robot{Robot: robotmodel.Robot{ID: 30, ExpiresAt: 1}}, nil).Once()

		err := retentionCallback(context.Background(), legacyScheduleParam)
		requireRejected(t, ctl, err, "robot 30 is expired")
	})
}
