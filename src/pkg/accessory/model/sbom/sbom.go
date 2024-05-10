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

package sbom

import (
	"github.com/goharbor/harbor/src/pkg/accessory/model"
	"github.com/goharbor/harbor/src/pkg/accessory/model/base"
)

// HarborSBOM is the sbom accessory for harbor
type HarborSBOM struct {
	base.Default
}

// Kind gives the reference type of accessory.
func (c *HarborSBOM) Kind() string {
	return model.RefHard
}

// IsHard ...
func (c *HarborSBOM) IsHard() bool {
	return true
}

// New returns sbom accessory
func New(data model.AccessoryData) model.Accessory {
	return &HarborSBOM{base.Default{
		Data: data,
	}}
}

func init() {
	model.Register(model.TypeHarborSBOM, New)
}
