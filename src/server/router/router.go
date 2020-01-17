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
	"errors"
	"github.com/astaxie/beego"
	beegocontext "github.com/astaxie/beego/context"
	"github.com/goharbor/harbor/src/server/middleware"
	"net/http"
)

type contextKeyInput struct{}

// NewRoute creates a new route
func NewRoute() *Route {
	return &Route{}
}

// Route stores the information that matches a request
type Route struct {
	methods     []string
	path        string
	middlewares []middleware.Middleware
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
	filterFunc := beego.FilterFunc(func(ctx *beegocontext.Context) {
		ctx.Request = ctx.Request.WithContext(
			context.WithValue(ctx.Request.Context(), contextKeyInput{}, ctx.Input))
		// TODO remove the WithMiddlewares?
		middleware.WithMiddlewares(handler, r.middlewares...).
			ServeHTTP(ctx.ResponseWriter, ctx.Request)
	})
	if len(r.methods) == 0 {
		beego.Any(r.path, filterFunc)
		return
	}
	for _, method := range r.methods {
		switch method {
		case http.MethodGet:
			beego.Get(r.path, filterFunc)
		case http.MethodHead:
			beego.Head(r.path, filterFunc)
		case http.MethodPut:
			beego.Put(r.path, filterFunc)
		case http.MethodPatch:
			beego.Patch(r.path, filterFunc)
		case http.MethodPost:
			beego.Post(r.path, filterFunc)
		case http.MethodDelete:
			beego.Delete(r.path, filterFunc)
		case http.MethodOptions:
			beego.Options(r.path, filterFunc)
		}
	}
}

// HandlerFunc sets the handler function that handles the request
func (r *Route) HandlerFunc(f http.HandlerFunc) {
	r.Handler(f)
}

// GetInput returns the input object from the context
func GetInput(context context.Context) (*beegocontext.BeegoInput, error) {
	if context == nil {
		return nil, errors.New("context is nil")
	}
	input, ok := context.Value(contextKeyInput{}).(*beegocontext.BeegoInput)
	if !ok {
		return nil, errors.New("input not found in the context")
	}
	return input, nil
}

// Param returns the router param by a given key from the context
func Param(ctx context.Context, key string) (string, error) {
	input, err := GetInput(ctx)
	if err != nil {
		return "", err
	}
	return input.Param(key), nil
}

// Middleware registers the global middleware that executed for all requests that match the path
func Middleware(path string, middleware middleware.Middleware) {
	// TODO add middleware function to register global middleware after upgrading to the latest version of beego
}
