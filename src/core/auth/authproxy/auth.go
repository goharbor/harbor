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
	"fmt"
	"github.com/goharbor/harbor/src/common"
	"github.com/goharbor/harbor/src/common/dao"
	"github.com/goharbor/harbor/src/common/models"
	"github.com/goharbor/harbor/src/common/utils/log"
	"github.com/goharbor/harbor/src/core/auth"
	"io/ioutil"
	"net/http"
	"os"
	"strings"
	"sync"
)

// Auth implements HTTP authenticator the required attributes.
// The attribute Endpoint is the HTTP endpoint to which the POST request should be issued for authentication
type Auth struct {
	auth.DefaultAuthenticateHelper
	sync.Mutex
	Endpoint       string
	SkipCertVerify bool
	AlwaysOnboard  bool
	client         *http.Client
}

// Authenticate issues http POST request to Endpoint if it returns 200 the authentication is considered success.
func (a *Auth) Authenticate(m models.AuthModel) (*models.User, error) {
	a.ensure()
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
		return &models.User{Username: m.Principal}, nil
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

// OnBoardUser delegates to dao pkg to insert/update data in DB.
func (a *Auth) OnBoardUser(u *models.User) error {
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

// SearchUser - TODO: Remove this workaround when #6767 is fixed.
// When the flag is set it always return the default model without searching
func (a *Auth) SearchUser(username string) (*models.User, error) {
	a.ensure()
	var queryCondition = models.User{
		Username: username,
	}
	u, err := dao.GetUser(queryCondition)
	if err != nil {
		return nil, err
	}
	if a.AlwaysOnboard && u == nil {
		u = &models.User{Username: username}
		if err := a.fillInModel(u); err != nil {
			return nil, err
		}
	}
	return u, nil
}

func (a *Auth) fillInModel(u *models.User) error {
	if strings.TrimSpace(u.Username) == "" {
		return fmt.Errorf("username cannot be empty")
	}
	u.Realname = u.Username
	u.Password = "1234567ab"
	u.Comment = "By Authproxy"
	if strings.Contains(u.Username, "@") {
		u.Email = u.Username
	} else {
		u.Email = fmt.Sprintf("%s@placeholder.com", u.Username)
	}
	return nil
}

func (a *Auth) ensure() {
	a.Lock()
	defer a.Unlock()
	if a.Endpoint == "" {
		a.Endpoint = os.Getenv("AUTHPROXY_ENDPOINT")
		a.SkipCertVerify = strings.EqualFold(os.Getenv("AUTHPROXY_SKIP_CERT_VERIFY"), "true")
		a.AlwaysOnboard = strings.EqualFold(os.Getenv("AUTHPROXY_ALWAYS_ONBOARD"), "true")
	}
	if a.client == nil {
		tr := &http.Transport{
			TLSClientConfig: &tls.Config{
				InsecureSkipVerify: a.SkipCertVerify,
			},
		}
		a.client = &http.Client{
			Transport: tr,
		}
	}
}

func init() {
	auth.Register(common.HTTPAuth, &Auth{})
}
