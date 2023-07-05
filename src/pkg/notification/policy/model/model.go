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

	"github.com/beego/beego/v2/client/orm"
)

func init() {
	orm.RegisterModel(&Policy{})
}

// Policy ...
type Policy struct {
	ID           int64         `orm:"pk;auto;column(id)" json:"id"`
	Name         string        `orm:"column(name)" json:"name"`
	Description  string        `orm:"column(description)" json:"description"`
	ProjectID    int64         `orm:"column(project_id)" json:"project_id"`
	TargetsDB    string        `orm:"column(targets)" json:"-"`
	Targets      []EventTarget `orm:"-" json:"targets"`
	EventTypesDB string        `orm:"column(event_types)" json:"-"`
	EventTypes   []string      `orm:"-" json:"event_types"`
	Creator      string        `orm:"column(creator)" json:"creator"`
	CreationTime time.Time     `orm:"column(creation_time);auto_now_add" json:"creation_time" sort:"default:desc"`
	UpdateTime   time.Time     `orm:"column(update_time);auto_now_add" json:"update_time"`
	Enabled      bool          `orm:"column(enabled)" json:"enabled"`
}

// TableName set table name for ORM.
func (w *Policy) TableName() string {
	return "notification_policy"
}

// ConvertToDBModel convert struct data in notification policy to DB model data
func (w *Policy) ConvertToDBModel() error {
	if len(w.Targets) != 0 {
		targets, err := json.Marshal(w.Targets)
		if err != nil {
			return err
		}
		w.TargetsDB = string(targets)
	}
	if len(w.EventTypes) != 0 {
		eventTypes, err := json.Marshal(w.EventTypes)
		if err != nil {
			return err
		}
		w.EventTypesDB = string(eventTypes)
	}

	return nil
}

// ConvertFromDBModel convert from DB model data to struct data
func (w *Policy) ConvertFromDBModel() error {
	targets := []EventTarget{}
	if len(w.TargetsDB) != 0 {
		err := json.Unmarshal([]byte(w.TargetsDB), &targets)
		if err != nil {
			return err
		}
	}
	w.Targets = targets

	types := []string{}
	if len(w.EventTypesDB) != 0 {
		err := json.Unmarshal([]byte(w.EventTypesDB), &types)
		if err != nil {
			return err
		}
	}
	w.EventTypes = types

	return nil
}

// EventTarget defines the structure of target a notification send to
type EventTarget struct {
	Type           string `json:"type"`
	Address        string `json:"address"`
	AuthHeader     string `json:"auth_header,omitempty"`
	SkipCertVerify bool   `json:"skip_cert_verify"`
	PayloadFormat  string `json:"payload_format,omitempty"`
}
