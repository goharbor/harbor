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

package route

import (
	"github.com/goharbor/harbor/src/server/handler/base"
	"github.com/goharbor/harbor/src/server/handler/v2.1/handler"
	"github.com/goharbor/harbor/src/server/middleware/apiversion"
	"github.com/goharbor/harbor/src/server/router"
)

// RegisterRoutes for Harbor v2.1 APIs
func RegisterRoutes() {
	registerLegacyRoutes()
	router.NewRoute().Path("/api/" + base.APIVersionV21 + "/*").
		Middleware(apiversion.Middleware(base.APIVersionV21)).
		Handler(handler.New())
}
