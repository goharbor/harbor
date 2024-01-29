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
	"fmt"
	"strings"

	"github.com/opencontainers/go-digest"
	"github.com/volcengine/volcengine-go-sdk/service/cr"

	"github.com/goharbor/harbor/src/common/utils"
	"github.com/goharbor/harbor/src/lib/log"
	"github.com/goharbor/harbor/src/pkg/reg/model"
	"github.com/goharbor/harbor/src/pkg/reg/util"
)

// DeleteManifest VolcCR will use our own openAPI to delete Manifest
func (a *adapter) DeleteManifest(repository, reference string) (err error) {
	parts := strings.SplitN(repository, "/", 2)
	if len(parts) != 2 {
		return fmt.Errorf("VolcEngineCR only support repo in format <namespace>/<repo>, but got: %s", repository)
	}
	log.Warningf("namespace=%s, repository=%s, tag=%s", parts[0], parts[1], reference)

	if _, err := digest.Parse(reference); err != nil {
		// get digest
		resp, err := a.volcCrClient.ListTags(&cr.ListTagsInput{
			Registry:   a.registryName,
			Namespace:  &parts[0],
			Repository: &parts[1],
			Filter: &cr.FilterForListTagsInput{
				Names: []*string{
					&reference,
				},
			},
		})
		if err != nil {
			return err
		}
		if resp == nil || resp.TotalCount == nil {
			return fmt.Errorf("[VolcEngineCR.DeleteManifest] ListTags resp nil")
		}
		if *resp.TotalCount == 0 {
			return nil
		}
		if resp.Items[0] == nil {
			return fmt.Errorf("[VolcEngineCR.DeleteManifest] ListTags resp nil")
		}
		reference = *resp.Items[0].Digest
	}
	// listCandidateTags based on digest
	tags, err := a.listCandidateTags(parts[0], parts[1], reference)
	if err != nil {
		log.Errorf("DeleteManifest error :%v", err)
		return err
	}
	// deleteTags
	err = a.deleteTags(parts[0], parts[1], tags)
	if err != nil {
		log.Errorf("DeleteManifest error :%v", err)
	}
	return
}

// DeleteTag VolcCR will use our own openAPI to delete tag
func (a *adapter) DeleteTag(repository, tag string) (err error) {
	parts := strings.SplitN(repository, "/", 2)
	if len(parts) != 2 {
		return fmt.Errorf("VolcEngineCR only support repo in format <namespace>/<repo>, but got: %s", repository)
	}
	log.Warningf("namespace=%s, repository=%s, tag=%s", parts[0], parts[1], tag)

	err = a.deleteTags(parts[0], parts[1], []*string{
		&tag,
	})
	if err != nil {
		log.Errorf("deleteTag error: %v", err)
	}
	return
}

// FetchArtifacts VolcCR not support /v2/_catalog of Registry
func (a *adapter) FetchArtifacts(filters []*model.Filter) ([]*model.Resource, error) {
	log.Debug("FetchArtifacts filters", "filters", filters)
	// 1. get filter pattern
	var repoPattern string
	var tagsPattern string
	for _, filter := range filters {
		if filter.Type == model.FilterTypeName {
			repoPattern = filter.Value.(string)
		}
		if filter.Type == model.FilterTypeTag {
			tagsPattern = filter.Value.(string)
		}
	}
	namespacePattern := strings.Split(repoPattern, "/")[0]

	log.Debug("read in filter patterns", "repoPattern", repoPattern, "tagsPattern", tagsPattern)

	// 2. list namespace candidtes
	namespaces, err := a.listCandidateNamespaces(namespacePattern)
	if err != nil {
		log.Errorf("FetchArtifacts error: %v", err)
		return nil, err
	}
	log.Debug("FetchArtifacts filtered namespace", "namespace", namespaces)

	// 3. list repos
	var nsRepos []string
	for _, ns := range namespaces {
		repoCandidates, err := a.listRepositories(ns)
		if err != nil {
			log.Error("FetchArtifacts error", "error", err)
			return nil, err
		}
		log.Debug(" FetchArtifacts list repo", "repos: ", repoCandidates)
		for _, r := range repoCandidates {
			nsRepoCandidate := fmt.Sprintf("%s/%s", ns, r)
			ok, err := util.Match(repoPattern, nsRepoCandidate)
			if err != nil {
				log.Error("FetchArtifacts error", "error", err)
				return nil, err
			}
			log.Debug("filter namespaced repository", "repoPattern: ", repoPattern, "repo: ", nsRepoCandidate)
			if ok {
				nsRepos = append(nsRepos, nsRepoCandidate)
			}
		}
	}
	log.Debug("filter namespaced repository", "length", len(nsRepos))

	// 4. list tags
	var rawResources = make([]*model.Resource, len(nsRepos))
	resources := make([]*model.Resource, 0)
	runner := utils.NewLimitedConcurrentRunner(concurrentLimit)

	for idx, repo := range nsRepos {
		i := idx
		nsRepo := repo
		runner.AddTask(func() error {
			repoArr := strings.SplitN(nsRepo, "/", 2)
			// note list tag don't tell different oci types now
			candidateTags, err := a.listAllTags(repoArr[0], repoArr[1])
			if err != nil {
				log.Error("fail to list all tags", "nsRepo", nsRepo)
				return fmt.Errorf("volcengineCR fail to list all tags %w", err)
			}

			tags := make([]string, 0)
			if tagsPattern != "" {
				for _, candidateTag := range candidateTags {
					ok, err := util.Match(tagsPattern, candidateTag)
					if err != nil {
						return fmt.Errorf("fail to match tag pattern, error=%w", err)
					}
					if ok {
						tags = append(tags, candidateTag)
					}
				}
			} else {
				tags = candidateTags
			}

			log.Debug("filter tags")

			if len(tags) > 0 {
				rawResources[i] = &model.Resource{
					Type:     model.ResourceTypeImage,
					Registry: a.registry,
					Metadata: &model.ResourceMetadata{
						Repository: &model.Repository{
							Name: nsRepo,
						},
						Vtags: tags,
					},
				}
			}
			return nil
		})
	}
	if err = runner.Wait(); err != nil {
		return nil, fmt.Errorf("failed to fetch artifacts: %w", err)
	}

	for _, res := range rawResources {
		if res != nil {
			resources = append(resources, res)
		}
	}

	return resources, nil
}
