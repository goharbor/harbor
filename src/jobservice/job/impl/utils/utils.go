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

package utils

import (
	"fmt"
	"net/http"
	"os"
	"sync"

	httpauth "github.com/goharbor/harbor/src/common/http/modifier/auth"
	"github.com/goharbor/harbor/src/common/utils/registry"
)

var coreClient *http.Client
var mutex = &sync.Mutex{}

// UserAgentModifier adds the "User-Agent" header to the request
type UserAgentModifier struct {
	UserAgent string
}

// Modify adds user-agent header to the request
func (u *UserAgentModifier) Modify(req *http.Request) error {
	req.Header.Set(http.CanonicalHeaderKey("User-Agent"), u.UserAgent)
	return nil
}

// GetClient returns the HTTP client that will attach jobservce secret to the request, which can be used for
// accessing Harbor's Core Service.
// This function returns error if the secret of Job service is not set.
func GetClient() (*http.Client, error) {
	mutex.Lock()
	defer mutex.Unlock()
	if coreClient == nil {
		secret := os.Getenv("JOBSERVICE_SECRET")
		if len(secret) == 0 {
			return nil, fmt.Errorf("unable to load secret for job service")
		}
		modifier := httpauth.NewSecretAuthorizer(secret)
		coreClient = &http.Client{Transport: registry.NewTransport(&http.Transport{}, modifier)}
	}
	return coreClient, nil
}
