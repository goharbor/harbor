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
	o := GetOrmer()
	sql := "insert into repository (owner_id, project_id, name, description, pull_count, star_count, creation_time, update_time) " +
		"select (select user_id as owner_id from user where username=?), " +
		"(select project_id as project_id from project where name=?), ?, ?, ?, ?, ?, NULL "

	_, err := o.Raw(sql, repo.OwnerName, repo.ProjectName, repo.Name, repo.Description,
		repo.PullCount, repo.StarCount, time.Now()).Exec()
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

//GetTopRepos returns the most popular repositories
func GetTopRepos(userID int, count int) ([]*models.RepoRecord, error) {
	sql :=
		`select r.repository_id, r.name, r.owner_id, 
			r.project_id, r.description, r.pull_count, 
			r.star_count, r.creation_time, r.update_time
		from repository r
		inner join project p on r.project_id = p.project_id
		where (
			p.deleted = 0 and (
				p.public = 1 or (
					? <> ? and (
						exists (
							select 1 from user u
							where u.user_id = ? and u.sysadmin_flag = 1
						) or exists (
							select 1 from project_member pm
							where pm.project_id = p.project_id and pm.user_id = ?
		)))))
		order by r.pull_count desc, r.name limit ?`
	repositories := []*models.RepoRecord{}
	_, err := GetOrmer().Raw(sql, userID, NonExistUserID, userID, userID, count).QueryRows(&repositories)

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

// GetTotalOfPublicRepositories ...
func GetTotalOfPublicRepositories(name string) (int64, error) {
	params := []interface{}{}
	sql := `select count(*) from repository r 
		join project p 
		on r.project_id = p.project_id and p.public = 1 `
	if len(name) != 0 {
		sql += ` where r.name like ?`
		params = append(params, "%"+escape(name)+"%")
	}

	var total int64
	err := GetOrmer().Raw(sql, params).QueryRow(&total)
	return total, err
}

// GetTotalOfUserRelevantRepositories ...
func GetTotalOfUserRelevantRepositories(userID int, name string) (int64, error) {
	params := []interface{}{}
	sql := `select count(*) 
		from repository r 
		join (
			select p.project_id, p.public 
				from project p
				join project_member pm
				on p.project_id = pm.project_id
				where pm.user_id = ?
		) as pp 
		on r.project_id = pp.project_id `
	params = append(params, userID)
	if len(name) != 0 {
		sql += ` where r.name like ?`
		params = append(params, "%"+escape(name)+"%")
	}

	var total int64
	err := GetOrmer().Raw(sql, params).QueryRow(&total)
	return total, err
}

// GetTotalOfRepositoriesByProject ...
func GetTotalOfRepositoriesByProject(projectID int64, name string) (int64, error) {
	qs := GetOrmer().QueryTable(&models.RepoRecord{}).
		Filter("ProjectID", projectID)

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
