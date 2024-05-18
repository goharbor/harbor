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

const (
	// SBOMRepository ...
	SBOMRepository = "sbom_repository"
	// SBOMDigest ...
	SBOMDigest = "sbom_digest"
	// StartTime ...
	StartTime = "start_time"
	// EndTime ...
	EndTime = "end_time"
	// Duration ...
	Duration = "duration"
	// ScanStatus ...
	ScanStatus = "scan_status"
	// ReportID ...
	ReportID = "report_id"
	// Scanner ...
	Scanner = "scanner"
)

// Summary includes the sbom summary information
type Summary map[string]interface{}

// SBOMAccArt returns the repository and digest of the SBOM
func (s Summary) SBOMAccArt() (repo, digest string) {
	if repo, ok := s[SBOMRepository].(string); ok {
		if digest, ok := s[SBOMDigest].(string); ok {
			return repo, digest
		}
	}
	return "", ""
}
