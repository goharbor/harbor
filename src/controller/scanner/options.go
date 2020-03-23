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

package scanner

// Options keep the settings/configurations for scanner.
type Options struct {
	// Mark the scan triggered by who.
	// Identified by the UUID.
	Ping bool
}

// Option represents an option item by func template.
// The validation result of the options are marked by nil/non-nil error.
// e.g:
// If the option is required and the input arg is empty,
// then a non nil error should be returned at then.
type Option func(options *Options) error

// WithPing sets the requester option.
func WithPing(ping bool) Option {
	return func(options *Options) error {
		options.Ping = ping

		return nil
	}
}

func newOptions(options ...Option) *Options {
	opts := &Options{Ping: true}

	for _, o := range options {
		o(opts)
	}

	return opts
}
