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
	"errors"
	"testing"

	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/goharbor/harbor/src/common"
	"github.com/goharbor/harbor/src/common/models"
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

const legacyScheduleParam = `{"PolicyID":2,"Trigger":"Schedule","Operator":"harbor-jobservice"}`

type callbackController struct {
	Controller

	metadata       *policy.Metadata
	triggerCalls   int
	triggerContext context.Context
}

func (c *callbackController) GetRetention(context.Context, int64) (*policy.Metadata, error) {
	return c.metadata, nil
}

func (c *callbackController) TriggerRetentionExec(
	ctx context.Context,
	_ int64,
	_ string,
	_ bool,
) (int64, error) {
	c.triggerCalls++
	c.triggerContext = ctx
	return 1, nil
}

func callbackMetadata(t *testing.T, principal string) *policy.Metadata {
	t.Helper()
	raw := `{
        "algorithm":"or",
        "rules":[],
        "trigger":{"kind":"Schedule","settings":{"cron":"0 0 0 * * *"}},
        "scope":{"level":"project","ref":1}` + principal + `}`
	metadata := &policy.Metadata{}
	require.NoError(t, json.Unmarshal([]byte(raw), metadata))
	return metadata
}

func metadataWithPrincipal(
	t *testing.T,
	metadata *policy.Metadata,
	principalType string,
	principalID int64,
) *policy.Metadata {
	t.Helper()
	raw, err := json.Marshal(metadata)
	require.NoError(t, err)
	data := map[string]any{}
	require.NoError(t, json.Unmarshal(raw, &data))
	data["execution_principal"] = map[string]any{
		"type": principalType,
		"id":   principalID,
	}
	raw, err = json.Marshal(data)
	require.NoError(t, err)
	metadata = &policy.Metadata{}
	require.NoError(t, json.Unmarshal(raw, metadata))
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

func TestRetentionCallbackRejectsLegacySchedule(t *testing.T) {
	ctl := useCallbackController(t, callbackMetadata(t, ""))

	err := retentionCallback(context.Background(), legacyScheduleParam)
	require.ErrorContains(t, err, "no authorized execution principal")
	require.Zero(t, ctl.triggerCalls)
}

func TestRetentionCallbackRevalidatesLocalPrincipal(t *testing.T) {
	metadata := metadataWithPrincipal(
		t,
		policy.WithNDaysSinceLastPull(1, 7),
		"local",
		17,
	)
	ctl := useCallbackController(t, metadata)

	originalUsers := user.Mgr
	users := &usertesting.Manager{}
	user.Mgr = users
	t.Cleanup(func() {
		user.Mgr = originalUsers
	})
	users.On("Get", mock.Anything, 17).Return(&models.User{
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

func TestRetentionCallbackRejectsDeveloperPrincipal(t *testing.T) {
	metadata := callbackMetadata(
		t,
		`,"execution_principal":{"type":"local","id":18}`,
	)
	ctl := useCallbackController(t, metadata)

	originalUsers := user.Mgr
	users := &usertesting.Manager{}
	user.Mgr = users
	t.Cleanup(func() {
		user.Mgr = originalUsers
	})
	developer := &models.User{UserID: 18, Username: "developer"}
	users.On("Get", mock.Anything, 18).Return(developer, nil).Once()

	originalProjects := projectCtl.Ctl
	projects := &projecttesting.Controller{}
	projectCtl.Ctl = projects
	t.Cleanup(func() {
		projectCtl.Ctl = originalProjects
	})
	projects.On("Get", mock.Anything, int64(1), mock.Anything).
		Return(&projectmodels.Project{ProjectID: 1}, nil).Once()
	projects.On("ListRoles", mock.Anything, int64(1), developer).
		Return([]int{common.RoleDeveloper}, nil).Once()
	err := retentionCallback(context.Background(), legacyScheduleParam)
	require.ErrorContains(t, err, "cannot delete project artifacts")
	require.Zero(t, ctl.triggerCalls)
}

func TestRetentionCallbackRejectsUnavailablePrincipals(t *testing.T) {
	t.Run("invalid principal ID", func(t *testing.T) {
		metadata := callbackMetadata(
			t,
			`,"execution_principal":{"type":"local","id":0}`,
		)
		ctl := useCallbackController(t, metadata)

		err := retentionCallback(context.Background(), legacyScheduleParam)
		require.ErrorContains(t, err, "invalid execution principal")
		require.Zero(t, ctl.triggerCalls)
	})

	t.Run("deleted local user", func(t *testing.T) {
		metadata := callbackMetadata(
			t,
			`,"execution_principal":{"type":"local","id":19}`,
		)
		ctl := useCallbackController(t, metadata)

		originalUsers := user.Mgr
		users := &usertesting.Manager{}
		user.Mgr = users
		t.Cleanup(func() {
			user.Mgr = originalUsers
		})
		users.On("Get", mock.Anything, 19).
			Return(nil, errors.New("user not found")).Once()

		err := retentionCallback(context.Background(), legacyScheduleParam)
		require.ErrorContains(t, err, "user not found")
		require.Zero(t, ctl.triggerCalls)
	})

	t.Run("deleted local account", func(t *testing.T) {
		metadata := callbackMetadata(
			t,
			`,"execution_principal":{"type":"local","id":20}`,
		)
		ctl := useCallbackController(t, metadata)

		originalUsers := user.Mgr
		users := &usertesting.Manager{}
		user.Mgr = users
		t.Cleanup(func() {
			user.Mgr = originalUsers
		})
		users.On("Get", mock.Anything, 20).Return(&models.User{
			UserID:   20,
			Username: "deleted-user",
			Deleted:  true,
		}, nil).Once()

		err := retentionCallback(context.Background(), legacyScheduleParam)
		require.ErrorContains(t, err, "local user 20 is deleted")
		require.Zero(t, ctl.triggerCalls)
	})

	t.Run("disabled robot", func(t *testing.T) {
		metadata := callbackMetadata(
			t,
			`,"execution_principal":{"type":"robot","id":29}`,
		)
		ctl := useCallbackController(t, metadata)

		originalRobots := robotCtl.Ctl
		robots := &robottesting.Controller{}
		robotCtl.Ctl = robots
		t.Cleanup(func() {
			robotCtl.Ctl = originalRobots
		})
		robots.On("Get", mock.Anything, int64(29), mock.Anything).
			Return(&robotCtl.Robot{Robot: robotmodel.Robot{
				ID:       29,
				Disabled: true,
			}}, nil).Once()

		err := retentionCallback(context.Background(), legacyScheduleParam)
		require.ErrorContains(t, err, "robot 29 is disabled")
		require.Zero(t, ctl.triggerCalls)
	})

	t.Run("expired robot", func(t *testing.T) {
		metadata := callbackMetadata(
			t,
			`,"execution_principal":{"type":"robot","id":30}`,
		)
		ctl := useCallbackController(t, metadata)

		originalRobots := robotCtl.Ctl
		robots := &robottesting.Controller{}
		robotCtl.Ctl = robots
		t.Cleanup(func() {
			robotCtl.Ctl = originalRobots
		})
		robots.On("Get", mock.Anything, int64(30), mock.Anything).
			Return(&robotCtl.Robot{Robot: robotmodel.Robot{
				ID:        30,
				ExpiresAt: 1,
			}}, nil).Once()

		err := retentionCallback(context.Background(), legacyScheduleParam)
		require.ErrorContains(t, err, "robot 30 is expired")
		require.Zero(t, ctl.triggerCalls)
	})
}
