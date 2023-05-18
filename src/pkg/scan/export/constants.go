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

package export

// CsvJobVendorID  specific type to be used in contexts
type CsvJobVendorID string

const (
	ProjectIDsAttribute    = "project_ids"
	JobNameAttribute       = "job_name"
	UserNameAttribute      = "user_name"
	StatusMessageAttribute = "status_message"
	// the scan data is a temporary file, use /tmp directory to avoid the permission issue.
	ScanDataExportDir  = "/tmp"
	QueryPageSize      = 100000
	ArtifactGroupSize  = 10000
	DigestKey          = "artifact_digest"
	CreateTimestampKey = "create_ts"
	Vendor             = "SCAN_DATA_EXPORT"
	CsvJobVendorIDKey  = CsvJobVendorID("vendorId")
)
