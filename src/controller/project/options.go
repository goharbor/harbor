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

package project

// Option option for `Get` and `Exist` method of `Controller`
type Option func(*Options)

// Options options used by `Get` method of `Controller`
type Options struct {
	WithDetail             bool
	WithCVEAllowlist       bool // get project with cve allowlist
	WithEffectCVEAllowlist bool // get project with effect cve allowlist
	WithMetadata           bool // get project with metadata
	WithOwner              bool // get project with owner name
}

// Detail set WithDetail for the Options
func Detail(detail bool) Option {
	return func(opts *Options) {
		opts.WithDetail = detail
	}
}

// WithCVEAllowlist set WithCVEAllowlist for the Options
func WithCVEAllowlist() Option {
	return func(opts *Options) {
		opts.WithCVEAllowlist = true
	}
}

// WithEffectCVEAllowlist set WithEffectCVEAllowlist for the Options
func WithEffectCVEAllowlist() Option {
	return func(opts *Options) {
		opts.WithMetadata = true // we need `reuse_sys_cve_allowlist` value in the metadata of project
		opts.WithEffectCVEAllowlist = true
	}
}

// Metadata set WithMetadata for the Options
func Metadata(metadata bool) Option {
	return func(opts *Options) {
		opts.WithMetadata = metadata
	}
}

// WithOwner set WithOwner for the Options
func WithOwner() Option {
	return func(opts *Options) {
		opts.WithOwner = true
	}
}

func newOptions(options ...Option) *Options {
	opts := &Options{
		WithDetail:   true, // default get project details
		WithMetadata: true, // default get project with metadata
	}

	for _, f := range options {
		f(opts)
	}

	return opts
}
