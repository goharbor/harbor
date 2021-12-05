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

package util

import (
	"net/http"
	"strings"

	"github.com/docker/distribution/registry/client/auth/challenge"
	"github.com/goharbor/harbor/src/lib/errors"
	"github.com/goharbor/harbor/src/pkg/reg/model"

	commonhttp "github.com/goharbor/harbor/src/common/http"
)

// GetHTTPTransport can be used to share the common HTTP transport
func GetHTTPTransport(insecure bool) http.RoundTripper {
	if insecure {
		return commonhttp.GetHTTPTransport(commonhttp.WithInsecure(true))
	}
	return commonhttp.GetHTTPTransport()
}

func Ping(registry *model.Registry) (string, string, error) {
	client := &http.Client{
		Transport: GetHTTPTransport(registry.Insecure),
	}

	resp, err := client.Get(registry.URL + "/v2/")
	if err != nil {
		return "", "", err
	}
	defer resp.Body.Close()
	challenges := challenge.ResponseChallenges(resp)
	for _, challenge := range challenges {
		if challenge.Scheme == "bearer" {
			return challenge.Parameters["realm"], challenge.Parameters["service"], nil
		}
	}
	return "", "", errors.New(nil).WithCode(errors.ChallengesUnsupportedCode).WithMessage("bearer auth scheme isn't supported: %v", challenges)
}

// ParseRepository parses the "repository" provided into two parts: namespace and the rest
// the string before the last "/" is the namespace part
// c -> [,c]
// b/c -> [b,c]
// a/b/c -> [a/b,c]
func ParseRepository(repository string) (string, string) {
	if len(repository) == 0 {
		return "", ""
	}
	index := strings.LastIndex(repository, "/")
	if index == -1 {
		return "", repository
	}
	return repository[:index], repository[index+1:]
}
