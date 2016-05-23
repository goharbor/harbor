/*
   Copyright (c) 2016 VMware, Inc. All Rights Reserved.
   Licensed under the Apache License, Version 2.0 (the "License");
   you may not use this file except in compliance with the License.
   You may obtain a copy of the License at

       http://www.apache.org/licenses/LICENSE-2.0

   Unless required by applicable law or agreed to in writing, software
   distributed under the License is distributed on an "AS IS" BASIS,
   WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
   See the License for the specific language governing permissions and
   limitations under the License.
*/

package models

import (
	"time"
)

// Repo holds information about repositories.
type Repo struct {
	Repositories []string `json:"repositories"`
}

// RepoItem holds manifest of an image.
type RepoItem struct {
	ID            string    `json:"Id"`
	Parent        string    `json:"Parent"`
	Created       time.Time `json:"Created"`
	DurationDays  string    `json:"Duration Days"`
	Author        string    `json:"Author"`
	Architecture  string    `json:"Architecture"`
	DockerVersion string    `json:"Docker Version"`
	Os            string    `json:"OS"`
	//Size           int       `json:"Size"`
}

// RepoRecord holds the record of an Repo in DB, all the infors are from the registry notification event.
type RepoRecord struct {
	ID          int64  `orm:"column(repo_id)" json:"RepoId"`
	Name        string `orm:"column(name)" json:"Name"`
	OwnerName   string
	OwnerID     int64 `orm:"column(owner_id)"  json:"OwnerId"`
	ProjectName string
	ProjectID   int64     `orm:"column(project_id)"  json:"ProjectId"`
	Created     time.Time `orm:"column(creation_time)"`
	Url         string    `orm:"column(url)" json:"Url"`
	Deleted     int       `orm:"column(deleted)"`
	UpdateTime  time.Time `orm:"update_time" json:"update_time"`
	PullCount   int       `orm:"column(pull_count)"`
	StarCount   int       `orm:"column(star_count)"`
}

// Tag holds information about a tag.
type Tag struct {
	Version string `json:"version"`
	ImageID string `json:"image_id"`
}

// Manifest ...
type Manifest struct {
	SchemaVersion int           `json:"schemaVersion"`
	Name          string        `json:"name"`
	Tag           string        `json:"tag"`
	Architecture  string        `json:"architecture"`
	FsLayers      []blobSumItem `json:"fsLayers"`
	History       []histroyItem `json:"history"`
}

type histroyItem struct {
	V1Compatibility string `json:"v1Compatibility"`
}

type blobSumItem struct {
	BlobSum string `json:"blobSum"`
}
