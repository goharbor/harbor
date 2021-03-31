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

package quay

import (
	"fmt"
	"strings"
)

type cred struct {
	OAuth2Token       string `json:"oauth2_token"`
	AccountName       string `json:"account_name"`
	DockerCliPassword string `json:"docker_cli_password"`
}

type orgCreate struct {
	Name  string `json:"name"`
	Email string `json:"email"`
}

func buildOrgURL(endpoint, orgName string) string {
	return fmt.Sprintf("%s/api/v1/organization/%s", strings.TrimRight(endpoint, "/"), orgName)
}
