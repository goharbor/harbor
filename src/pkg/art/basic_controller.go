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

package art

import (
	"github.com/goharbor/harbor/src/common/models"
	"github.com/goharbor/harbor/src/pkg/q"
)

// DefaultController for easy referring as a singleton instance
var DefaultController = NewController()

// basicController is the default implementation of controller
type basicController struct {
	m Manager
}

// NewController ...
func NewController() Controller {
	return &basicController{
		m: NewManager(),
	}
}

// List artifacts
func (b *basicController) List(query *q.Query) ([]*models.Artifact, error) {
	return b.m.List(query)
}
