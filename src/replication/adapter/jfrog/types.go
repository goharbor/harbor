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

package jfrog

type repository struct {
	Key         string `json:"key"`
	Type        string `json:"type"`
	URL         string `json:"url"`
	PackageType string `json:"packageType"`
}

type repositoryCreate struct {
	Key           string `json:"key"`
	Rclass        string `json:"rclass"`
	PackageType   string `json:"packageType"`
	RepoLayoutRef string `json:"repoLayoutRef"`
}

func newDefaultDockerLocalRepository(key string) *repositoryCreate {
	return &repositoryCreate{
		Key:           key,
		Rclass:        "local",
		PackageType:   "docker",
		RepoLayoutRef: "simple-default",
	}
}
