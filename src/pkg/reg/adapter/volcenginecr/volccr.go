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
	"math"

	"github.com/opencontainers/go-digest"
	"github.com/volcengine/volcengine-go-sdk/service/cr"

	"github.com/goharbor/harbor/src/lib/log"
	"github.com/goharbor/harbor/src/pkg/reg/util"
)

func (a *adapter) createNamespace(namespace string) (err error) {
	if a.volcCrClient == nil {
		return errNilVolcCrClient
	}
	// check if exists
	exist, err := a.namespaceExist(namespace)
	if err != nil {
		return
	}
	// if exists; skip create
	if exist {
		return nil
	}

	// create Namespace
	_, err = a.volcCrClient.CreateNamespace(&cr.CreateNamespaceInput{
		Registry: a.registryName,
		Name:     &namespace,
	})
	if err != nil {
		log.Debugf("CreateNamespace error:%v", err)
		return err
	}
	return nil
}

func (a *adapter) createRepository(namespace, repository string) (err error) {
	if a.volcCrClient == nil {
		return errNilVolcCrClient
	}
	// check if exists
	res, err := a.volcCrClient.ListRepositories(&cr.ListRepositoriesInput{
		Registry: a.registryName,
		Filter: &cr.FilterForListRepositoriesInput{
			Names: []*string{
				&repository,
			},
			Namespaces: []*string{
				&namespace,
			},
		},
	})
	if err != nil {
		log.Debugf("ListRepositories error:%v", err)
		return err
	}
	// if exists; skip create
	if res != nil && res.TotalCount != nil && *res.TotalCount > 0 {
		return nil
	}

	// create Repository
	_, err = a.volcCrClient.CreateRepository(&cr.CreateRepositoryInput{
		Registry:  a.registryName,
		Namespace: &namespace,
		Name:      &repository,
	})
	if err != nil {
		log.Debugf("CreateRepository error:%v", err)
		return err
	}
	return nil
}

func (a *adapter) deleteTags(namespace, repository string, tags []*string) error {
	_, err := a.volcCrClient.DeleteTags(&cr.DeleteTagsInput{
		Registry:   a.registryName,
		Namespace:  &namespace,
		Repository: &repository,
		Names:      tags,
	})
	return err
}

func (a *adapter) listCandidateNamespaces(namespacePattern string) ([]string, error) {
	if a.volcCrClient == nil {
		return []string{}, errNilVolcCrClient
	}
	namespaces := make([]string, 0)
	// filter namespaces
	if len(namespacePattern) > 0 {
		if nms, ok := util.IsSpecificPathComponent(namespacePattern); ok {
			// Check if namespace exist
			for _, ns := range nms {
				exist, err := a.namespaceExist(ns)
				if err != nil {
					return nil, err
				}
				if !exist {
					continue
				}
				namespaces = append(namespaces, nms...)
			}
		}
	}

	if len(namespaces) > 0 {
		log.Debug("list candidate namespace", "pattern", namespacePattern, "namespaces", namespaces)
		return namespaces, nil
	}

	// list all
	return a.listNamespaces()
}

func (a *adapter) listNamespaces() ([]string, error) {
	if a.volcCrClient == nil {
		return []string{}, errNilVolcCrClient
	}
	pageSize := MaxPageSize
	pageNumber := int64(1)
	initCondition := true
	var remain int64 = math.MaxInt64
	var nsList []string

	for remain > 0 {
		resp, err := a.volcCrClient.ListNamespaces(
			&cr.ListNamespacesInput{
				Registry:   a.registryName,
				PageSize:   &pageSize,
				PageNumber: &pageNumber,
			})

		if err != nil {
			return nil, err
		}
		if resp == nil || resp.TotalCount == nil {
			return nil, errListNamespaceResp
		}
		if initCondition {
			nsList = make([]string, 0, *resp.TotalCount)
			remain = *resp.TotalCount - pageSize
			initCondition = false
		} else {
			remain -= pageSize
		}
		// be careful with state machine.
		pageNumber++
		for _, nsInfo := range resp.Items {
			if nsInfo != nil && nsInfo.Name != nil {
				nsList = append(nsList, *nsInfo.Name)
			}
		}
	}
	return nsList, nil
}

func (a *adapter) listRepositories(namespace string) ([]string, error) {
	if a.volcCrClient == nil {
		return []string{}, errNilVolcCrClient
	}
	pageSize := MaxPageSize
	pageNumber := int64(1)
	initCondition := true
	var remain int64 = math.MaxInt64
	var repoList []string

	for remain > 0 {
		resp, err := a.volcCrClient.ListRepositories(
			&cr.ListRepositoriesInput{
				Registry:   a.registryName,
				PageSize:   &pageSize,
				PageNumber: &pageNumber,
				Filter: &cr.FilterForListRepositoriesInput{
					Namespaces: []*string{&namespace},
				},
			})
		if err != nil {
			return nil, err
		}
		if resp == nil || resp.TotalCount == nil {
			return nil, errListRepositoriesResp
		}
		if initCondition {
			repoList = make([]string, 0, *resp.TotalCount)
			remain = *resp.TotalCount - pageSize
			initCondition = false
		} else {
			remain -= pageSize
		}
		// be careful with state machine.
		pageNumber++
		for _, repoInfo := range resp.Items {
			if repoInfo != nil && repoInfo.Name != nil {
				repoList = append(repoList, *repoInfo.Name)
			}
		}
	}
	return repoList, nil
}

// listAllTags list all tags of different artifacts with given namespace and repo
func (a *adapter) listAllTags(namespace, repo string) ([]string, error) {
	if a.volcCrClient == nil {
		return []string{}, errNilVolcCrClient
	}
	pageSize := MaxPageSize
	pageNumber := int64(1)
	initCondition := true
	var remain int64 = math.MaxInt64
	var tagList []string

	for remain > 0 {
		resp, err := a.volcCrClient.ListTags(
			&cr.ListTagsInput{
				Registry:   a.registryName,
				Namespace:  &namespace,
				Repository: &repo,
				PageSize:   &pageSize,
				PageNumber: &pageNumber,
			})
		if err != nil {
			return nil, err
		}
		if resp == nil || resp.TotalCount == nil {
			return nil, errListTagsResp
		}
		if initCondition {
			tagList = make([]string, 0, *resp.TotalCount)
			remain = *resp.TotalCount - pageSize
			initCondition = false
		} else {
			remain -= pageSize
		}
		pageNumber++
		for _, tagInfo := range resp.Items {
			tagList = append(tagList, *tagInfo.Name)
		}
	}
	return tagList, nil
}

func (a *adapter) listCandidateTags(namespace, repository, reference string) ([]*string, error) {
	if a.volcCrClient == nil {
		return []*string{}, errNilVolcCrClient
	}
	pageSize := MaxPageSize
	pageNumber := int64(1)
	initCondition := true
	var remain int64 = math.MaxInt64
	var tagList []*string
	desiredDig, err := digest.Parse(reference)
	if err != nil {
		return tagList, errPareseDigest
	}

	for remain > 0 {
		resp, err := a.volcCrClient.ListTags(
			&cr.ListTagsInput{
				Registry:   a.registryName,
				Namespace:  &namespace,
				Repository: &repository,
				PageSize:   &pageSize,
				PageNumber: &pageNumber,
			})
		if err != nil {
			return nil, err
		}
		if resp == nil || resp.TotalCount == nil {
			return nil, errPareseDigest
		}
		if initCondition {
			tagList = make([]*string, 0, *resp.TotalCount)
			remain = *resp.TotalCount - pageSize
			initCondition = false
		} else {
			remain -= pageSize
		}
		pageNumber++
		for _, tagInfo := range resp.Items {
			if tagInfo != nil && tagInfo.Name != nil && tagInfo.Digest != nil {
				dig, err := digest.Parse(*tagInfo.Digest)
				if err != nil {
					log.Debug("fail to parase digest", "tag", tagInfo)
					continue
				}
				if desiredDig.String() == dig.String() {
					tagList = append(tagList, tagInfo.Name)
				}
			}
		}
	}
	return tagList, nil
}

func (a *adapter) namespaceExist(namespace string) (bool, error) {
	if a.volcCrClient == nil {
		return false, errNilVolcCrClient
	}
	resp, err := a.volcCrClient.ListNamespaces(&cr.ListNamespacesInput{
		Registry: a.registryName,
		Filter: &cr.FilterForListNamespacesInput{
			Names: []*string{
				&namespace,
			},
		},
	})
	if err != nil {
		return false, err
	}
	if resp == nil || resp.TotalCount == nil {
		return false, errListNamespaceResp
	}
	if *resp.TotalCount > 0 {
		return true, nil
	}
	return false, nil
}
