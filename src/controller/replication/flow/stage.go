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

package flow

import (
	"fmt"
	"github.com/goharbor/harbor/src/lib/errors"
	"strings"

	repctlmodel "github.com/goharbor/harbor/src/controller/replication/model"
	"github.com/goharbor/harbor/src/lib/log"
	adp "github.com/goharbor/harbor/src/pkg/reg/adapter"
	"github.com/goharbor/harbor/src/pkg/reg/model"
	"github.com/goharbor/harbor/src/pkg/reg/util"
)

// get/create the source registry, destination registry, source adapter and destination adapter
func initialize(policy *repctlmodel.Policy) (adp.Adapter, adp.Adapter, error) {
	var srcAdapter, dstAdapter adp.Adapter
	var err error

	// create the source registry adapter
	srcFactory, err := adp.GetFactory(policy.SrcRegistry.Type)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to get adapter factory for registry type %s: %v", policy.SrcRegistry.Type, err)
	}
	srcAdapter, err = srcFactory.Create(policy.SrcRegistry)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to create adapter for source registry %s: %v", policy.SrcRegistry.URL, err)
	}

	// create the destination registry adapter
	dstFactory, err := adp.GetFactory(policy.DestRegistry.Type)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to get adapter factory for registry type %s: %v", policy.DestRegistry.Type, err)
	}
	dstAdapter, err = dstFactory.Create(policy.DestRegistry)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to create adapter for destination registry %s: %v", policy.DestRegistry.URL, err)
	}
	log.Debug("replication flow initialization completed")
	return srcAdapter, dstAdapter, nil
}

// fetch resources from the source registry
func fetchResources(adapter adp.Adapter, policy *repctlmodel.Policy) ([]*model.Resource, error) {
	var resTypes []string
	for _, filter := range policy.Filters {
		if filter.Type == model.FilterTypeResource {
			resTypes = append(resTypes, filter.Value.(string))
		}
	}
	if len(resTypes) == 0 {
		info, err := adapter.Info()
		if err != nil {
			return nil, fmt.Errorf("failed to get the adapter info: %v", err)
		}
		resTypes = append(resTypes, info.SupportedResourceTypes...)
	}

	fetchArtifact := false
	fetchChart := false
	for _, resType := range resTypes {
		if resType == model.ResourceTypeChart {
			fetchChart = true
			continue
		}
		fetchArtifact = true
	}

	var resources []*model.Resource
	// artifacts
	if fetchArtifact {
		reg, ok := adapter.(adp.ArtifactRegistry)
		if !ok {
			return nil, fmt.Errorf("the adapter doesn't implement the ArtifactRegistry interface")
		}
		res, err := reg.FetchArtifacts(policy.Filters)
		if err != nil {
			return nil, fmt.Errorf("failed to fetch artifacts: %v", err)
		}
		resources = append(resources, res...)
		log.Debug("fetch artifacts completed")
	}
	// charts
	if fetchChart {
		reg, ok := adapter.(adp.ChartRegistry)
		if !ok {
			return nil, fmt.Errorf("the adapter doesn't implement the ChartRegistry interface")
		}
		res, err := reg.FetchCharts(policy.Filters)
		if err != nil {
			return nil, fmt.Errorf("failed to fetch charts: %v", err)
		}
		resources = append(resources, res...)
		log.Debug("fetch charts completed")
	}

	log.Debug("fetch resources from the source registry completed")
	return resources, nil
}

// assemble the source resources by filling the registry information
func assembleSourceResources(resources []*model.Resource,
	policy *repctlmodel.Policy) []*model.Resource {
	for _, resource := range resources {
		resource.Registry = policy.SrcRegistry
	}
	log.Debug("assemble the source resources completed")
	return resources
}

// assemble the destination resources by filling the metadata, registry and override properties
func assembleDestinationResources(resources []*model.Resource,
	policy *repctlmodel.Policy) ([]*model.Resource, error) {
	var result []*model.Resource
	for _, resource := range resources {
		name, err := replaceNamespace(resource.Metadata.Repository.Name, policy.DestNamespace, policy.DestNamespaceReplaceCount)
		if err != nil {
			return nil, err
		}
		res := &model.Resource{
			Type:         resource.Type,
			Registry:     policy.DestRegistry,
			ExtendedInfo: resource.ExtendedInfo,
			Deleted:      resource.Deleted,
			IsDeleteTag:  resource.IsDeleteTag,
			Override:     policy.Override,
		}
		res.Metadata = &model.ResourceMetadata{
			Repository: &model.Repository{
				Name:     name,
				Metadata: resource.Metadata.Repository.Metadata,
			},
			Vtags:     resource.Metadata.Vtags,
			Artifacts: resource.Metadata.Artifacts,
		}
		result = append(result, res)
	}
	log.Debug("assemble the destination resources completed")
	return result, nil
}

// do the prepare work for pushing/uploading the resources: create the namespace or repository
func prepareForPush(adapter adp.Adapter, resources []*model.Resource) error {
	if err := adapter.PrepareForPush(resources); err != nil {
		return fmt.Errorf("failed to do the prepare work for pushing/uploading resources: %v", err)
	}
	log.Debug("the prepare work for pushing/uploading resources completed")
	return nil
}

// return the name with format "res_name" or "res_name:[vtag1,vtag2,vtag3]"
// if the resource has vtags
func getResourceName(res *model.Resource) string {
	if res == nil {
		return ""
	}
	meta := res.Metadata
	if meta == nil {
		return ""
	}
	n := 0
	if len(meta.Artifacts) > 0 {
		for _, artifact := range meta.Artifacts {
			// contains tags
			if len(artifact.Tags) > 0 {
				n += len(artifact.Tags)
				continue
			}
			// contains no tag, count digest
			if len(artifact.Digest) > 0 {
				n++
			}
		}
	} else {
		n = len(meta.Vtags)
	}

	return fmt.Sprintf("%s [%d item(s) in total]", meta.Repository.Name, n)
}

// repository:a/b/c namespace:n replaceCount: -1 -> n/c
// repository:a/b/c namespace:n replaceCount: 0 -> n/a/b/c
// repository:a/b/c namespace:n replaceCount: 1 -> n/b/c
// repository:a/b/c namespace:n replaceCount: 2 -> n/c
// repository:a/b/c namespace:n replaceCount: 3 -> n
func replaceNamespace(repository string, namespace string, replaceCount int8) (string, error) {
	if len(namespace) == 0 {
		return repository, nil
	}

	// legacy logic to keep backward compatibility
	if replaceCount < 0 {
		_, rest := util.ParseRepository(repository)
		return fmt.Sprintf("%s/%s", namespace, rest), nil
	}

	subs := strings.Split(repository, "/")
	len := len(subs)
	switch {
	case replaceCount == 0:
		return fmt.Sprintf("%s/%s", namespace, repository), nil
	case int(replaceCount) == len:
		return namespace, nil
	case int(replaceCount) > len:
		return "", errors.New(nil).WithCode(errors.BadRequestCode).
			WithMessage("the repository %s contains only %d substrings, but the destination namespace replace count is %d",
				repository, len, replaceCount)
	default:
		return fmt.Sprintf("%s/%s", namespace, strings.Join(subs[replaceCount:], "/")), nil
	}
}
