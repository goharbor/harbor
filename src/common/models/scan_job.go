// Copyright (c) 2017 VMware, Inc. All Rights Reserved.
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

package models

import "time"

//ScanJobTable is the name of the table whose data is mapped by ScanJob struct.
const ScanJobTable = "img_scan_job"

//ScanOverviewTable is the name of the table whose data is mapped by ImgScanOverview struct.
const ScanOverviewTable = "img_scan_overview"

//ScanJob is the model to represent a job for image scan in DB.
type ScanJob struct {
	ID           int64     `orm:"pk;auto;column(id)" json:"id"`
	Status       string    `orm:"column(status)" json:"status"`
	Repository   string    `orm:"column(repository)" json:"repository"`
	Tag          string    `orm:"column(tag)" json:"tag"`
	Digest       string    `orm:"column(digest)" json:"digest"`
	CreationTime time.Time `orm:"column(creation_time);auto_now_add" json:"creation_time"`
	UpdateTime   time.Time `orm:"column(update_time);auto_now" json:"update_time"`
}

// Severity represents the severity of a image/component in terms of vulnerability.
type Severity int64

// Sevxxx is the list of severity of image after scanning.
const (
	_ Severity = iota
	SevNone
	SevUnknown
	SevLow
	SevMedium
	SevHigh
)

//TableName is required by by beego orm to map ScanJob to table img_scan_job
func (s *ScanJob) TableName() string {
	return ScanJobTable
}

//ImgScanOverview mapped to a record of image scan overview.
type ImgScanOverview struct {
	Digest          string              `orm:"pk;column(image_digest)" json:"image_digest"`
	Status          string              `orm:"-" json:"scan_status"`
	JobID           int64               `orm:"column(scan_job_id)" json:"job_id"`
	Sev             int                 `orm:"column(severity)" json:"severity"`
	CompOverviewStr string              `orm:"column(components_overview)" json:"-"`
	CompOverview    *ComponentsOverview `orm:"-" json:"components,omitempty"`
	DetailsKey      string              `orm:"column(details_key)" json:"details_key"`
	CreationTime    time.Time           `orm:"column(creation_time);auto_now_add" json:"creation_time,omitempty"`
	UpdateTime      time.Time           `orm:"column(update_time);auto_now" json:"update_time,omitempty"`
}

//TableName ...
func (iso *ImgScanOverview) TableName() string {
	return ScanOverviewTable
}

//ComponentsOverview has the total number and a list of components number of different serverity level.
type ComponentsOverview struct {
	Total   int                        `json:"total"`
	Summary []*ComponentsOverviewEntry `json:"summary"`
}

//ComponentsOverviewEntry ...
type ComponentsOverviewEntry struct {
	Sev   int `json:"severity"`
	Count int `json:"count"`
}
