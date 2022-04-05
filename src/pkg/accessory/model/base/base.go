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

package base

import (
	"github.com/goharbor/harbor/src/pkg/accessory/model"
)

var _ model.Accessory = (*Default)(nil)

// Default default model with TypeNone and RefNone
type Default struct {
	Data model.AccessoryData
}

// Kind ...
func (a *Default) Kind() string {
	return model.RefNone
}

// IsSoft ...
func (a *Default) IsSoft() bool {
	return false
}

// IsHard ...
func (a *Default) IsHard() bool {
	return false
}

// Display ...
func (a *Default) Display() bool {
	return false
}

// GetData ...
func (a *Default) GetData() model.AccessoryData {
	return a.Data
}

// New returns base
func New(data model.AccessoryData) model.Accessory {
	return &Default{Data: data}
}

func init() {
	model.Register(model.TypeNone, New)
}
