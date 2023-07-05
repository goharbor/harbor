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

package formats

import (
	"context"
	"net/http"

	"github.com/goharbor/harbor/src/lib/errors"
	"github.com/goharbor/harbor/src/pkg/notifier/model"
)

// formatsRegistry is the service registry for formats.
var formatsRegistry map[string]Formatter

// registerFormats registers the format to formatsRegistry.
func registerFormats(formatType string, formatter Formatter) {
	if formatsRegistry == nil {
		formatsRegistry = make(map[string]Formatter)
	}

	formatsRegistry[formatType] = formatter
}

// Formatter is the interface for event which for implementing different drivers to
// organize their customize data format.
type Formatter interface {
	// Format formats the data to expected format and return request headers and encoded payload
	Format(context.Context, *model.HookEvent) (http.Header, []byte, error)
}

// GetFormatter returns corresponding formatter from format type.
func GetFormatter(formatType string) (Formatter, error) {
	if formatter, ok := formatsRegistry[formatType]; ok {
		return formatter, nil
	}

	return nil, errors.Errorf("unknown format type: %s", formatType)
}
