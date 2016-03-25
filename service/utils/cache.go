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
	"encoding/json"
	"time"

	"github.com/vmware/harbor/models"
	log "github.com/vmware/harbor/utils/log"

	"github.com/astaxie/beego/cache"
)

// Cache is the global cache in system.
var Cache cache.Cache

const catalogKey string = "catalog"

func init() {
	var err error
	Cache, err = cache.NewCache("memory", `{"interval":720}`)
	if err != nil {
		log.Error("Failed to initialize cache, error:", err)
	}
}

// RefreshCatalogCache calls registry's API to get repository list and write it to cache.
func RefreshCatalogCache() error {
	result, err := RegistryAPIGet(BuildRegistryURL("_catalog"), "")
	if err != nil {
		return err
	}
	repoResp := models.Repo{}
	err = json.Unmarshal(result, &repoResp)
	if err != nil {
		return err
	}
	Cache.Put(catalogKey, repoResp.Repositories, 600*time.Second)
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
