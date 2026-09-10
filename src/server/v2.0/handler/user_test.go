package handler

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/go-openapi/strfmt"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"

	"github.com/goharbor/harbor/src/common"
	commonmodels "github.com/goharbor/harbor/src/common/models"
	"github.com/goharbor/harbor/src/common/rbac"
	"github.com/goharbor/harbor/src/common/security"
	"github.com/goharbor/harbor/src/common/security/local"
	robotsecurity "github.com/goharbor/harbor/src/common/security/robot"
	"github.com/goharbor/harbor/src/controller/robot"
	"github.com/goharbor/harbor/src/controller/user"
	"github.com/goharbor/harbor/src/pkg/permission/types"
	robotmodel "github.com/goharbor/harbor/src/pkg/robot/model"
	"github.com/goharbor/harbor/src/server/v2.0/models"
	"github.com/goharbor/harbor/src/server/v2.0/restapi"
	operation "github.com/goharbor/harbor/src/server/v2.0/restapi/operations/user"
	usertesting "github.com/goharbor/harbor/src/testing/controller/user"
	"github.com/goharbor/harbor/src/testing/mock"
	htesting "github.com/goharbor/harbor/src/testing/server/v2.0/handler"
)

func TestGetUserOIDCSecret(t *testing.T) {
	const secret = "user-oidc-cli-secret"
	owner := local.NewSecurityContext(&commonmodels.User{UserID: 2, Username: "oidc-user"})
	admin := local.NewSecurityContext(&commonmodels.User{UserID: 1, Username: "admin", SysAdminFlag: true})
	readOnlyRobot := func(id int64) security.Context {
		return robotsecurity.NewSecurityContext(&robot.Robot{
			Robot: robotmodel.Robot{ID: id, Name: "robot$user-reader"},
			Level: robot.LEVELSYSTEM,
			Permissions: []*robot.Permission{{
				Kind:  robot.LEVELSYSTEM,
				Scope: robot.SCOPESYSTEM,
				Access: []*types.Policy{{
					Resource: rbac.ResourceUser,
					Action:   rbac.ActionRead,
					Effect:   types.EffectAllow,
				}},
			}},
		})
	}
	tests := []struct {
		name         string
		caller       security.Context
		authMode     string
		userID       int
		withOIDCInfo bool
		withMeta     bool
		wantSecret   string
	}{
		{"owner", owner, common.OIDCAuth, 2, true, true, secret},
		{"other user administrator", admin, common.OIDCAuth, 2, true, true, "*****"},
		{"system robot with user read", readOnlyRobot(10), common.OIDCAuth, 2, true, true, "*****"},
		{"system robot with matching ID", readOnlyRobot(2), common.OIDCAuth, 2, true, true, "*****"},
		{"missing OIDC metadata", admin, common.OIDCAuth, 2, true, false, ""},
		{"database authentication", admin, common.DBAuth, 2, false, false, ""},
		{"built-in administrator", admin, common.OIDCAuth, 1, false, false, ""},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := security.NewContext(context.Background(), tt.caller)
			us := &commonmodels.User{UserID: tt.userID, Username: "test-user", Email: "test@example.com"}
			if tt.withMeta {
				us.OIDCUserMeta = &commonmodels.OIDCUser{
					ID:           7,
					UserID:       tt.userID,
					SubIss:       "subject-issuer",
					PlainSecret:  secret,
					CreationTime: time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC),
					UpdateTime:   time.Date(2026, 1, 2, 0, 0, 0, 0, time.UTC),
				}
			}
			ctl := &usertesting.Controller{}
			ctl.On("Get", ctx, tt.userID, &user.Option{WithOIDCInfo: tt.withOIDCInfo}).Return(us, nil).Once()
			t.Cleanup(func() { ctl.AssertExpectations(t) })
			api := &usersAPI{
				ctl:     ctl,
				getAuth: func(context.Context) (string, error) { return tt.authMode, nil },
			}

			res := api.GetUser(ctx, operation.GetUserParams{UserID: int64(tt.userID)})
			require.IsType(t, &operation.GetUserOK{}, res)
			payload := res.(*operation.GetUserOK).Payload
			require.NotNil(t, payload)
			assert.Equal(t, int64(tt.userID), payload.UserID)
			assert.Equal(t, us.Username, payload.Username)
			assert.Equal(t, us.Email, payload.Email)
			if !tt.withMeta {
				assert.Nil(t, payload.OIDCUserMeta)
				return
			}
			assert.Equal(t, &models.OIDCUserInfo{
				ID:           us.OIDCUserMeta.ID,
				UserID:       int64(tt.userID),
				Subiss:       us.OIDCUserMeta.SubIss,
				Secret:       tt.wantSecret,
				CreationTime: strfmt.DateTime(us.OIDCUserMeta.CreationTime),
				UpdateTime:   strfmt.DateTime(us.OIDCUserMeta.UpdateTime),
			}, payload.OIDCUserMeta)
			assert.Equal(t, secret, us.OIDCUserMeta.PlainSecret, "redaction must not mutate the controller's user")
		})
	}
}

func TestGetCurrentUserInfoOIDCSecret(t *testing.T) {
	us := &commonmodels.User{
		UserID:   2,
		Username: "oidc-user",
		OIDCUserMeta: &commonmodels.OIDCUser{
			UserID:      2,
			PlainSecret: "own-oidc-cli-secret",
		},
	}
	ctx := security.NewContext(context.Background(), local.NewSecurityContext(us))
	ctl := &usertesting.Controller{}
	ctl.On("Get", ctx, us.UserID, &user.Option{WithOIDCInfo: true}).Return(us, nil).Once()
	t.Cleanup(func() { ctl.AssertExpectations(t) })
	api := &usersAPI{
		ctl:     ctl,
		getAuth: func(context.Context) (string, error) { return common.OIDCAuth, nil },
	}

	res := api.GetCurrentUserInfo(ctx, operation.GetCurrentUserInfoParams{})
	require.IsType(t, &operation.GetCurrentUserInfoOK{}, res)
	payload := res.(*operation.GetCurrentUserInfoOK).Payload
	require.NotNil(t, payload)
	require.NotNil(t, payload.OIDCUserMeta)
	assert.Equal(t, int64(us.UserID), payload.UserID)
	assert.Equal(t, "own-oidc-cli-secret", payload.OIDCUserMeta.Secret)
}

func TestRequireValidSecret(t *testing.T) {
	cases := []struct {
		in       string
		hasError bool
	}{
		{"", true},
		{"12345678", true},
		{"passw0rd", true},
		{"PASSW0RD", true},
		{"Sh0rt", true},
		{"Passw0rd", false},
		{"Thisis1Valid_password", false},
		// secret of length 128 characters long should be ok, no error returned
		{"abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789abcd", false},
		// secret of length larger than 128 characters long, such as 129 characters long, should return error
		{"abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789abcde", true},
	}
	for _, c := range cases {
		e := requireValidSecret(c.in)
		assert.Equal(t, c.hasError, e != nil)
	}
}

type UserTestSuite struct {
	htesting.Suite
	uCtl *usertesting.Controller

	user *commonmodels.User
}

func (uts *UserTestSuite) SetupSuite() {
	uts.user = &commonmodels.User{
		UserID:   1,
		Username: "admin",
	}

	uts.uCtl = &usertesting.Controller{}
	uts.Config = &restapi.Config{
		UserAPI: &usersAPI{
			ctl: uts.uCtl,
			getAuth: func(ctx context.Context) (string, error) {
				return common.DBAuth, nil
			},
		},
	}

	uts.Suite.SetupSuite()
	uts.Security.On("IsAuthenticated").Return(true)
}

func (uts *UserTestSuite) TestUpdateUserPassword() {

	body := models.PasswordReq{
		OldPassword: "Harbor12345",
		NewPassword: "Passw0rd",
	}
	{
		url := "/users/2/password"
		uts.Security.On("Can", mock.Anything, mock.Anything, mock.Anything).Return(false).Times(1)
		uts.Security.On("GetUsername").Return("testuser")
		res, err := uts.Suite.PutJSON(url, &body)
		uts.NoError(err)
		uts.Equal(403, res.StatusCode)
	}
	{
		url := "/users/1/password"
		uts.Security.On("Can", mock.Anything, mock.Anything, mock.Anything).Return(true).Times(1)
		uts.Security.On("GetUsername").Return("admin")

		uts.uCtl.On("Get", mock.Anything, mock.Anything, mock.Anything).Return(uts.user, nil).Times(1)
		uts.uCtl.On("VerifyPassword", mock.Anything, "admin", "Passw0rd").Return(true, nil).Times(1)
		res, err := uts.Suite.PutJSON(url, &body)
		uts.NoError(err)
		uts.Equal(400, res.StatusCode)
	}
	{
		url := "/users/1/password"
		uts.Security.On("Can", mock.Anything, mock.Anything, mock.Anything).Return(true).Times(1)

		uts.uCtl.On("Get", mock.Anything, mock.Anything, mock.Anything).Return(uts.user, nil).Times(1)
		uts.uCtl.On("VerifyPassword", mock.Anything, "admin", mock.Anything).Return(false, nil).Times(1)
		uts.uCtl.On("UpdatePassword", mock.Anything, mock.Anything, mock.Anything).Return(nil)
		res, err := uts.Suite.PutJSON(url, &body)
		uts.NoError(err)
		uts.Equal(200, res.StatusCode)
	}
}

func (uts *UserTestSuite) TestGetRandomSecret() {
	for i := 1; i < 5; i++ {
		rSec, err := getRandomSecret()
		uts.NoError(err)
		uts.NoError(requireValidSecret(rSec))
	}
}

func TestUserTestSuite(t *testing.T) {
	suite.Run(t, &UserTestSuite{})
}

func Test_validateUserProfile(t *testing.T) {
	tooLongUsername := "mike012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789mike012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789mike012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789"
	type args struct {
		user   *commonmodels.User
		create bool
	}
	tests := []struct {
		name    string
		args    args
		wantErr assert.ErrorAssertionFunc
	}{
		{"normal_test", args{&commonmodels.User{Username: "mike", Realname: "mike", Email: "mike@example.com"}, true}, assert.NoError},
		{"illegall_username_,", args{&commonmodels.User{Username: "mike,mike", Realname: "mike", Email: "mike@example.com"}, true}, assert.Error},
		{"illegall_username_$", args{&commonmodels.User{Username: "mike$mike", Realname: "mike", Email: "mike@example.com"}, true}, assert.Error},
		{"illegall_username_%", args{&commonmodels.User{Username: "mike%mike", Realname: "mike", Email: "mike@example.com"}, true}, assert.Error},
		{"illegall_username_#", args{&commonmodels.User{Username: "mike#mike", Realname: "mike", Email: "mike@example.com"}, true}, assert.Error},
		{"illegall_realname", args{&commonmodels.User{Username: "mike", Realname: "mike,mike", Email: "mike@example.com"}, true}, assert.Error},
		{"update_profile", args{&commonmodels.User{Username: "", Realname: "mike", Email: "mike@example.com"}, false}, assert.NoError},
		{"username_too_long", args{&commonmodels.User{Username: tooLongUsername, Realname: "mike", Email: "mike@example.com"}, true}, assert.Error},
		{"invalid_email", args{&commonmodels.User{Username: "mike", Realname: "mike", Email: "mike#example.com"}, true}, assert.Error},
		{"invalid_comment", args{&commonmodels.User{Username: "mike", Realname: "mike", Email: "mike@example.com", Comment: tooLongUsername}, true}, assert.Error},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.wantErr(t, validateUserProfile(tt.args.user, tt.args.create), fmt.Sprintf("validateUserProfile(%v)", tt.args.user))
		})
	}
}
