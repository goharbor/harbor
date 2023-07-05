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

package gitlab

// TokenResp is response of login.
type TokenResp struct {
	Token string `json:"token"`
}

// Project describes a project in Gitlab
type Project struct {
	ID              int64  `json:"id"`
	Name            string `json:"name"`
	FullPath        string `json:"path_with_namespace"`
	Visibility      string `json:"visibility"`
	RegistryEnabled bool   `json:"container_registry_enabled"`
}

// Repository describes a repository in Gitlab
type Repository struct {
	ID       int64  `json:"id"`
	Name     string `json:"name"`
	Path     string `json:"path"`
	Location string `json:"location"`
}

// Tag describes a tag in Gitlab
type Tag struct {
	Name     string `json:"name"`
	Path     string `json:"path"`
	Location string `json:"location"`
}
