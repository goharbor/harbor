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
	"regexp"

	"github.com/docker/distribution/registry/client/auth/challenge"

	"github.com/goharbor/harbor/src/lib/log"
	"github.com/goharbor/harbor/src/pkg/reg/util"
)

func getRegionRegistryName(url string) (string, string, error) {
	reg := regexp.MustCompile(`https://(.*)\.cr\.volces|ivolces\.com`)
	rs := reg.FindStringSubmatch(url)
	if rs == nil || len(rs) != 2 {
		return "", "", errors.New("Invalid url")
	}
	registryNameRegion := rs[1]
	for regionReg := range regionRegs {
		reg = regexp.MustCompile(regionReg)
		res := reg.FindStringSubmatch(registryNameRegion)
		if res == nil || len(res) != 3 {
			log.Debug("fail to match", "reg", regionReg)
			continue
		}
		return res[2], res[1], nil
	}

	return "", "", errors.New("invalid region")
}

func getRealmService(host string, insecure bool) (string, string, error) {
	client := &http.Client{
		Transport: util.GetHTTPTransport(insecure),
	}

	resp, err := client.Get(host + "/v2/")
	if err != nil {
		return "", "", err
	}
	defer resp.Body.Close() // nolint
	challenges := challenge.ResponseChallenges(resp)
	for _, challenge := range challenges {
		if challenge.Scheme == "bearer" {
			return challenge.Parameters["realm"], challenge.Parameters["service"], nil
		}
	}
	return "", "", fmt.Errorf("bearer auth scheme isn't supported: %v", challenges)
}
