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

package orm

import (
	"net/http"

	o "github.com/astaxie/beego/orm"
	"github.com/goharbor/harbor/src/lib/orm"
	"github.com/goharbor/harbor/src/server/middleware"
)

// Config defines the config for Orm middleware.
type Config struct {
	// Creator defines a function to create ormer
	Creator func() o.Ormer
}

var (
	// DefaultConfig default orm config
	DefaultConfig = Config{
		Creator: func() o.Ormer {
			return o.NewOrm()
		},
	}
)

// Middleware middleware which add ormer to the http request context with default config
func Middleware(skippers ...middleware.Skipper) func(http.Handler) http.Handler {
	return MiddlewareWithConfig(DefaultConfig, skippers...)
}

// MiddlewareWithConfig middleware which add ormer to the http request context with config
func MiddlewareWithConfig(config Config, skippers ...middleware.Skipper) func(http.Handler) http.Handler {
	if config.Creator == nil {
		config.Creator = DefaultConfig.Creator
	}

	return middleware.New(func(w http.ResponseWriter, r *http.Request, next http.Handler) {
		ctx := orm.NewContext(r.Context(), config.Creator())
		next.ServeHTTP(w, r.WithContext(ctx))
	}, skippers...)
}
