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

package adapter

import (
	"errors"

	"github.com/goharbor/harbor/src/pkg/reg/model"
)

// SourceConfigurer is implemented by source adapters that honor model.SourceOptions. Callers
// configure the adapter right after creating it and before any other call.
type SourceConfigurer interface {
	ConfigureSource(opts *model.SourceOptions) error
}

// ConfigureSource hands opts to a source adapter. Empty options are a no-op; non-empty options
// fail on an adapter that does not implement SourceConfigurer, so a copy never silently
// ignores them.
func ConfigureSource(a any, opts *model.SourceOptions) error {
	if opts.IsEmpty() {
		return nil
	}
	c, ok := a.(SourceConfigurer)
	if !ok {
		return errors.New("the source adapter does not support source options")
	}
	return c.ConfigureSource(opts)
}
