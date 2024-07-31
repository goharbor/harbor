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

package volcenginecr

import (
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/volcengine/volcengine-go-sdk/service/cr"

	"github.com/goharbor/harbor/src/common/http/modifier"
	"github.com/goharbor/harbor/src/lib/log"
)

// Credential ...
type Credential modifier.Modifier

type volcCredential struct {
	client    *cr.CR
	registry  string
	authCache *authCache
}

type authCache struct {
	username string
	password string
	expireAt *time.Time
}

var _ Credential = &volcCredential{}

// NewAuth will get a temporary  username and password via cr GetAuthorizationToken action for docker login
func NewAuth(client *cr.CR, registry string) Credential {
	return &volcCredential{
		client:    client,
		registry:  registry,
		authCache: &authCache{},
	}
}

func (c *volcCredential) Modify(r *http.Request) (err error) {
	if c.client == nil {
		return errNilVolcCrClient
	}
	if !c.isCacheAuthValid() {
		log.Debugf("update token %s\n", r.Host)
		authResp, err := c.client.GetAuthorizationToken(&cr.GetAuthorizationTokenInput{
			Registry: &c.registry,
		})
		if err != nil {
			return err
		}
		if authResp == nil || authResp.Username == nil || authResp.Token == nil || authResp.ExpireTime == nil {
			return errors.New("[VolcengineCR] GetAuthorizationToken output nil")
		}
		c.authCache.username = *authResp.Username
		c.authCache.password = *authResp.Token
		expireTime, err := time.Parse(time.RFC3339, *authResp.ExpireTime)
		if err != nil {
			log.Errorf("fail to parse expire time returned: %v", err)
			return fmt.Errorf("[VolcengineCR] fail to parse expire time returned: %v", err)
		}
		c.authCache.expireAt = &expireTime
	} else {
		log.Debug("token cached")
	}
	r.SetBasicAuth(c.authCache.username, c.authCache.password)
	return nil
}

func (c *volcCredential) isCacheAuthValid() bool {
	if c.authCache == nil || c.authCache.expireAt == nil {
		return false
	}
	if time.Now().After(*c.authCache.expireAt) {
		return false
	}
	return true
}
