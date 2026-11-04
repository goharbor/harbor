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
	"encoding/json"
	"strings"
	"time"

	"github.com/beego/beego/v2/client/orm"

	regmodel "github.com/goharbor/harbor/src/pkg/reg/model"
)

func init() {
	orm.RegisterModel(new(Policy))
}

const (
	// TriggerTypeManual is the manual trigger
	TriggerTypeManual = "manual"
	// TriggerTypeScheduled is the scheduled (cron) trigger
	TriggerTypeScheduled = "scheduled"
)

// Policy is the ORM model for model sync policy
type Policy struct {
	ID                 int64     `orm:"pk;auto;column(id)"`
	Name               string    `orm:"column(name)"`
	Description        string    `orm:"column(description)"`
	Creator            string    `orm:"column(creator)"`
	Enabled            bool      `orm:"column(enabled)"`
	RegistryID         int64     `orm:"column(registry_id)"`
	SrcRepository      string    `orm:"column(src_repository)"`
	SrcRevision        string    `orm:"column(src_revision)"`
	FileFilter         string    `orm:"column(file_filter)"`
	DestProjectID      int64     `orm:"column(dest_project_id)"`
	DestRepository     string    `orm:"column(dest_repository)"`
	TriggerType        string    `orm:"column(trigger_type)"`
	Cron               string    `orm:"column(cron)"`
	LastSyncedRevision string    `orm:"column(last_synced_revision)"`
	CreationTime       time.Time `orm:"column(creation_time);auto_now_add" sort:"default:desc"`
	UpdateTime         time.Time `orm:"column(update_time);auto_now"`
}

// TableName set table name for ORM
func (p *Policy) TableName() string {
	return "model_sync_policy"
}

// GetFileFilter decodes the stored file filter, nil when none is set
func (p *Policy) GetFileFilter() (*regmodel.FileFilter, error) {
	if strings.TrimSpace(p.FileFilter) == "" {
		return nil, nil // nolint:nilnil // nil is the filter that selects every file
	}
	f := &regmodel.FileFilter{}
	if err := json.Unmarshal([]byte(p.FileFilter), f); err != nil {
		return nil, err
	}
	if f.IsEmpty() {
		return nil, nil // nolint:nilnil // nil is the filter that selects every file
	}
	return f, nil
}

// SetFileFilter encodes the file filter as JSON, empty when the filter selects everything
func (p *Policy) SetFileFilter(f *regmodel.FileFilter) error {
	if f.IsEmpty() {
		p.FileFilter = ""
		return nil
	}
	data, err := json.Marshal(f)
	if err != nil {
		return err
	}
	p.FileFilter = string(data)
	return nil
}

// DefaultDestRepository derives the destination repository from the source one:
// the source repository is lower-cased and the path separators are kept, e.g.
// "Qwen/Qwen3-8B" becomes "qwen/qwen3-8b".
func DefaultDestRepository(srcRepository string) string {
	return strings.ToLower(strings.Trim(strings.TrimSpace(srcRepository), "/"))
}
