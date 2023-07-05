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

package dockerhub

import "fmt"

const (
	baseURL             = "https://hub.docker.com"
	registryURL         = "https://registry-1.docker.io"
	loginPath           = "/v2/users/login/"
	listNamespacePath   = "/v2/repositories/namespaces"
	createNamespacePath = "/v2/orgs/"

	metadataKeyCompany  = "company"
	metadataKeyFullName = "fullName"
)

func getNamespacePath(namespace string) string {
	return fmt.Sprintf("/v2/orgs/%s/", namespace)
}

func listReposPath(namespace, name string, page, pageSize int) string {
	if len(name) == 0 {
		return fmt.Sprintf("/v2/repositories/%s/?page=%d&page_size=%d", namespace, page, pageSize)
	}

	return fmt.Sprintf("/v2/repositories/%s/?name=%s&page=%d&page_size=%d", namespace, name, page, pageSize)
}

func listTagsPath(namespace, repo string, page, pageSize int) string {
	return fmt.Sprintf("/v2/repositories/%s/%s/tags/?page=%d&page_size=%d", namespace, repo, page, pageSize)
}

func deleteTagPath(namespace, repo, tag string) string {
	return fmt.Sprintf("/v2/repositories/%s/%s/tags/%s/", namespace, repo, tag)
}
