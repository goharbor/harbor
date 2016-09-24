/*
   Copyright (c) 2016 VMware, Inc. All Rights Reserved.
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

package cache

import (
	"time"

	"github.com/vmware/harbor/dao"
	"github.com/vmware/harbor/utils/log"
	"github.com/vmware/harbor/utils/registry"
	"github.com/vmware/harbor/utils/registry/auth"

	"github.com/astaxie/beego/cache"
)

var (
	// Cache is the global cache in system.
	Cache cache.Cache
)

const catalogKey string = "catalog"

func init() {
	var err error
	Cache, err = cache.NewCache("memory", `{"interval":720}`)
	if err != nil {
		log.Errorf("Failed to initialize cache, error:%v", err)
	}
}

// RefreshCatalogCache calls registry's API to get repository list and write it to cache.
func RefreshCatalogCache() error {
	log.Debug("refreshing catalog cache...")

	repos, err := getAllRepositories()
	if err != nil {
		return err
	}
	Cache.Put(catalogKey, repos, 600*time.Second)
	return nil
}

// GetRepoFromCache get repository list from cache, it refreshes the cache if it's empty.
func GetRepoFromCache() ([]string, error) {

	result := Cache.Get(catalogKey)
	if result == nil {
		err := RefreshCatalogCache()
		if err != nil {
			return nil, err
		}
		cached := Cache.Get(catalogKey)
		if cached != nil {
			return cached.([]string), nil
		}
		return nil, nil
	}
	return result.([]string), nil
}

func getAllRepositories() ([]string, error) {
	var repos []string
	rs, err := dao.GetAllRepositories()
	if err != nil {
		return repos, err
	}
	for _, e := range rs {
		repos = append(repos, e.Name)
	}
	return repos, nil
}

// NewRegistryClient ...
func NewRegistryClient(endpoint string, insecure bool, username, scopeType, scopeName string,
	scopeActions ...string) (*registry.Registry, error) {
	authorizer := auth.NewUsernameTokenAuthorizer(username, scopeType, scopeName, scopeActions...)

	store, err := auth.NewAuthorizerStore(endpoint, insecure, authorizer)
	if err != nil {
		return nil, err
	}

	client, err := registry.NewRegistryWithModifiers(endpoint, insecure, store)
	if err != nil {
		return nil, err
	}
	return client, nil
}

// NewRepositoryClient ...
func NewRepositoryClient(endpoint string, insecure bool, username, repository, scopeType, scopeName string,
	scopeActions ...string) (*registry.Repository, error) {

	authorizer := auth.NewUsernameTokenAuthorizer(username, scopeType, scopeName, scopeActions...)

	store, err := auth.NewAuthorizerStore(endpoint, insecure, authorizer)
	if err != nil {
		return nil, err
	}

	client, err := registry.NewRepositoryWithModifiers(repository, endpoint, insecure, store)
	if err != nil {
		return nil, err
	}
	return client, nil
}
