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

package middlewares

import (
	"net/http"

	"github.com/goharbor/harbor/src/common/utils/log"
	"github.com/goharbor/harbor/src/core/middlewares/chart"
	"github.com/goharbor/harbor/src/core/middlewares/contenttrust"
	"github.com/goharbor/harbor/src/core/middlewares/countquota"
	"github.com/goharbor/harbor/src/core/middlewares/immutable"
	"github.com/goharbor/harbor/src/core/middlewares/listrepo"
	"github.com/goharbor/harbor/src/core/middlewares/multiplmanifest"
	"github.com/goharbor/harbor/src/core/middlewares/readonly"
	"github.com/goharbor/harbor/src/core/middlewares/sizequota"
	"github.com/goharbor/harbor/src/core/middlewares/url"
	"github.com/goharbor/harbor/src/core/middlewares/vulnerable"
	"github.com/justinas/alice"
)

// DefaultCreator ...
type DefaultCreator struct {
	middlewares []string
}

// New ...
func New(middlewares []string) *DefaultCreator {
	return &DefaultCreator{
		middlewares: middlewares,
	}
}

// Create creates a middleware chain ...
func (b *DefaultCreator) Create() *alice.Chain {
	chain := alice.New()
	for _, mName := range b.middlewares {
		middlewareName := mName
		chain = chain.Append(func(next http.Handler) http.Handler {
			constructor := b.geMiddleware(middlewareName)
			if constructor == nil {
				log.Errorf("cannot init middle %s", middlewareName)
				return nil
			}
			return constructor(next)
		})
	}
	return &chain
}

func (b *DefaultCreator) geMiddleware(mName string) alice.Constructor {
	middlewares := map[string]alice.Constructor{
		CHART:            func(next http.Handler) http.Handler { return chart.New(next) },
		READONLY:         func(next http.Handler) http.Handler { return readonly.New(next) },
		URL:              func(next http.Handler) http.Handler { return url.New(next) },
		MUITIPLEMANIFEST: func(next http.Handler) http.Handler { return multiplmanifest.New(next) },
		LISTREPO:         func(next http.Handler) http.Handler { return listrepo.New(next) },
		CONTENTTRUST:     func(next http.Handler) http.Handler { return contenttrust.New(next) },
		VULNERABLE:       func(next http.Handler) http.Handler { return vulnerable.New(next) },
		SIZEQUOTA:        func(next http.Handler) http.Handler { return sizequota.New(next) },
		COUNTQUOTA:       func(next http.Handler) http.Handler { return countquota.New(next) },
		IMMUTABLE:        func(next http.Handler) http.Handler { return immutable.New(next) },
	}
	return middlewares[mName]
}
