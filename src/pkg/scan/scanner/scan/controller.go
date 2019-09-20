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

import "github.com/goharbor/harbor/src/pkg/scan/scanner/dao/scan"

// Options object for the scan action
type Options struct{}

// Option for scan action
type Option interface {
	// Apply option to the passing in options
	Apply(options *Options) error
}

// Controller defines operations for scan controlling
type Controller interface {
	// Scan the given artifact
	//
	//   Arguments:
	//     artifact *res.Artifact : artifact to be scanned
	//
	//   Returns:
	//     error  : non nil error if any errors occurred
	Scan(artifact *Artifact, options ...Option) error

	// GetReport gets the reports for the given artifact identified by the digest
	//
	//   Arguments:
	//     artifact *res.Artifact : the scanned artifact
	//
	//   Returns:
	//     []*scan.Report : scan results by different scanner vendors
	//     error          : non nil error if any errors occurred
	GetReport(artifact *Artifact) ([]*scan.Report, error)
}
