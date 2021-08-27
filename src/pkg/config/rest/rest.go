//  Copyright Project Harbor Authors
//
//  Licensed under the Apache License, Version 2.0 (the "License");
//  you may not use this file except in compliance with the License.
//  You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
//  Unless required by applicable law or agreed to in writing, software
//  distributed under the License is distributed on an "AS IS" BASIS,
//  WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
//  See the License for the specific language governing permissions and
//  limitations under the License.

package rest

import (
	"context"
	"errors"

	commonhttp "github.com/goharbor/harbor/src/common/http"
	"github.com/goharbor/harbor/src/common/http/modifier"
	"github.com/goharbor/harbor/src/lib/log"
)

// Driver - config store driver based on REST API
type Driver struct {
	configRESTURL string
	client        *commonhttp.Client
}

// NewRESTDriver - Create Driver
func NewRESTDriver(configRESTURL string, modifiers ...modifier.Modifier) *Driver {
	return &Driver{configRESTURL: configRESTURL, client: commonhttp.NewClient(nil, modifiers...)}
}

// Value ...
type Value struct {
	Val      interface{} `json:"value"`
	Editable bool        `json:"editable"`
}

// Load - load config data from REST server
func (h *Driver) Load(ctx context.Context) (map[string]interface{}, error) {
	cfgMap := map[string]interface{}{}
	origMap := map[string]*Value{}
	log.Infof("get configuration from url: %+v", h.configRESTURL)
	err := h.client.Get(h.configRESTURL, &origMap)
	if err != nil {
		log.Errorf("Failed on load rest config err:%v, url:%v", err, h.configRESTURL)
	}
	if len(origMap) < 1 {
		return cfgMap, errors.New("failed to load rest config")
	}
	for k, v := range origMap {
		cfgMap[k] = v.Val
	}
	return cfgMap, err
}

// Save - Save config data to REST server by PUT method
func (h *Driver) Save(ctx context.Context, cfg map[string]interface{}) error {
	return h.client.Put(h.configRESTURL, cfg)
}
