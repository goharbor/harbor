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
	v2 "github.com/goharbor/harbor/src/server/v2.0/route"
)

// RegisterRoutes register all routes
func RegisterRoutes() {
	registerRoutes()          // service/internal API/UI controller/etc.
	registry.RegisterRoutes() // OCI registry APIs
	v2.RegisterRoutes()       // v2.0 APIs
}
