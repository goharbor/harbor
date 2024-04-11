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

package model

// OverviewOptions define the option to query overview info
type OverviewOptions struct {
	WithVuln bool
	WithSBOM bool
}

// Option define the func to build options
type Option func(*OverviewOptions)

// NewOverviewOptions create a new OverviewOptions
func NewOverviewOptions(options ...Option) *OverviewOptions {
	opts := &OverviewOptions{}
	for _, f := range options {
		f(opts)
	}
	return opts
}

// WithVuln set the option to query vulnerability info
func WithVuln(enable bool) Option {
	return func(o *OverviewOptions) {
		o.WithVuln = enable
	}
}

// WithSBOM set the option to query SBOM info
func WithSBOM(enable bool) Option {
	return func(o *OverviewOptions) {
		o.WithSBOM = enable
	}
}
