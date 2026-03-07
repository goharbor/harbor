package handler

import (
	"context"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"

	"github.com/goharbor/harbor/src/common"
	commonmodels "github.com/goharbor/harbor/src/common/models"
	"github.com/goharbor/harbor/src/server/v2.0/models"
	"github.com/goharbor/harbor/src/server/v2.0/restapi"
	usertesting "github.com/goharbor/harbor/src/testing/controller/user"
	"github.com/goharbor/harbor/src/testing/mock"
	htesting "github.com/goharbor/harbor/src/testing/server/v2.0/handler"
)

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
		res, err := uts.Suite.PutJSON(url, &body)
		uts.NoError(err)
		uts.Equal(403, res.StatusCode)
	}
	{
		url := "/users/1/password"
		uts.Security.On("Can", mock.Anything, mock.Anything, mock.Anything).Return(true).Times(1)

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
