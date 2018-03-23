// Copyright (c) 2017 VMware, Inc. All Rights Reserved.
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

package api

import (
	"encoding/json"
	"io/ioutil"
	"net/http"

	"github.com/vmware/harbor/src/adminserver/systemcfg"
	"github.com/vmware/harbor/src/common/utils/log"
)

// ListCfgs lists configurations
func ListCfgs(w http.ResponseWriter, r *http.Request) {
	cfg, err := systemcfg.CfgStore.Read()
	if err != nil {
		log.Errorf("failed to get system configurations: %v", err)
		handleInternalServerError(w)
		return
	}

	if err = writeJSON(w, cfg); err != nil {
		log.Errorf("failed to write response: %v", err)
		return
	}
}

// UpdateCfgs updates configurations
func UpdateCfgs(w http.ResponseWriter, r *http.Request) {
	b, err := ioutil.ReadAll(r.Body)
	if err != nil {
		log.Errorf("failed to read request body: %v", err)
		handleInternalServerError(w)
		return
	}

	m := map[string]interface{}{}
	if err = json.Unmarshal(b, &m); err != nil {
		handleBadRequestError(w, err.Error())
		return
	}

	if err = systemcfg.CfgStore.Write(m); err != nil {
		log.Errorf("failed to update system configurations: %v", err)
		handleInternalServerError(w)
		return
	}
}

// ResetCfgs resets configurations from environment variables
func ResetCfgs(w http.ResponseWriter, r *http.Request) {
	cfgs := map[string]interface{}{}
	if err := systemcfg.LoadFromEnv(cfgs, true); err != nil {
		log.Errorf("failed to reset system configurations: %v", err)
		handleInternalServerError(w)
		return
	}

	if err := systemcfg.CfgStore.Write(cfgs); err != nil {
		log.Errorf("failed to write system configurations to storage: %v", err)
		handleInternalServerError(w)
		return
	}
}
