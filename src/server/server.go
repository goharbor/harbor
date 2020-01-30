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

package server

import (
	"github.com/goharbor/harbor/src/server/registry"
	// "github.com/goharbor/harbor/src/server/registry"
	v1 "github.com/goharbor/harbor/src/server/v1.0/route"
	v2 "github.com/goharbor/harbor/src/server/v2.0/route"
)

// RegisterRoutes register all routes
func RegisterRoutes() {
	// TODO move the v1 APIs to v2
	v1.RegisterRoutes()       // v1.0 APIs
	v2.RegisterRoutes()       // v2.0 APIs
	registry.RegisterRoutes() // OCI registry APIs
}
