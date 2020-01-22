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
	"github.com/astaxie/beego"
	beegocontext "github.com/astaxie/beego/context"
	"github.com/goharbor/harbor/src/server/middleware"
	"net/http"
	"path/filepath"
)

type contextKeyInput struct{}

// NewRoute creates a new route
func NewRoute() *Route {
	return &Route{}
}

// Route stores the information that matches a request
type Route struct {
	parent      *Route
	methods     []string
	path        string
	middlewares []middleware.Middleware
}

// NewRoute returns a sub route based on the current one
func (r *Route) NewRoute() *Route {
	return &Route{
		parent: r,
	}
}

// Method sets the method that the route matches
func (r *Route) Method(method string) *Route {
	r.methods = append(r.methods, method)
	return r
}

// Path sets the path that the route matches. Path uses the beego router path pattern
func (r *Route) Path(path string) *Route {
	r.path = path
	return r
}

// Middleware sets the middleware that executed when handling the request
func (r *Route) Middleware(middleware middleware.Middleware) *Route {
	r.middlewares = append(r.middlewares, middleware)
	return r
}

// Handler sets the handler that handles the request
func (r *Route) Handler(handler http.Handler) {
	methods := r.methods
	if len(methods) == 0 && r.parent != nil {
		methods = r.parent.methods
	}

	path := r.path
	if r.parent != nil {
		path = filepath.Join(r.parent.path, path)
	}

	var middlewares []middleware.Middleware
	if r.parent != nil {
		middlewares = r.parent.middlewares
	}

	middlewares = append(middlewares, r.middlewares...)
	filterFunc := beego.FilterFunc(func(ctx *beegocontext.Context) {
		ctx.Request = ctx.Request.WithContext(
			context.WithValue(ctx.Request.Context(), contextKeyInput{}, ctx.Input))
		// TODO remove the WithMiddlewares?
		middleware.WithMiddlewares(handler, middlewares...).
			ServeHTTP(ctx.ResponseWriter, ctx.Request)
	})
	if len(methods) == 0 {
		beego.Any(r.path, filterFunc)
		return
	}
	for _, method := range methods {
		switch method {
		case http.MethodGet:
			beego.Get(path, filterFunc)
		case http.MethodHead:
			beego.Head(path, filterFunc)
		case http.MethodPut:
			beego.Put(path, filterFunc)
		case http.MethodPatch:
			beego.Patch(path, filterFunc)
		case http.MethodPost:
			beego.Post(path, filterFunc)
		case http.MethodDelete:
			beego.Delete(path, filterFunc)
		case http.MethodOptions:
			beego.Options(path, filterFunc)
		}
	}
}

// HandlerFunc sets the handler function that handles the request
func (r *Route) HandlerFunc(f http.HandlerFunc) {
	r.Handler(f)
}

// Param returns the beego router param by a given key from the context
func Param(ctx context.Context, key string) string {
	if ctx == nil {
		return ""
	}
	input, ok := ctx.Value(contextKeyInput{}).(*beegocontext.BeegoInput)
	if !ok {
		return ""
	}
	return input.Param(key)
}
