/*
Copyright The ORAS Authors.
Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/
package remote

import (
	"fmt"

	"oras.land/oras-go/pkg/registry"
)

// buildScheme returns HTTP scheme used to access the remote registry.
func buildScheme(plainHTTP bool) string {
	if plainHTTP {
		return "http"
	}
	return "https"
}

// buildRepositoryBaseURL builds the base endpoint of the remote repository.
// Format: <scheme>://<registry>/v2/<repository>
func buildRepositoryBaseURL(plainHTTP bool, ref registry.Reference) string {
	return fmt.Sprintf("%s://%s/v2/%s", buildScheme(plainHTTP), ref.Host(), ref.Repository)
}

// buildRepositoryTagListURL builds the URL for accessing the tag list API.
// Format: <scheme>://<registry>/v2/<repository>/tags/list
// Reference: https://docs.docker.com/registry/spec/api/#tags
func buildRepositoryTagListURL(plainHTTP bool, ref registry.Reference) string {
	return buildRepositoryBaseURL(plainHTTP, ref) + "/tags/list"
}
