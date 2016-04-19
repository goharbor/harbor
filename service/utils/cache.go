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

package utils

import (
	"os"
	"time"

	"github.com/vmware/harbor/utils/log"
	"github.com/vmware/harbor/utils/registry"

	"github.com/astaxie/beego/cache"
)

// Cache is the global cache in system.
var Cache cache.Cache

var registryClient *registry.Registry

const catalogKey string = "catalog"

func init() {
	var err error
	Cache, err = cache.NewCache("memory", `{"interval":720}`)
	if err != nil {
		log.Errorf("Failed to initialize cache, error:%v", err)
	}

	endpoint := os.Getenv("REGISTRY_URL")
	client := registry.NewClientUsernameAuthHandlerEmbeded("admin")
	registryClient, err = registry.New(endpoint, client)
	if err != nil {
		log.Fatalf("error occurred while initializing authentication handler used by cache: %v", err)
	}
}

// RefreshCatalogCache calls registry's API to get repository list and write it to cache.
func RefreshCatalogCache() error {
	log.Debug("refreshing catalog cache...")
	rs, err := registryClient.Catalog()
	if err != nil {
		return err
	}

	repos := []string{}

	for _, repo := range rs {
		tags, err := registryClient.ListTag(repo)
		if err != nil {
			log.Errorf("error occurred while list tag for %s: %v", repo, err)
			return err
		}

		if len(tags) != 0 {
			repos = append(repos, repo)
			log.Debugf("add %s to catalog cache", repo)
		}
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
