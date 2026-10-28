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

package handler

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"testing"
	"time"

	gooidc "github.com/coreos/go-oidc/v3/oidc"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"golang.org/x/oauth2"

	"github.com/goharbor/harbor/src/common"
	commonmodels "github.com/goharbor/harbor/src/common/models"
	"github.com/goharbor/harbor/src/controller/event/metadata"
	"github.com/goharbor/harbor/src/controller/user"
	"github.com/goharbor/harbor/src/lib/config"
	cfgModels "github.com/goharbor/harbor/src/lib/config/models"
	"github.com/goharbor/harbor/src/lib/encrypt"
	"github.com/goharbor/harbor/src/lib/errors"
	_ "github.com/goharbor/harbor/src/pkg/config/inmemory"
	notifierevent "github.com/goharbor/harbor/src/pkg/notifier/event"
	oidcpkg "github.com/goharbor/harbor/src/pkg/oidc"
	"github.com/goharbor/harbor/src/server/v2.0/models"
	"github.com/goharbor/harbor/src/server/v2.0/restapi"
	usertesting "github.com/goharbor/harbor/src/testing/controller/user"
	htesting "github.com/goharbor/harbor/src/testing/server/v2.0/handler"
)

const (
	testIDToken = "header.payload.signature"
	testIssuer  = "https://dex.example.com"
	testSubject = "repo:my-org/my-repo:ref:refs/heads/main"
)

type OIDCOnboardTestSuite struct {
	htesting.Suite
	api  *oidcAPI
	uCtl *usertesting.Controller

	authMode      string
	autoOnboard   bool
	expiry        time.Time
	claimUsername string
	events        []notifierevent.Metadata
}

func (s *OIDCOnboardTestSuite) SetupSuite() {
	config.InitWithSettings(map[string]any{}, &encrypt.PresetKeyProvider{Key: "naa4JtarA1Zsc3uY"})
	s.api = &oidcAPI{}
	s.Config = &restapi.Config{OIDCAPI: s.api}
	s.Suite.SetupSuite()
}

func (s *OIDCOnboardTestSuite) SetupTest() {
	s.authMode = common.OIDCAuth
	s.autoOnboard = false
	s.claimUsername = "My Repo"
	s.expiry = time.Now().Add(time.Hour).Round(0)
	s.uCtl = &usertesting.Controller{}
	s.api.userCtl = s.uCtl
	s.api.getAuth = func(_ context.Context) (string, error) {
		return s.authMode, nil
	}
	s.api.oidcSetting = func(_ context.Context) (*cfgModels.OIDCSetting, error) {
		return &cfgModels.OIDCSetting{AutoOnboard: s.autoOnboard, UserClaim: "name"}, nil
	}
	s.api.verifyToken = func(_ context.Context, rawIDToken string) (*oidcpkg.Token, *oidcpkg.UserInfo, error) {
		if rawIDToken != testIDToken {
			return nil, nil, errors.UnauthorizedError(nil).WithMessage("invalid ID token")
		}
		token := &oidcpkg.Token{
			Token:      oauth2.Token{AccessToken: rawIDToken, TokenType: "Bearer", Expiry: s.expiry},
			RawIDToken: rawIDToken,
		}
		info := &oidcpkg.UserInfo{Issuer: testIssuer, Subject: testSubject, Username: s.claimUsername}
		return token, info, nil
	}
	s.api.injectGroups = func(_ *oidcpkg.UserInfo, _ *commonmodels.User) {}
	s.events = nil
	s.api.addEvent = func(_ context.Context, m notifierevent.Metadata) {
		s.events = append(s.events, m)
	}
}

func (s *OIDCOnboardTestSuite) TearDownTest() {
	s.uCtl.AssertExpectations(s.T())
}

func (s *OIDCOnboardTestSuite) onboard(header string, body any) *http.Response {
	var reader io.Reader
	if body != nil {
		b, err := json.Marshal(body)
		s.Require().NoError(err)
		reader = bytes.NewReader(b)
	}
	headers := map[string]string{}
	if header != "" {
		headers["Authorization"] = header
	}
	res, err := s.Post("/oidc/onboard", reader, headers)
	s.Require().NoError(err)
	return res
}

func (s *OIDCOnboardTestSuite) decodeUser(res *http.Response) *models.UserResp {
	defer res.Body.Close()
	u := &models.UserResp{}
	s.Require().NoError(json.NewDecoder(res.Body).Decode(u))
	return u
}

func (s *OIDCOnboardTestSuite) storedToken(refreshToken string, expiry time.Time) string {
	tb, err := json.Marshal(&oidcpkg.Token{Token: oauth2.Token{AccessToken: "stored", RefreshToken: refreshToken, Expiry: expiry}})
	s.Require().NoError(err)
	_, t, err := oidcpkg.EncryptSecretAndToken(tb)
	s.Require().NoError(err)
	return t
}

func (s *OIDCOnboardTestSuite) mockNotOnboarded() {
	s.uCtl.On("GetBySubIss", mock.Anything, testSubject, testIssuer).Return(nil, errors.NotFoundError(nil)).Once()
}

func (s *OIDCOnboardTestSuite) mockOnboard(username string, err error) {
	s.uCtl.On("OnboardOIDCUser", mock.Anything, mock.MatchedBy(func(u *commonmodels.User) bool {
		if u.Username != username || u.OIDCUserMeta == nil || u.OIDCUserMeta.SubIss != testSubject+testIssuer {
			return false
		}
		token, err := oidcpkg.DecryptToken(u.OIDCUserMeta.Token)
		return err == nil && token.RawIDToken == testIDToken && token.RefreshToken == "" && token.Expiry.Equal(s.expiry)
	})).Run(func(args mock.Arguments) {
		args.Get(1).(*commonmodels.User).UserID = 3
	}).Return(err).Once()
	if err == nil {
		s.uCtl.On("Get", mock.Anything, 3, (*user.Option)(nil)).Return(&commonmodels.User{UserID: 3, Username: username}, nil).Once()
	}
}

func (s *OIDCOnboardTestSuite) TestNotOIDCAuthMode() {
	s.authMode = common.DBAuth
	res := s.onboard("Bearer "+testIDToken, nil)
	s.Equal(http.StatusPreconditionFailed, res.StatusCode)
}

func (s *OIDCOnboardTestSuite) TestMissingToken() {
	res := s.onboard("", &models.OIDCOnboardReq{Username: "ci-user"})
	s.Equal(http.StatusUnauthorized, res.StatusCode)

	res = s.onboard("Basic dXNlcjpwYXNz", &models.OIDCOnboardReq{Username: "ci-user"})
	s.Equal(http.StatusUnauthorized, res.StatusCode)
}

func (s *OIDCOnboardTestSuite) TestInvalidToken() {
	res := s.onboard("Bearer invalid", &models.OIDCOnboardReq{Username: "ci-user"})
	s.Equal(http.StatusUnauthorized, res.StatusCode)
	s.Empty(s.events)
}

func (s *OIDCOnboardTestSuite) TestOnboardWithUsername() {
	s.mockNotOnboarded()
	s.mockOnboard("ci-user", nil)
	res := s.onboard("bearer "+testIDToken, &models.OIDCOnboardReq{Username: "ci-user"})
	s.Require().Equal(http.StatusOK, res.StatusCode)
	u := s.decodeUser(res)
	s.Equal(int64(3), u.UserID)
	s.Equal("ci-user", u.Username)
	s.Nil(u.OIDCUserMeta)
	s.Equal([]notifierevent.Metadata{&metadata.OIDCOnboardEventMetadata{Username: "ci-user"}}, s.events)
}

func (s *OIDCOnboardTestSuite) TestOnboardWithInvalidUsername() {
	s.mockNotOnboarded()
	res := s.onboard("Bearer "+testIDToken, &models.OIDCOnboardReq{Username: "robot$ci"})
	s.Equal(http.StatusBadRequest, res.StatusCode)
}

// conflictErr is the error of the user controller when the username, or the OIDC identity, is already onboarded
var conflictErr = errors.Wrap(errors.ConflictError(nil).WithMessage("user ci-user or email  already exists"), "failed to create user record")

func (s *OIDCOnboardTestSuite) TestOnboardUsernameTaken() {
	s.mockNotOnboarded()
	s.mockOnboard("ci-user", conflictErr)
	// the conflict is not from a concurrent onboarding of the same user
	s.mockNotOnboarded()
	res := s.onboard("Bearer "+testIDToken, &models.OIDCOnboardReq{Username: "ci-user"})
	s.Equal(http.StatusConflict, res.StatusCode)
	s.Empty(s.events)
}

func (s *OIDCOnboardTestSuite) TestConcurrentOnboard() {
	s.mockNotOnboarded()
	s.mockOnboard("ci-user", conflictErr)
	// the same user got onboarded by a concurrent request
	existing := &commonmodels.User{UserID: 4, Username: "ci-user"}
	s.uCtl.On("GetBySubIss", mock.Anything, testSubject, testIssuer).Return(existing, nil).Once()
	s.uCtl.On("Get", mock.Anything, 4, &user.Option{WithOIDCInfo: true}).Return(&commonmodels.User{
		UserID:       4,
		Username:     "ci-user",
		OIDCUserMeta: &commonmodels.OIDCUser{ID: 8, UserID: 4, Token: s.storedToken("", time.Time{})},
	}, nil).Once()
	s.uCtl.On("UpdateOIDCMeta", mock.Anything, mock.MatchedBy(func(meta *commonmodels.OIDCUser) bool {
		return meta.ID == 8
	}), "token").Return(nil).Once()
	s.uCtl.On("Get", mock.Anything, 4, (*user.Option)(nil)).Return(existing, nil).Once()

	res := s.onboard("Bearer "+testIDToken, &models.OIDCOnboardReq{Username: "ci-user"})
	s.Require().Equal(http.StatusOK, res.StatusCode)
	s.Equal(int64(4), s.decodeUser(res).UserID)
	s.Equal([]notifierevent.Metadata{&metadata.OIDCOnboardEventMetadata{Username: "ci-user"}}, s.events)
}

func (s *OIDCOnboardTestSuite) TestAutoOnboardWithoutUsername() {
	s.autoOnboard = true
	s.mockNotOnboarded()
	s.mockOnboard("My_Repo", nil)
	res := s.onboard("Bearer "+testIDToken, nil)
	s.Require().Equal(http.StatusOK, res.StatusCode)
	s.Equal("My_Repo", s.decodeUser(res).Username)
}

func (s *OIDCOnboardTestSuite) TestAutoOnboardReplacesColons() {
	s.autoOnboard = true
	s.claimUsername = testSubject
	s.mockNotOnboarded()
	s.mockOnboard("repo.my-org/my-repo.ref.refs/heads/main", nil)
	res := s.onboard("Bearer "+testIDToken, nil)
	s.Require().Equal(http.StatusOK, res.StatusCode)
	s.Equal("repo.my-org/my-repo.ref.refs/heads/main", s.decodeUser(res).Username)
}

func (s *OIDCOnboardTestSuite) TestOnboardWithColonInUsername() {
	s.mockNotOnboarded()
	res := s.onboard("Bearer "+testIDToken, &models.OIDCOnboardReq{Username: "ci:user"})
	s.Equal(http.StatusBadRequest, res.StatusCode)
}

func (s *OIDCOnboardTestSuite) TestWithoutUsernameAutoOnboardDisabled() {
	s.mockNotOnboarded()
	res := s.onboard("Bearer "+testIDToken, &models.OIDCOnboardReq{})
	s.Equal(http.StatusBadRequest, res.StatusCode)
}

func (s *OIDCOnboardTestSuite) TestAlreadyOnboardedReplacesToken() {
	existing := &commonmodels.User{UserID: 3, Username: "ci-user"}
	s.uCtl.On("GetBySubIss", mock.Anything, testSubject, testIssuer).Return(existing, nil).Once()
	s.uCtl.On("Get", mock.Anything, 3, &user.Option{WithOIDCInfo: true}).Return(&commonmodels.User{
		UserID:       3,
		Username:     "ci-user",
		OIDCUserMeta: &commonmodels.OIDCUser{ID: 7, UserID: 3, Token: s.storedToken("", s.expiry.Add(-time.Minute))},
	}, nil).Once()
	s.uCtl.On("UpdateOIDCMeta", mock.Anything, mock.MatchedBy(func(meta *commonmodels.OIDCUser) bool {
		token, err := oidcpkg.DecryptToken(meta.Token)
		return meta.ID == 7 && err == nil && token.RawIDToken == testIDToken && token.Expiry.Equal(s.expiry)
	}), "token").Return(nil).Once()
	s.uCtl.On("Get", mock.Anything, 3, (*user.Option)(nil)).Return(&commonmodels.User{UserID: 3, Username: "ci-user"}, nil).Once()

	// the username in the request is ignored for users already onboarded
	res := s.onboard("Bearer "+testIDToken, &models.OIDCOnboardReq{Username: "another-user"})
	s.Require().Equal(http.StatusOK, res.StatusCode)
	s.Equal("ci-user", s.decodeUser(res).Username)
	s.Equal([]notifierevent.Metadata{&metadata.OIDCOnboardEventMetadata{Username: "ci-user"}}, s.events)
}

func (s *OIDCOnboardTestSuite) TestAlreadyOnboardedKeepsRefreshableToken() {
	existing := &commonmodels.User{UserID: 3, Username: "jane"}
	s.uCtl.On("GetBySubIss", mock.Anything, testSubject, testIssuer).Return(existing, nil).Once()
	s.uCtl.On("Get", mock.Anything, 3, &user.Option{WithOIDCInfo: true}).Return(&commonmodels.User{
		UserID:       3,
		Username:     "jane",
		OIDCUserMeta: &commonmodels.OIDCUser{ID: 7, UserID: 3, Token: s.storedToken("refresh-token", time.Time{})},
	}, nil).Once()
	s.uCtl.On("Get", mock.Anything, 3, (*user.Option)(nil)).Return(&commonmodels.User{UserID: 3, Username: "jane"}, nil).Once()

	res := s.onboard("Bearer "+testIDToken, nil)
	s.Require().Equal(http.StatusOK, res.StatusCode)
	s.Equal("jane", s.decodeUser(res).Username)
	s.uCtl.AssertNotCalled(s.T(), "UpdateOIDCMeta", mock.Anything, mock.Anything, mock.Anything)
}

func (s *OIDCOnboardTestSuite) TestAlreadyOnboardedKeepsLaterExpiringToken() {
	existing := &commonmodels.User{UserID: 3, Username: "ci-user"}
	s.uCtl.On("GetBySubIss", mock.Anything, testSubject, testIssuer).Return(existing, nil).Once()
	s.uCtl.On("Get", mock.Anything, 3, &user.Option{WithOIDCInfo: true}).Return(&commonmodels.User{
		UserID:       3,
		Username:     "ci-user",
		OIDCUserMeta: &commonmodels.OIDCUser{ID: 7, UserID: 3, Token: s.storedToken("", s.expiry.Add(time.Minute))},
	}, nil).Once()
	s.uCtl.On("Get", mock.Anything, 3, (*user.Option)(nil)).Return(existing, nil).Once()

	res := s.onboard("Bearer "+testIDToken, nil)
	s.Require().Equal(http.StatusOK, res.StatusCode)
	s.uCtl.AssertNotCalled(s.T(), "UpdateOIDCMeta", mock.Anything, mock.Anything, mock.Anything)
}

func TestOIDCOnboardTestSuite(t *testing.T) {
	suite.Run(t, &OIDCOnboardTestSuite{})
}

func TestVerifyTokenError(t *testing.T) {
	cases := []struct {
		name    string
		err     error
		code    string
		message string
	}{
		{
			name:    "expired token",
			err:     errors.UnauthorizedError(&gooidc.TokenExpiredError{Expiry: time.Now()}).WithMessage("invalid ID token"),
			code:    errors.UnAuthorizedCode,
			message: "the token is expired",
		},
		{
			name:    "invalid token",
			err:     errors.UnauthorizedError(fmt.Errorf(`oidc: expected audience "harbor" got ["other"]`)).WithMessage("invalid ID token"),
			code:    errors.UnAuthorizedCode,
			message: "invalid token",
		},
		{
			name: "provider not reachable",
			err:  fmt.Errorf("failed to create OIDC provider, error: connection refused"),
			code: errors.GeneralCode,
		},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			err := convertVerifyTokenError(context.Background(), c.err)
			assert.Equal(t, c.code, errors.ErrCode(err))
			if c.message != "" {
				var e *errors.Error
				require.True(t, errors.As(err, &e))
				assert.Equal(t, c.message, e.Error())
			}
		})
	}
}
