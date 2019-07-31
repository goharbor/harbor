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

package quota

import (
	"net/http"

	"github.com/goharbor/harbor/src/common/quota"
	"github.com/goharbor/harbor/src/pkg/types"
)

// Option ...
type Option func(*Options)

// Action ...
type Action string

const (
	// AddAction action to add resources
	AddAction Action = "add"
	// SubtractAction action to subtract resources
	SubtractAction Action = "subtract"
)

// Options ...
type Options struct {
	Action     Action
	Manager    *quota.Manager
	MutexKeys  []string
	Resources  types.ResourceList
	StatusCode int

	OnResources func(*http.Request) types.ResourceList
	OnFulfilled func(http.ResponseWriter, *http.Request) error
	OnRejected  func(http.ResponseWriter, *http.Request) error
	OnFinally   func(http.ResponseWriter, *http.Request) error
}

func newOptions(opt ...Option) Options {
	opts := Options{}

	for _, o := range opt {
		o(&opts)
	}

	if opts.Action == "" {
		opts.Action = AddAction
	}

	if opts.StatusCode == 0 {
		opts.StatusCode = http.StatusOK
	}

	return opts
}

// WithAction sets the interceptor action
func WithAction(a Action) Option {
	return func(o *Options) {
		o.Action = a
	}
}

// Manager sets the interceptor manager
func Manager(m *quota.Manager) Option {
	return func(o *Options) {
		o.Manager = m
	}
}

// WithManager sets the interceptor manager by reference and referenceID
func WithManager(reference, referenceID string) Option {
	return func(o *Options) {
		m, err := quota.NewManager(reference, referenceID)
		if err != nil {
			return
		}

		o.Manager = m
	}
}

// MutexKeys set the interceptor mutex keys
func MutexKeys(keys ...string) Option {
	return func(o *Options) {
		o.MutexKeys = keys
	}
}

// Resources set the interceptor resources
func Resources(r types.ResourceList) Option {
	return func(o *Options) {
		o.Resources = r
	}
}

// StatusCode set the interceptor status code
func StatusCode(c int) Option {
	return func(o *Options) {
		o.StatusCode = c
	}
}

// OnResources sets the interceptor on resources function
func OnResources(f func(*http.Request) types.ResourceList) Option {
	return func(o *Options) {
		o.OnResources = f
	}
}
