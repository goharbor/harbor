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

package metadata

import (
	"sync"
)

var metaDataOnce sync.Once
var metaDataInstance *CfgMetaData

// Instance - Get Instance, make it singleton because there is only one copy of metadata in an env
func Instance() *CfgMetaData {
	metaDataOnce.Do(func() {
		metaDataInstance = newCfgMetaData()
		metaDataInstance.init()
	})
	return metaDataInstance
}

func newCfgMetaData() *CfgMetaData {
	return &CfgMetaData{metaMap: make(map[string]Item)}
}

// CfgMetaData ...
type CfgMetaData struct {
	metaMap map[string]Item
}

// init ...
func (c *CfgMetaData) init() {
	c.initFromArray(ConfigList)
}

// initFromArray - Initial metadata from an array
func (c *CfgMetaData) initFromArray(items []Item) {
	c.metaMap = make(map[string]Item)
	for _, item := range items {
		c.metaMap[item.Name] = item
	}
}

// GetByName - Get current metadata of current name, if not defined, return false in second params
func (c *CfgMetaData) GetByName(name string) (*Item, bool) {
	if item, ok := c.metaMap[name]; ok {
		return &item, true
	}
	return nil, false
}

// GetAll - Get all metadata in current env
func (c *CfgMetaData) GetAll() []Item {
	metaDataList := make([]Item, 0)
	for _, value := range c.metaMap {
		metaDataList = append(metaDataList, value)
	}
	return metaDataList
}
