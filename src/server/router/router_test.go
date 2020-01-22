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

package router

import (
	"context"
	beegocontext "github.com/astaxie/beego/context"
	"github.com/goharbor/harbor/src/server/middleware"
	"github.com/stretchr/testify/suite"
	"net/http"
	"testing"
)

type routerTestSuite struct {
	suite.Suite
	route *Route
}

func (r *routerTestSuite) SetupTest() {
	r.route = NewRoute()
}

func (r *routerTestSuite) TestNewRoute() {
	sub := r.route.Path("/v2").NewRoute()
	r.Require().NotNil(sub.parent)
	r.Equal("/v2", sub.parent.path)
}

func (r *routerTestSuite) TestMethod() {
	r.route.Method(http.MethodGet)
	r.Equal(http.MethodGet, r.route.methods[0])
	r.route.Method(http.MethodDelete)
	r.Equal(http.MethodDelete, r.route.methods[1])
}

func (r *routerTestSuite) TestPath() {
	r.route.Path("/api/*")
	r.Equal("/api/*", r.route.path)
}

func (r *routerTestSuite) TestMiddleware() {
	m1 := middleware.Middleware(func(handler http.Handler) http.Handler { return nil })
	m2 := middleware.Middleware(func(handler http.Handler) http.Handler { return nil })
	r.route.Middleware(m1)
	r.Len(r.route.middlewares, 1)
	r.route.Middleware(m2)
	r.Len(r.route.middlewares, 2)
}

func (r *routerTestSuite) TestParam() {
	// nil context
	value := Param(nil, "key")
	r.Empty(value)

	// context contains wrong type input
	value = Param(context.WithValue(context.Background(), contextKeyInput{}, &Route{}), "key")
	r.Empty(value)

	// success
	input := &beegocontext.BeegoInput{}
	input.SetParam("key", "value")
	value = Param(context.WithValue(context.Background(), contextKeyInput{}, input), "key")
	r.Equal("value", value)
}

func TestRouterTestSuite(t *testing.T) {
	suite.Run(t, &routerTestSuite{})
}
