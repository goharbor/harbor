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

package aliacr

import (
	"encoding/json"
	"fmt"
	"strconv"
	"time"

	"github.com/aliyun/alibaba-cloud-sdk-go/sdk/requests"
	"github.com/aliyun/alibaba-cloud-sdk-go/services/cr"
	"github.com/aliyun/alibaba-cloud-sdk-go/services/cr_ee"
)

type repository struct {
	Name      string
	Namespace string
	ID        string
}

type authToken struct {
	user      string
	password  string
	expiresAt time.Time
}

// openapi is an interface that defines methods for interacting with an open API.
type openapi interface {
	// ListNamespace returns a list of all namespaces.
	ListNamespace() ([]string, error)
	// ListRepository returns a list of all repositories for a specified namespace.
	ListRepository(namespaceName string) ([]*repository, error)
	// ListRepoTag returns a list of all tags for a specified repository.
	ListRepoTag(repo *repository) ([]string, error)
	// GetAuthorizationToken returns the authorization token for repository.
	GetAuthorizationToken() (*authToken, error)
}

type acrOpenapi struct {
	client *cr.Client
	domain string
}

var _ openapi = &acrOpenapi{}

// newAcrOpenapi creates a new acrOpenapi instance.
func newAcrOpenapi(accessKeyID string, accessKeySecret string, regionID string) (openapi, error) {
	client, err := cr.NewClientWithAccessKey(regionID, accessKeyID, accessKeySecret)
	if err != nil {
		return nil, err
	}
	return &acrOpenapi{
		client: client,
		domain: fmt.Sprintf(endpointTpl, regionID),
	}, nil
}

// ListNamespace returns a list of namespaces
func (acr *acrOpenapi) ListNamespace() ([]string, error) {
	var namespaces []string
	nsReq := cr.CreateGetNamespaceListRequest()
	nsReq.SetDomain(acr.domain)
	nsResp, err := acr.client.GetNamespaceList(nsReq)
	if err != nil {
		return nil, err
	}
	var resp = &aliACRNamespaceResp{}
	err = json.Unmarshal(nsResp.GetHttpContentBytes(), resp)
	if err != nil {
		return nil, err
	}
	for _, ns := range resp.Data.Namespaces {
		namespaces = append(namespaces, ns.Namespace)
	}
	return namespaces, nil
}

// ListRepository returns a list of repositories in the specified namespace
func (acr *acrOpenapi) ListRepository(namespaceName string) ([]*repository, error) {
	var repos []*repository
	reposReq := cr.CreateGetRepoListByNamespaceRequest()
	reposReq.SetDomain(acr.domain)
	reposReq.RepoNamespace = namespaceName
	var page = 1
	for {
		reposReq.Page = requests.NewInteger(page)
		reposResp, err := acr.client.GetRepoListByNamespace(reposReq)
		if err != nil {
			return nil, err
		}
		var resp = &aliReposResp{}
		err = json.Unmarshal(reposResp.GetHttpContentBytes(), resp)
		if err != nil {
			return nil, err
		}

		for _, repo := range resp.Data.Repos {
			repos = append(repos, &repository{
				Name:      repo.RepoName,
				Namespace: repo.RepoNamespace,
			})
		}

		if resp.Data.Total <= (resp.Data.Page * resp.Data.PageSize) {
			break
		}
		page++
	}
	return repos, nil
}

// ListRepoTag returns a list of tags in the specified repository
func (acr *acrOpenapi) ListRepoTag(repo *repository) ([]string, error) {
	var tags []string
	tagsReq := cr.CreateGetRepoTagsRequest()
	tagsReq.SetDomain(acr.domain)
	tagsReq.RepoNamespace = repo.Namespace
	tagsReq.RepoName = repo.Name
	var page = 1
	for {
		tagsReq.Page = requests.NewInteger(page)
		tagsResp, err := acr.client.GetRepoTags(tagsReq)
		if err != nil {
			return nil, err
		}
		var resp = &aliTagResp{}
		err = json.Unmarshal(tagsResp.GetHttpContentBytes(), resp)
		if err != nil {
			return nil, err
		}
		for _, tag := range resp.Data.Tags {
			tags = append(tags, tag.Tag)
		}

		if resp.Data.Total <= (resp.Data.Page * resp.Data.PageSize) {
			break
		}
		page++
	}
	return tags, nil
}

// GetAuthorizationToken returns the authorization token for repository.
func (acr *acrOpenapi) GetAuthorizationToken() (*authToken, error) {
	tokenRequest := cr.CreateGetAuthorizationTokenRequest()
	tokenRequest.SetDomain(acr.domain)
	tokenResponse, err := acr.client.GetAuthorizationToken(tokenRequest)
	if err != nil {
		return nil, err
	}
	var v authorizationToken
	err = json.Unmarshal(tokenResponse.GetHttpContentBytes(), &v)
	if err != nil {
		return nil, err
	}
	return &authToken{
		user:      v.Data.TempUserName,
		password:  v.Data.AuthorizationToken,
		expiresAt: v.Data.ExpireDate.ToTime(),
	}, nil
}

type acreeOpenapi struct {
	client     *cr_ee.Client
	domain     string
	instanceID string
}

var _ openapi = &acreeOpenapi{}

// newAcreeOpenapi creates a new acreeOpenapi instance.
func newAcreeOpenapi(accessKeyID string, accessKeySecret string, regionID string, instanceID string) (openapi, error) {
	client, err := cr_ee.NewClientWithAccessKey(regionID, accessKeyID, accessKeySecret)
	if err != nil {
		return nil, err
	}
	return &acreeOpenapi{
		client:     client,
		domain:     fmt.Sprintf(endpointTpl, regionID),
		instanceID: instanceID,
	}, nil
}

// ListNamespace returns a list of namespaces
func (acree *acreeOpenapi) ListNamespace() ([]string, error) {
	var namespaces []string
	nsReq := cr_ee.CreateListNamespaceRequest()
	nsReq.SetDomain(acree.domain)
	nsReq.InstanceId = acree.instanceID
	page := 1
	for {
		nsReq.PageNo = requests.NewInteger(page)
		nsResp, err := acree.client.ListNamespace(nsReq)
		if err != nil {
			return nil, err
		}
		for _, ns := range nsResp.Namespaces {
			namespaces = append(namespaces, ns.NamespaceName)
		}
		if !nsResp.ListNamespaceIsSuccess {
			return nil, fmt.Errorf("failed to list namespace: %v", nsResp)
		}
		total, err := strconv.Atoi(nsResp.TotalCount)
		if err != nil {
			return nil, err
		}
		if total-(nsResp.PageNo*nsResp.PageSize) <= 0 {
			break
		}
		page++
	}
	return namespaces, nil
}

// ListRepository returns a list of repositories in the specified namespace
func (acree *acreeOpenapi) ListRepository(namespaceName string) ([]*repository, error) {
	var repos []*repository
	reposReq := cr_ee.CreateListRepositoryRequest()
	reposReq.SetDomain(acree.domain)
	reposReq.InstanceId = acree.instanceID
	reposReq.RepoNamespaceName = namespaceName
	reposReq.RepoStatus = "NORMAL"
	page := 1
	for {
		reposReq.PageNo = requests.NewInteger(page)
		reposResp, err := acree.client.ListRepository(reposReq)
		if err != nil {
			return nil, err
		}
		if !reposResp.ListRepositoryIsSuccess {
			return nil, fmt.Errorf("failed to list repo: %s", reposResp.GetHttpContentString())
		}
		for _, repo := range reposResp.Repositories {
			repos = append(repos, &repository{
				Name:      repo.RepoName,
				Namespace: repo.RepoNamespaceName,
				ID:        repo.RepoId,
			})
		}
		total, err := strconv.Atoi(reposResp.TotalCount)
		if err != nil {
			return nil, err
		}
		if total <= page*reposResp.PageSize {
			break
		}
		page++
	}
	return repos, nil
}

// ListRepoTag returns a list of tags in the specified repository
func (acree *acreeOpenapi) ListRepoTag(repo *repository) ([]string, error) {
	var tags []string
	tagsReq := cr_ee.CreateListRepoTagRequest()
	tagsReq.SetDomain(acree.domain)
	tagsReq.InstanceId = acree.instanceID
	tagsReq.RepoId = repo.ID
	var page = 1
	for {
		tagsReq.PageNo = requests.NewInteger(page)
		tagsResp, err := acree.client.ListRepoTag(tagsReq)
		if err != nil {
			return nil, err
		}
		for _, image := range tagsResp.Images {
			tags = append(tags, image.Tag)
		}
		if !tagsResp.ListRepoTagIsSuccess {
			return nil, fmt.Errorf("failed to list repo tag: %s", tagsResp.GetHttpContentString())
		}
		total, err := strconv.Atoi(tagsResp.TotalCount)
		if err != nil {
			return nil, err
		}
		if total <= page*tagsResp.PageSize {
			break
		}
		page++
	}
	return tags, nil
}

// GetAuthorizationToken returns the authorization token for repository.
func (acree *acreeOpenapi) GetAuthorizationToken() (*authToken, error) {
	tokenRequest := cr_ee.CreateGetAuthorizationTokenRequest()
	// FIXME: use vpc endpoint if vpc is enabled
	tokenRequest.SetDomain(acree.domain)
	tokenRequest.InstanceId = acree.instanceID
	tokenResponse, err := acree.client.GetAuthorizationToken(tokenRequest)
	if err != nil {
		return nil, err
	}
	if !tokenResponse.GetAuthorizationTokenIsSuccess {
		return nil, fmt.Errorf("failed to get authorization token: %s", tokenResponse.GetHttpContentString())
	}
	return &authToken{
		user:      tokenResponse.TempUsername,
		password:  tokenResponse.AuthorizationToken,
		expiresAt: time.Unix(tokenResponse.ExpireTime/1000, 0),
	}, nil
}
