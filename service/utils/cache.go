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

	"github.com/astaxie/beego"

	"github.com/astaxie/beego/cache"
)

var Cache cache.Cache

const CATALOG string = "catalog"

func init() {
	var err error
	Cache, err = cache.NewCache("memory", `{"interval":720}`)
	if err != nil {
		beego.Error("Failed to initialize cache, error:", err)
	}
}

func RefreshCatalogCache() error {
	result, err := RegistryApiGet(BuildRegistryUrl("_catalog"), "")
	if err != nil {
		return err
	}
	repoResp := models.Repo{}
	err = json.Unmarshal(result, &repoResp)
	if err != nil {
		return err
	}
	Cache.Put(CATALOG, repoResp.Repositories, 600*time.Second)
	return nil
}

func GetRepoFromCache() ([]string, error) {

	result := Cache.Get(CATALOG)
	if result == nil {
		err := RefreshCatalogCache()
		if err != nil {
			return nil, err
		}
		cached := Cache.Get(CATALOG)
		if cached != nil {
			return cached.([]string), nil
		}
		return nil, nil
	}
	return result.([]string), nil
}
