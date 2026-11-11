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

package aliacr

import (
	"errors"
	"net/http"
	"time"

	"github.com/goharbor/harbor/src/common/http/modifier"
	"github.com/goharbor/harbor/src/lib/log"
)

// Credential ...
type Credential modifier.Modifier

// Implements interface Credential
type aliyunAuthCredential struct {
	acrAPI              openapi
	cacheToken          *registryTemporaryToken
	cacheTokenExpiredAt time.Time
}

type registryTemporaryToken struct {
	user     string
	password string
}

var _ Credential = &aliyunAuthCredential{}

// NewAuth will get a temporary docker registry username and password via aliyun cr service API.
func NewAuth(acrAPI openapi) Credential {
	return &aliyunAuthCredential{
		acrAPI:     acrAPI,
		cacheToken: &registryTemporaryToken{},
	}
}

func (a *aliyunAuthCredential) Modify(r *http.Request) (err error) {
	if !a.isCacheTokenValid() {
		log.Debugf("[aliyunAuthCredential.Modify.updateToken]Host: %s\n", r.Host)
		if a.acrAPI == nil {
			return errors.New("acr api is nil")
		}
		v, err := a.acrAPI.GetAuthorizationToken()
		if err != nil {
			return err
		}
		a.cacheTokenExpiredAt = v.expiresAt
		a.cacheToken.user = v.user
		a.cacheToken.password = v.password
	} else {
		log.Debug("[aliyunAuthCredential] USE CACHE TOKEN!!!")
	}

	r.SetBasicAuth(a.cacheToken.user, a.cacheToken.password)
	return
}

func (a *aliyunAuthCredential) isCacheTokenValid() bool {
	if a.cacheTokenExpiredAt.IsZero() {
		return false
	}
	if a.cacheToken == nil {
		return false
	}
	if time.Now().After(a.cacheTokenExpiredAt) {
		return false
	}

	return true
}
