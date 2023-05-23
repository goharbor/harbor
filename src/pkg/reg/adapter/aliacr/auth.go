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
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/aliyun/alibaba-cloud-sdk-go/services/cr"

	"github.com/goharbor/harbor/src/common/http/modifier"
	"github.com/goharbor/harbor/src/lib/log"
)

// Credential ...
type Credential modifier.Modifier

// Implements interface Credential
type aliyunAuthCredential struct {
	region              string
	accessKey           string
	secretKey           string
	cacheToken          *registryTemporaryToken
	cacheTokenExpiredAt time.Time
}

type registryTemporaryToken struct {
	user     string
	password string
}

var _ Credential = &aliyunAuthCredential{}

// NewAuth will get a temporary docker registry username and password via aliyun cr service API.
func NewAuth(region, accessKey, secretKey string) Credential {
	return &aliyunAuthCredential{
		region:     region,
		accessKey:  accessKey,
		secretKey:  secretKey,
		cacheToken: &registryTemporaryToken{},
	}
}

func (a *aliyunAuthCredential) Modify(r *http.Request) (err error) {
	if !a.isCacheTokenValid() {
		log.Debugf("[aliyunAuthCredential.Modify.updateToken]Host: %s\n", r.Host)
		var client *cr.Client
		client, err = cr.NewClientWithAccessKey(a.region, a.accessKey, a.secretKey)
		if err != nil {
			return
		}

		var tokenRequest = cr.CreateGetAuthorizationTokenRequest()
		var tokenResponse *cr.GetAuthorizationTokenResponse
		tokenRequest.SetDomain(fmt.Sprintf(endpointTpl, a.region))
		tokenResponse, err = client.GetAuthorizationToken(tokenRequest)
		if err != nil {
			return
		}
		var v authorizationToken
		err = json.Unmarshal(tokenResponse.GetHttpContentBytes(), &v)
		if err != nil {
			return
		}
		a.cacheTokenExpiredAt = v.Data.ExpireDate.ToTime()
		a.cacheToken.user = v.Data.TempUserName
		a.cacheToken.password = v.Data.AuthorizationToken
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
