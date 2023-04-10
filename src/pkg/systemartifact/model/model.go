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

import (
	"time"

	"github.com/goharbor/harbor/src/lib/orm"
)

func init() {
	orm.RegisterModel(
		new(SystemArtifact),
	)
}

// SystemArtifact represents a tracking record for each system artifact that has been
// created within registry using the system artifact manager
type SystemArtifact struct {
	ID int64 `orm:"pk;auto;column(id)"`
	// the name of repository associated with the artifact
	Repository string `orm:"column(repository)"`
	// the SHA-256 digest of the artifact data.
	Digest string `orm:"column(digest)"`
	// the size of the artifact data in bytes
	Size int64 `orm:"column(size)"`
	// the harbor subsystem that created the artifact
	Vendor string `orm:"column(vendor)"`
	// the type of the system artifact.
	// the type field specifies the type of artifact data and is useful when a harbor component generates more than one
	// kind of artifact. for e.g. a scan data export job could create a detailed CSV export data file as well
	// as an summary export file. here type could be set to "CSVDetail" and "ScanSummary"
	Type string `orm:"column(type)"`
	// the time of creation of the system artifact
	CreateTime time.Time `orm:"column(create_time)"`
	// optional extra attributes for the system artifact
	ExtraAttrs string `orm:"column(extra_attrs)"`
}

func (sa *SystemArtifact) TableName() string {
	return "system_artifact"
}

func (sa *SystemArtifact) TableUnique() [][]string {
	return [][]string{{"vendor", "repository_name", "digest"}}
}
