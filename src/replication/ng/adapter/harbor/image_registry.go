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

package harbor

import (
	"fmt"

	"github.com/goharbor/harbor/src/replication/ng/model"
)

type repository struct {
	Name string `json:"name"`
}

type tag struct {
	Name string `json:"name"`
}

// TODO implement filter
func (a *adapter) FetchImages(namespaces []string, filters []*model.Filter) ([]*model.Resource, error) {
	resources := []*model.Resource{}
	for _, namespace := range namespaces {
		project, err := a.getProject(namespace)
		if err != nil {
			return nil, err
		}
		repositories := []*repository{}
		url := fmt.Sprintf("%s/api/repositories?project_id=%d", a.coreServiceURL, project.ID)
		if err = a.client.Get(url, &repositories); err != nil {
			return nil, err
		}

		for _, repository := range repositories {
			url := fmt.Sprintf("%s/api/repositories/%s/tags", a.coreServiceURL, repository.Name)
			tags := []*tag{}
			if err = a.client.Get(url, &tags); err != nil {
				return nil, err
			}
			vtags := []string{}
			for _, tag := range tags {
				vtags = append(vtags, tag.Name)
			}
			resources = append(resources, &model.Resource{
				Type:     model.ResourceTypeRepository,
				Registry: a.registry,
				Metadata: &model.ResourceMetadata{
					Namespace: namespace,
					Name:      repository.Name,
					Vtags:     vtags,
				},
			})
		}
	}

	return resources, nil
}

// override the default implementation from the default image registry
// by calling Harbor API directly
func (a *adapter) DeleteManifest(repository, reference string) error {
	url := fmt.Sprintf("%s/api/repositories/%s/tags/%s", a.coreServiceURL, repository, reference)
	return a.client.Delete(url)
}
