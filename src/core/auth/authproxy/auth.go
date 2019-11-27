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
	"crypto/tls"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/goharbor/harbor/src/common/dao/group"

	"github.com/goharbor/harbor/src/common"
	"github.com/goharbor/harbor/src/common/dao"
	"github.com/goharbor/harbor/src/common/models"
	"github.com/goharbor/harbor/src/common/utils/log"
	"github.com/goharbor/harbor/src/core/auth"
	"github.com/goharbor/harbor/src/core/config"
	"github.com/goharbor/harbor/src/pkg/authproxy"
)

const refreshDuration = 2 * time.Second
const userEntryComment = "By Authproxy"

var secureTransport = &http.Transport{
	Proxy: http.ProxyFromEnvironment,
}
var insecureTransport = &http.Transport{
	TLSClientConfig: &tls.Config{
		InsecureSkipVerify: true,
	},
	Proxy: http.ProxyFromEnvironment,
}

// Auth implements HTTP authenticator the required attributes.
// The attribute Endpoint is the HTTP endpoint to which the POST request should be issued for authentication
type Auth struct {
	auth.DefaultAuthenticateHelper
	sync.Mutex
	Endpoint            string
	TokenReviewEndpoint string
	SkipCertVerify      bool
	SkipSearch          bool
	// When this attribute is set to false, the name of user/group will be converted to lower-case when onboarded to Harbor, so
	// as long as the authentication is successful there's no difference in terms of upper or lower case that is used.
	// It will be mapped to one entry in Harbor's User/Group table.
	CaseSensitive    bool
	settingTimeStamp time.Time
	client           *http.Client
}

type session struct {
	SessionID string `json:"session_id,omitempty"`
}

// Authenticate issues http POST request to Endpoint if it returns 200 the authentication is considered success.
func (a *Auth) Authenticate(m models.AuthModel) (*models.User, error) {
	err := a.ensure()
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
	defer resp.Body.Close()
	if resp.StatusCode == http.StatusOK {
		name := a.normalizeName(m.Principal)
		user := &models.User{Username: name}
		data, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			log.Warningf("Failed to read response body, error: %v", err)
			return nil, auth.ErrAuth{}
		}
		s := session{}
		err = json.Unmarshal(data, &s)
		if err != nil {
			return nil, auth.NewErrAuth(fmt.Sprintf("failed to read session %v", err))
		}
		if err := a.tokenReview(s.SessionID, user); err != nil {
			return nil, auth.NewErrAuth(err.Error())
		}
		return user, nil
	} else if resp.StatusCode == http.StatusUnauthorized {
		return nil, auth.ErrAuth{}
	} else {
		data, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			log.Warningf("Failed to read response body, error: %v", err)
		}
		return nil, fmt.Errorf("failed to authenticate, status code: %d, text: %s", resp.StatusCode, string(data))
	}
}

func (a *Auth) tokenReview(sessionID string, user *models.User) error {
	httpAuthProxySetting, err := config.HTTPAuthProxySetting()
	if err != nil {
		return err
	}
	reviewStatus, err := authproxy.TokenReview(sessionID, httpAuthProxySetting)
	if err != nil {
		return err
	}
	u2, err := authproxy.UserFromReviewStatus(reviewStatus)
	if err != nil {
		return err
	}
	user.GroupIDs = u2.GroupIDs
	return nil
}

// OnBoardUser delegates to dao pkg to insert/update data in DB.
func (a *Auth) OnBoardUser(u *models.User) error {
	u.Username = a.normalizeName(u.Username)
	return dao.OnBoardUser(u)
}

// PostAuthenticate generates the user model and on board the user.
func (a *Auth) PostAuthenticate(u *models.User) error {
	if res, _ := dao.GetUser(*u); res != nil {
		return nil
	}
	if err := a.fillInModel(u); err != nil {
		return err
	}
	return a.OnBoardUser(u)
}

// SearchUser returns nil as authproxy does not have such capability.
// When SkipSearch is set it always return the default model,
// the username will be switch to lowercase if it's configured as case-insensitive
func (a *Auth) SearchUser(username string) (*models.User, error) {
	err := a.ensure()
	if err != nil {
		log.Warningf("Failed to refresh configuration for HTTP Auth Proxy Authenticator, error: %v, the default settings will be used", err)
	}
	username = a.normalizeName(username)
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
func (a *Auth) SearchGroup(groupKey string) (*models.UserGroup, error) {
	err := a.ensure()
	if err != nil {
		log.Warningf("Failed to refresh configuration for HTTP Auth Proxy Authenticator, error: %v, the default settings will be used", err)
	}
	groupKey = a.normalizeName(groupKey)
	var ug *models.UserGroup
	if a.SkipSearch {
		ug = &models.UserGroup{
			GroupName: groupKey,
			GroupType: common.HTTPGroupType,
		}
		return ug, nil
	}
	return nil, nil
}

// OnBoardGroup create user group entity in Harbor DB, altGroupName is not used.
func (a *Auth) OnBoardGroup(u *models.UserGroup, altGroupName string) error {
	// if group name provided, on board the user group
	if len(u.GroupName) == 0 {
		return errors.New("Should provide a group name")
	}
	u.GroupType = common.HTTPGroupType
	u.GroupName = a.normalizeName(u.GroupName)
	err := group.OnBoardUserGroup(u)
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

func (a *Auth) ensure() error {
	a.Lock()
	defer a.Unlock()
	if a.client == nil {
		a.client = &http.Client{}
	}
	if time.Now().Sub(a.settingTimeStamp) >= refreshDuration {
		setting, err := config.HTTPAuthProxySetting()
		if err != nil {
			return err
		}
		a.Endpoint = setting.Endpoint
		a.TokenReviewEndpoint = setting.TokenReviewEndpoint
		a.SkipCertVerify = !setting.VerifyCert
		a.SkipSearch = setting.SkipSearch
	}
	if a.SkipCertVerify {
		a.client.Transport = insecureTransport
	} else {
		a.client.Transport = secureTransport
	}

	return nil
}

func (a *Auth) normalizeName(n string) string {
	if !a.CaseSensitive {
		return strings.ToLower(n)
	}
	return n
}

func init() {
	auth.Register(common.HTTPAuth, &Auth{})
}
