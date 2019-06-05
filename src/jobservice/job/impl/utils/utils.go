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

	"github.com/docker/distribution/registry/auth/token"
	httpauth "github.com/goharbor/harbor/src/common/http/modifier/auth"
	"github.com/goharbor/harbor/src/common/utils/registry"
	"github.com/goharbor/harbor/src/common/utils/registry/auth"
)

var coreClient *http.Client
var mutex = &sync.Mutex{}

// NewRepositoryClient creates a repository client with standard token authorizer
func NewRepositoryClient(endpoint string, insecure bool, credential auth.Credential,
	tokenServiceEndpoint, repository string) (*registry.Repository, error) {

	transport := registry.GetHTTPTransport(insecure)

	authorizer := auth.NewStandardTokenAuthorizer(&http.Client{
		Transport: transport,
	}, credential, tokenServiceEndpoint)

	uam := &UserAgentModifier{
		UserAgent: "harbor-registry-client",
	}

	return registry.NewRepository(repository, endpoint, &http.Client{
		Transport: registry.NewTransport(transport, authorizer, uam),
	})
}

// NewRepositoryClientForJobservice creates a repository client that can only be used to
// access the internal registry
func NewRepositoryClientForJobservice(repository, internalRegistryURL, secret, internalTokenServiceURL string) (*registry.Repository, error) {
	transport := registry.GetHTTPTransport()
	credential := httpauth.NewSecretAuthorizer(secret)

	authorizer := auth.NewStandardTokenAuthorizer(&http.Client{
		Transport: transport,
	}, credential, internalTokenServiceURL)

	uam := &UserAgentModifier{
		UserAgent: "harbor-registry-client",
	}

	return registry.NewRepository(repository, internalRegistryURL, &http.Client{
		Transport: registry.NewTransport(transport, authorizer, uam),
	})
}

// UserAgentModifier adds the "User-Agent" header to the request
type UserAgentModifier struct {
	UserAgent string
}

// Modify adds user-agent header to the request
func (u *UserAgentModifier) Modify(req *http.Request) error {
	req.Header.Set(http.CanonicalHeaderKey("User-Agent"), u.UserAgent)
	return nil
}

// BuildBlobURL ...
func BuildBlobURL(endpoint, repository, digest string) string {
	return fmt.Sprintf("%s/v2/%s/blobs/%s", endpoint, repository, digest)
}

// GetTokenForRepo is used for job handler to get a token for clair.
func GetTokenForRepo(repository, secret, internalTokenServiceURL string) (string, error) {
	credential := httpauth.NewSecretAuthorizer(secret)
	t, err := auth.GetToken(internalTokenServiceURL, false, credential,
		[]*token.ResourceActions{{
			Type:    "repository",
			Name:    repository,
			Actions: []string{"pull"},
		}})
	if err != nil {
		return "", err
	}

	return t.Token, nil
}

// GetClient returns the HTTP client that will attach jobservice secret to the request, which can be used for
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
