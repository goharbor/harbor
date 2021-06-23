package handler

import (
	"context"
	"testing"

	"github.com/goharbor/harbor/src/common"
	"github.com/goharbor/harbor/src/server/v2.0/models"
	"github.com/goharbor/harbor/src/server/v2.0/restapi"
	usertesting "github.com/goharbor/harbor/src/testing/controller/user"
	"github.com/goharbor/harbor/src/testing/mock"
	htesting "github.com/goharbor/harbor/src/testing/server/v2.0/handler"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
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
	}
	for _, c := range cases {
		e := requireValidSecret(c.in)
		assert.Equal(t, c.hasError, e != nil)
	}
}

type UserTestSuite struct {
	htesting.Suite
	uCtl *usertesting.Controller
}

func (uts *UserTestSuite) SetupSuite() {
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
		uts.Security.On("GetUsername").Return("admin").Times(1)

		uts.uCtl.On("VerifyPassword", mock.Anything, "admin", "Passw0rd").Return(true, nil).Times(1)
		res, err := uts.Suite.PutJSON(url, &body)
		uts.NoError(err)
		uts.Equal(400, res.StatusCode)
	}
	{
		url := "/users/1/password"
		uts.Security.On("Can", mock.Anything, mock.Anything, mock.Anything).Return(true).Times(1)
		uts.Security.On("GetUsername").Return("admin").Times(1)

		uts.uCtl.On("VerifyPassword", mock.Anything, "admin", mock.Anything).Return(false, nil).Times(1)
		uts.uCtl.On("UpdatePassword", mock.Anything, mock.Anything, mock.Anything).Return(nil)
		res, err := uts.Suite.PutJSON(url, &body)
		uts.NoError(err)
		uts.Equal(200, res.StatusCode)
	}
}

func TestUserTestSuite(t *testing.T) {
	suite.Run(t, &UserTestSuite{})
}
