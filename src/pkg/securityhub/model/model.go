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

import "github.com/goharbor/harbor/src/pkg/scan/dao/scan"

// Summary is the summary of scan result
type Summary struct {
	CriticalCnt        int64                       `json:"critical_cnt"`
	HighCnt            int64                       `json:"high_cnt"`
	MediumCnt          int64                       `json:"medium_cnt"`
	LowCnt             int64                       `json:"low_cnt"`
	NoneCnt            int64                       `json:"none_cnt"`
	UnknownCnt         int64                       `json:"unknown_cnt"`
	FixableCnt         int64                       `json:"fixable_cnt"`
	ScannedCnt         int64                       `json:"scanned_cnt"`
	NotScanCnt         int64                       `json:"not_scan_cnt"`
	TotalArtifactCnt   int64                       `json:"total_artifact_cnt"`
	DangerousCVEs      []*scan.VulnerabilityRecord `json:"dangerous_cves"`
	DangerousArtifacts []*DangerousArtifact        `json:"dangerous_artifacts"`
}

// DangerousArtifact define the most dangerous artifact
type DangerousArtifact struct {
	Project     int64  `json:"project" orm:"column(project)"`
	Repository  string `json:"repository" orm:"column(repository)"`
	Digest      string `json:"digest" orm:"column(digest)"`
	CriticalCnt int64  `json:"critical_cnt" orm:"column(critical_cnt)"`
	HighCnt     int64  `json:"high_cnt" orm:"column(high_cnt)"`
	MediumCnt   int64  `json:"medium_cnt" orm:"column(medium_cnt)"`
	LowCnt      int64  `json:"low_cnt" orm:"column(low_cnt)"`
}

// VulnerabilityItem is the item of vulnerability
type VulnerabilityItem struct {
	scan.VulnerabilityRecord
	ArtifactID     int64    `orm:"column(artifact_id)"`
	RepositoryName string   `orm:"column(repository_name)"`
	Digest         string   `orm:"column(digest)"`
	Tags           []string `orm:"-"`
	ProjectID      int64    `orm:"column(project_id)"`
}
