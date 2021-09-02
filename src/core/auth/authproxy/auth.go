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

package authproxy

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/goharbor/harbor/src/common"
	"github.com/goharbor/harbor/src/common/models"
	"github.com/goharbor/harbor/src/controller/usergroup"
	"github.com/goharbor/harbor/src/core/auth"
	"github.com/goharbor/harbor/src/jobservice/logger"
	"github.com/goharbor/harbor/src/lib/config"
	cfgModels "github.com/goharbor/harbor/src/lib/config/models"
	harborErrors "github.com/goharbor/harbor/src/lib/errors"
	"github.com/goharbor/harbor/src/lib/log"
	"github.com/goharbor/harbor/src/pkg/authproxy"
	"github.com/goharbor/harbor/src/pkg/user"
	"github.com/goharbor/harbor/src/pkg/usergroup/model"
)

const refreshDuration = 2 * time.Second
const userEntryComment = "By Authproxy"

var transport = &http.Transport{
	Proxy: http.ProxyFromEnvironment,
}

// Auth implements HTTP authenticator the required attributes.
// The attribute Endpoint is the HTTP endpoint to which the POST request should be issued for authentication
type Auth struct {
	auth.DefaultAuthenticateHelper
	sync.Mutex
	Endpoint            string
	TokenReviewEndpoint string
	SkipSearch          bool
	settingTimeStamp    time.Time
	client              *http.Client
	userMgr             user.Manager
}

type session struct {
	SessionID string `json:"session_id,omitempty"`
}

// Authenticate issues http POST request to Endpoint if it returns 200 the authentication is considered success.
func (a *Auth) Authenticate(ctx context.Context, m models.AuthModel) (*models.User, error) {
	err := a.ensure(ctx)
	if err != nil {
		if a.Endpoint == "" {
			return nil, fmt.Errorf("failed to initialize HTTP Auth Proxy Authenticator, error: %v", err)
		}
		log.Warningf("Failed to refresh configuration for HTTP Auth Proxy Authenticator, error: %v, old settings will be used", err)
	}
	req, err := http.NewRequest(http.MethodPost, a.Endpoint, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to send request, error: %v", err)
	}
	req.SetBasicAuth(m.Principal, m.Password)
	resp, err := a.client.Do(req)
	if err != nil {
		return nil, err
	}
	data, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Warningf("Failed to read response body, error: %v", err)
		return nil, auth.ErrAuth{}
	}
	defer resp.Body.Close()
	if resp.StatusCode == http.StatusOK {
		s := session{}
		err = json.Unmarshal(data, &s)
		if err != nil {
			return nil, auth.NewErrAuth(fmt.Sprintf("failed to read session %v", err))
		}
		user, err := a.tokenReview(ctx, s.SessionID)
		if err != nil {
			return nil, auth.NewErrAuth(fmt.Sprintf("failed to do token review, error: %v", err))
		}
		return user, nil
	} else if resp.StatusCode == http.StatusUnauthorized {
		return nil, auth.NewErrAuth(string(data))
	} else {
		data, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			log.Warningf("Failed to read response body, error: %v", err)
		}
		return nil, fmt.Errorf("failed to authenticate, status code: %d, text: %s", resp.StatusCode, string(data))
	}
}

func (a *Auth) tokenReview(ctx context.Context, sessionID string) (*models.User, error) {
	httpAuthProxySetting, err := config.HTTPAuthProxySetting(ctx)
	if err != nil {
		return nil, err
	}
	reviewStatus, err := authproxy.TokenReview(sessionID, httpAuthProxySetting)
	if err != nil {
		return nil, err
	}
	u, err := authproxy.UserFromReviewStatus(reviewStatus, httpAuthProxySetting.AdminGroups, httpAuthProxySetting.AdminUsernames)
	if err != nil {
		return nil, err
	}
	return u, nil
}

// VerifyToken reviews the token to generate the user model
func (a *Auth) VerifyToken(ctx context.Context, token string) (*models.User, error) {
	if err := a.ensure(ctx); err != nil {
		return nil, err
	}
	return a.tokenReview(ctx, token)
}

// OnBoardUser delegates to dao pkg to insert/update data in DB.
func (a *Auth) OnBoardUser(ctx context.Context, u *models.User) error {
	return a.userMgr.Onboard(ctx, u)
}

// PostAuthenticate generates the user model and on board the user.
func (a *Auth) PostAuthenticate(ctx context.Context, u *models.User) error {
	_, err := a.userMgr.GetByName(ctx, u.Username)
	if harborErrors.IsNotFoundErr(err) {
		if err2 := a.fillInModel(u); err2 != nil {
			return err2
		}
		return a.OnBoardUser(ctx, u)
	} else if err != nil {
		return err
	}
	// do nothing if user exists in DB
	return nil
}

// SearchUser returns nil as authproxy does not have such capability.
// When SkipSearch is set it always return the default model,
// the username will be switch to lowercase if it's configured as case-insensitive
func (a *Auth) SearchUser(ctx context.Context, username string) (*models.User, error) {
	err := a.ensure(ctx)
	if err != nil {
		log.Warningf("Failed to refresh configuration for HTTP Auth Proxy Authenticator, error: %v, the default settings will be used", err)
	}
	var u *models.User
	if a.SkipSearch {
		u = &models.User{Username: username}
		if err := a.fillInModel(u); err != nil {
			return nil, err
		}
	}
	return u, nil
}

// SearchGroup search group exist in the authentication provider, for HTTP auth, if SkipSearch is true, it assume this group exist in authentication provider.
func (a *Auth) SearchGroup(ctx context.Context, groupKey string) (*model.UserGroup, error) {
	err := a.ensure(ctx)
	if err != nil {
		log.Warningf("Failed to refresh configuration for HTTP Auth Proxy Authenticator, error: %v, the default settings will be used", err)
	}
	var ug *model.UserGroup
	if a.SkipSearch {
		ug = &model.UserGroup{
			GroupName: groupKey,
			GroupType: common.HTTPGroupType,
		}
		return ug, nil
	}
	return nil, nil
}

// OnBoardGroup create user group entity in Harbor DB, altGroupName is not used.
func (a *Auth) OnBoardGroup(ctx context.Context, u *model.UserGroup, altGroupName string) error {
	// if group name provided, on board the user group
	if len(u.GroupName) == 0 {
		return errors.New("should provide a group name")
	}
	u.GroupType = common.HTTPGroupType
	err := usergroup.Ctl.Ensure(ctx, u)
	if err != nil {
		return err
	}
	return nil
}

func (a *Auth) fillInModel(u *models.User) error {
	if strings.TrimSpace(u.Username) == "" {
		return fmt.Errorf("username cannot be empty")
	}
	u.Realname = u.Username
	u.Password = "1234567ab"
	u.Comment = userEntryComment
	if strings.Contains(u.Username, "@") {
		u.Email = u.Username
	}
	return nil
}

func (a *Auth) ensure(ctx context.Context) error {
	a.Lock()
	defer a.Unlock()
	if a.client == nil {
		a.client = &http.Client{}
	}
	if time.Now().Sub(a.settingTimeStamp) >= refreshDuration {
		setting, err := config.HTTPAuthProxySetting(ctx)
		if err != nil {
			return err
		}
		a.Endpoint = setting.Endpoint
		a.TokenReviewEndpoint = setting.TokenReviewEndpoint
		a.SkipSearch = setting.SkipSearch
		tlsCfg, err := getTLSConfig(setting)
		if err != nil {
			return err
		}
		transport.TLSClientConfig = tlsCfg
		a.client.Transport = transport
	}
	return nil
}

func getTLSConfig(setting *cfgModels.HTTPAuthProxy) (*tls.Config, error) {
	c := setting.ServerCertificate
	if setting.VerifyCert && len(c) > 0 {
		certs := x509.NewCertPool()
		if !certs.AppendCertsFromPEM([]byte(c)) {
			logger.Errorf("Failed to pin server certificate, please double check if it's valid, certificate: %s", c)
			return nil, fmt.Errorf("failed to pin server certificate for authproxy")
		}
		return &tls.Config{RootCAs: certs}, nil
	}
	return &tls.Config{InsecureSkipVerify: !setting.VerifyCert}, nil
}

func init() {
	auth.Register(common.HTTPAuth, &Auth{
		userMgr: user.New(),
	})
}
