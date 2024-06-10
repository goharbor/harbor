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

package scan

import v1 "github.com/goharbor/harbor/src/pkg/scan/rest/v1"

// Options keep the settings/configurations for scanning.
type Options struct {
	ExecutionID int64  // The execution id to scan artifact
	Tag         string // The tag of the artifact to scan
	ScanType    string // The scan type could be sbom or vulnerability
	FromEvent   bool   // indicate the current call from event or not
}

// GetScanType returns the scan type. for backward compatibility, the default type is vulnerability.
func (o *Options) GetScanType() string {
	if len(o.ScanType) == 0 {
		o.ScanType = v1.ScanTypeVulnerability
	}
	return o.ScanType
}

// Option represents an option item by func template.
// The validation result of the options are marked by nil/non-nil error.
// e.g:
// If the option is required and the input arg is empty,
// then a non nil error should be returned at then.
type Option func(options *Options) error

// WithExecutionID sets the execution id option.
func WithExecutionID(executionID int64) Option {
	return func(options *Options) error {
		options.ExecutionID = executionID

		return nil
	}
}

// WithTag sets the tag option.
func WithTag(tag string) Option {
	return func(options *Options) error {
		options.Tag = tag

		return nil
	}
}

// WithScanType set the scanType
func WithScanType(scanType string) Option {
	return func(options *Options) error {
		options.ScanType = scanType
		return nil
	}
}

// WithFromEvent set the caller's source
func WithFromEvent(fromEvent bool) Option {
	return func(options *Options) error {
		options.FromEvent = fromEvent
		return nil
	}
}
