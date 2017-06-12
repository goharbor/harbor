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

package dao

import (
	"fmt"
	"time"

	"github.com/astaxie/beego/orm"
	"github.com/vmware/harbor/src/common/models"
)

// AddRepository adds a repo to the database.
func AddRepository(repo models.RepoRecord) error {
	if repo.ProjectID == 0 {
		return fmt.Errorf("invalid project ID: %d", repo.ProjectID)
	}

	o := GetOrmer()
	now := time.Now()
	repo.CreationTime = now
	repo.UpdateTime = now
	_, err := o.Insert(&repo)
	return err
}

// GetRepositoryByName ...
func GetRepositoryByName(name string) (*models.RepoRecord, error) {
	o := GetOrmer()
	r := models.RepoRecord{Name: name}
	err := o.Read(&r, "Name")
	if err == orm.ErrNoRows {
		return nil, nil
	}
	return &r, err
}

// GetAllRepositories ...
func GetAllRepositories() ([]models.RepoRecord, error) {
	o := GetOrmer()
	var repos []models.RepoRecord
	_, err := o.QueryTable("repository").
		OrderBy("Name").All(&repos)
	return repos, err
}

// DeleteRepository ...
func DeleteRepository(name string) error {
	o := GetOrmer()
	_, err := o.QueryTable("repository").Filter("name", name).Delete()
	return err
}

// UpdateRepository ...
func UpdateRepository(repo models.RepoRecord) error {
	o := GetOrmer()
	repo.UpdateTime = time.Now()
	_, err := o.Update(&repo)
	return err
}

// IncreasePullCount ...
func IncreasePullCount(name string) (err error) {
	o := GetOrmer()
	num, err := o.QueryTable("repository").Filter("name", name).Update(
		orm.Params{
			"pull_count":  orm.ColValue(orm.ColAdd, 1),
			"update_time": time.Now(),
		})
	if num == 0 {
		err = fmt.Errorf("Failed to increase repository pull count with name: %s %s", name, err.Error())
	}
	return err
}

//RepositoryExists returns whether the repository exists according to its name.
func RepositoryExists(name string) bool {
	o := GetOrmer()
	return o.QueryTable("repository").Filter("name", name).Exist()
}

// GetRepositoryByProjectName ...
func GetRepositoryByProjectName(name string) ([]*models.RepoRecord, error) {
	sql := `select * from repository 
		where project_id = (
			select project_id from project
			where name = ?
		)`
	repos := []*models.RepoRecord{}
	_, err := GetOrmer().Raw(sql, name).QueryRows(&repos)
	return repos, err
}

//GetTopRepos returns the most popular repositories whose project ID is
// in projectIDs
func GetTopRepos(projectIDs []int64, n int) ([]*models.RepoRecord, error) {
	repositories := []*models.RepoRecord{}
	_, err := GetOrmer().QueryTable(&models.RepoRecord{}).
		Filter("project_id__in", projectIDs).
		OrderBy("-pull_count").
		Limit(n).
		All(&repositories)

	return repositories, err
}

// GetTotalOfRepositories ...
func GetTotalOfRepositories(name string) (int64, error) {
	qs := GetOrmer().QueryTable(&models.RepoRecord{})
	if len(name) != 0 {
		qs = qs.Filter("Name__contains", name)
	}
	return qs.Count()
}

// GetTotalOfRepositoriesByProject ...
func GetTotalOfRepositoriesByProject(projectIDs []int64, name string) (int64, error) {
	if len(projectIDs) == 0 {
		return 0, nil
	}

	qs := GetOrmer().QueryTable(&models.RepoRecord{}).
		Filter("project_id__in", projectIDs)

	if len(name) != 0 {
		qs = qs.Filter("Name__contains", name)
	}

	return qs.Count()
}

// GetRepositoriesByProject ...
func GetRepositoriesByProject(projectID int64, name string,
	limit, offset int64) ([]*models.RepoRecord, error) {

	repositories := []*models.RepoRecord{}

	qs := GetOrmer().QueryTable(&models.RepoRecord{}).
		Filter("ProjectID", projectID)

	if len(name) != 0 {
		qs = qs.Filter("Name__contains", name)
	}

	_, err := qs.Limit(limit).
		Offset(offset).All(&repositories)

	return repositories, err
}
