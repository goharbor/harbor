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
	"time"

	beego_orm "github.com/beego/beego/v2/client/orm"

	"github.com/goharbor/harbor/src/common/utils"
	"github.com/goharbor/harbor/src/lib/errors"
	"github.com/goharbor/harbor/src/lib/q"
)

func init() {
	beego_orm.RegisterModel(&Policy{})
}

const (
	// TriggerTypeManual represents a policy that only runs when requested.
	TriggerTypeManual TriggerType = "manual"
	// TriggerTypeScheduled represents a policy that runs on a cron schedule.
	TriggerTypeScheduled TriggerType = "scheduled"
)

// TriggerType represents the import policy trigger type.
type TriggerType = string

// Trigger stores manual or scheduled sync trigger settings.
type Trigger struct {
	Type     TriggerType `json:"type"`
	Settings struct {
		Cron string `json:"cron,omitempty"`
	} `json:"trigger_setting,omitempty"`
}

// Policy stores a project-scoped Hugging Face model import policy.
type Policy struct {
	ID                  int64          `orm:"pk;auto;column(id)" json:"id"`
	Name                string         `orm:"column(name)" json:"name"`
	Description         string         `orm:"column(description)" json:"description"`
	ProjectID           int64          `orm:"column(project_id)" json:"project_id"`
	TargetRepository    string         `orm:"column(target_repository)" json:"target_repository"`
	TargetTag           string         `orm:"column(target_tag)" json:"target_tag"`
	HuggingFaceRepoID   string         `orm:"column(hf_repo_id)" json:"hf_repo_id"`
	HuggingFaceRevision string         `orm:"column(hf_revision)" json:"hf_revision"`
	HuggingFaceSubpath  string         `orm:"column(hf_subpath)" json:"hf_subpath"`
	HuggingFaceToken    string         `orm:"column(hf_token)" json:"-"`
	PlainHFToken        string         `orm:"-" json:"hf_token,omitempty"`
	LastSyncedCommit    string         `orm:"column(last_synced_commit)" json:"last_synced_commit"`
	LastArtifactDigest  string         `orm:"column(last_artifact_digest)" json:"last_artifact_digest"`
	Creator             string         `orm:"column(creator)" json:"creator"`
	Enabled             bool           `orm:"column(enabled)" json:"enabled"`
	Trigger             *Trigger       `orm:"-" json:"trigger"`
	TriggerStr          string         `orm:"column(trigger)" json:"-"`
	ExtraAttrs          map[string]any `orm:"-" json:"extra_attrs"`
	ExtraAttrsStr       string         `orm:"column(extra_attrs)" json:"-"`
	CreationTime        time.Time      `orm:"column(creation_time);auto_now_add" json:"creation_time"`
	UpdateTime          time.Time      `orm:"column(update_time);auto_now" json:"update_time"`
}

// TableName specifies the database table.
func (p *Policy) TableName() string {
	return "model_import_policy"
}

// GetDefaultSorts specifies the default sorts.
func (p *Policy) GetDefaultSorts() []*q.Sort {
	return []*q.Sort{
		{Key: "UpdateTime"},
		{Key: "ID", DESC: true},
	}
}

// Validate validates the import policy.
func (p *Policy) Validate() error {
	if p == nil {
		return errors.New("nil model import policy")
	}
	if len(p.Name) == 0 {
		return errors.New(nil).WithCode(errors.BadRequestCode).WithMessage("policy name is required")
	}
	if p.ProjectID <= 0 {
		return errors.New(nil).WithCode(errors.BadRequestCode).WithMessage("project id is required")
	}
	if len(p.TargetRepository) == 0 {
		return errors.New(nil).WithCode(errors.BadRequestCode).WithMessage("target repository is required")
	}
	if len(p.TargetTag) == 0 {
		p.TargetTag = "latest"
	}
	if len(p.HuggingFaceRepoID) == 0 {
		return errors.New(nil).WithCode(errors.BadRequestCode).WithMessage("Hugging Face repository id is required")
	}
	if len(p.HuggingFaceRevision) == 0 {
		p.HuggingFaceRevision = "main"
	}
	if p.Trigger == nil {
		p.Trigger = &Trigger{Type: TriggerTypeManual}
	}
	if p.Trigger.Type == TriggerTypeScheduled {
		if len(p.Trigger.Settings.Cron) == 0 {
			return errors.New(nil).WithCode(errors.BadRequestCode).WithMessage("cron is required for scheduled model import")
		}
		if err := utils.ValidateCronString(p.Trigger.Settings.Cron); err != nil {
			return errors.New(nil).WithCode(errors.BadRequestCode).
				WithMessagef("invalid cron string for model import: %s, error: %v", p.Trigger.Settings.Cron, err)
		}
	}
	return nil
}

// Encode serializes complex policy fields before persistence.
func (p *Policy) Encode() error {
	if p.Trigger != nil {
		data, err := json.Marshal(p.Trigger)
		if err != nil {
			return err
		}
		p.TriggerStr = string(data)
	}
	if p.ExtraAttrs != nil {
		data, err := json.Marshal(p.ExtraAttrs)
		if err != nil {
			return err
		}
		p.ExtraAttrsStr = string(data)
	}
	return nil
}

// Decode deserializes complex policy fields after persistence.
func (p *Policy) Decode() error {
	if len(p.TriggerStr) > 0 {
		trigger := &Trigger{}
		if err := json.Unmarshal([]byte(p.TriggerStr), trigger); err != nil {
			return err
		}
		p.Trigger = trigger
	}
	if len(p.ExtraAttrsStr) > 0 {
		extras := map[string]any{}
		if err := json.Unmarshal([]byte(p.ExtraAttrsStr), &extras); err != nil {
			return err
		}
		p.ExtraAttrs = extras
	}
	return nil
}
