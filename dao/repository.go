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

package dao

import (
	"fmt"

	"github.com/astaxie/beego/orm"
	"github.com/vmware/harbor/models"
)

// AddRepository adds a repo to the database.
func AddRepository(repo models.RepoRecord) error {
	o := GetOrmer()
	sql := "insert into repository (owner_id, project_id, name, description, pull_count, star_count, creation_time, update_time) " +
		"select (select user_id as owner_id from user where username=?), " +
		"(select project_id as project_id from project where name=?), ?, ?, ?, ?, NOW(), NULL "

	_, err := o.Raw(sql, repo.OwnerName, repo.ProjectName, repo.Name, repo.Description, repo.PullCount, repo.StarCount).Exec()
	return err
}

// GetRepositoryByName ...
func GetRepositoryByName(name string) (*models.RepoRecord, error) {
	o := GetOrmer()
	r := models.RepoRecord{Name: name}
	err := o.Read(&r)
	if err == orm.ErrNoRows {
		return nil, nil
	}
	return &r, err
}

// GetAllRepositories ...
func GetAllRepositories() ([]models.RepoRecord, error) {
	o := GetOrmer()
	var repos []models.RepoRecord
	_, err := o.QueryTable("repository").All(&repos)
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
	_, err := o.Update(&repo)
	return err
}

// IncreasePullCount ...
func IncreasePullCount(name string) (err error) {
	o := GetOrmer()
	num, err := o.QueryTable("repository").Filter("name", name).Update(
		orm.Params{
			"pull_count": orm.ColValue(orm.ColAdd, 1),
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
