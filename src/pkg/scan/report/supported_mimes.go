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

package report

import (
	"encoding/json"
	"reflect"

	"github.com/goharbor/harbor/src/lib/errors"
	v1 "github.com/goharbor/harbor/src/pkg/scan/rest/v1"
	"github.com/goharbor/harbor/src/pkg/scan/vuln"
)

// SupportedMimes indicates what mime types are supported to render at UI end.
var SupportedMimes = map[string]interface{}{
	// The native report type
	v1.MimeTypeNativeReport:               (*vuln.Report)(nil),
	v1.MimeTypeGenericVulnerabilityReport: (*vuln.Report)(nil),
}

// ResolveData is a helper func to parse the JSON data with the given mime type.
func ResolveData(mime string, jsonData []byte, options ...Option) (interface{}, error) {
	// If no resolver defined for the given mime types, directly ignore it.
	// The raw data will be used.
	t, ok := SupportedMimes[mime]
	if !ok {
		return nil, nil
	}

	if len(jsonData) == 0 {
		return nil, errors.New("empty JSON data")
	}

	ty := reflect.TypeOf(t)
	if ty.Kind() == reflect.Ptr {
		ty = ty.Elem()
	}

	// New one
	rp := reflect.New(ty).Elem().Addr().Interface()

	if err := json.Unmarshal(jsonData, rp); err != nil {
		return nil, err
	}

	ops := &Options{}
	for _, op := range options {
		op(ops)
	}

	if ops.ArtifactDigest != "" {
		if w, ok := rp.(interface{ WithArtifactDigest(string) }); ok {
			w.WithArtifactDigest(ops.ArtifactDigest)
		}
	}

	return rp, nil
}
