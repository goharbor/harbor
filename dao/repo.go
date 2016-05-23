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
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/vmware/harbor/models"

	"github.com/astaxie/beego/orm"
	"github.com/vmware/harbor/utils/log"
)

// AddRepo adds a repo to the database.
func AddRepo(repo models.RepoRecord) (int64, error) {
	o := orm.NewOrm()
	sql := "insert into  repo (owner_id, project_id, name, creation_time, url, deleted, update_time, pull_count, star_count) " +
		"select (select user_id as owner_id from user where username=?), " +
		"(select project_id as project_id from project where name=?), ?, ?, ?, ?, ?, ?, ? "

	now := time.Now()
	r, err := o.Raw(sql, repo.OwnerName, repo.ProjectName, repo.Name, repo.Created, repo.Url, repo.Deleted, now, repo.PullCount, repo.StarCount).Exec()
	if err != nil {
		log.Errorf("error in AddRepo: %v ", err)
	}

	repoID, err := r.LastInsertId()
	if err != nil {
		return 0, err
	}

	return repoID, err
}

// DeleteRepo ...
func DeleteRepo(repoID int) error {
	o := orm.NewOrm()
	_, err := o.Raw(`update repo set deleted = 1 where repo_id = ?`, repoID).Exec()
	return err
}

// IncreasePullCount ...
func IncreasePullCount(repo models.RepoRecord) (err error) {

	o := orm.NewOrm()

	var r sql.Result
	r, err = o.Raw(`update repo set pull_count=pull_count+1 where name=?`, repo.Name).Exec()

	if err != nil {
		return err
	}
	c, err := r.RowsAffected()
	if err != nil {
		return err
	}
	if c == 0 {
		return errors.New("No record has been modified, increase pull count failed.")
	}

	return nil
}

//RepoExists returns whether the repo exists according to its name of ID.
func RepoExists(nameOrID interface{}) (bool, error) {
	o := orm.NewOrm()
	type dummy struct{}
	sql := `select repo_id from repo where deleted = 0 and `
	switch nameOrID.(type) {
	case int64:
		sql += `repo_id = ?`
	case string:
		sql += `name = ?`
	default:
		return false, fmt.Errorf("Invalid nameOrId: %v", nameOrID)
	}

	var d []dummy
	num, err := o.Raw(sql, nameOrID).QueryRows(&d)
	if err != nil {
		return false, err
	}
	return num > 0, nil

}

// GetRepoByID ...
func GetRepoByID(id int64) (*models.RepoRecord, error) {
	o := orm.NewOrm()

	sql := `select r.repo_id, r.name, u.username as owner_name, p.name as project_name, r.creation_time, r.url, r.deleted, r.update_time, r.pull_count, r.star_count   
			from repo as r 
			left join user as u on r.owner_id = u.user_id 
			left join project as p on r.project_id = p.project_id 
			where r.deleted = 0 and r.repo_id = ?`
	queryParam := make([]interface{}, 1)
	queryParam = append(queryParam, id)

	log.Debugf(sql)

	r := []models.RepoRecord{}
	count, err := o.Raw(sql, queryParam).QueryRows(&r)

	if err != nil {
		return nil, err
	}

	if count == 0 {
		return nil, nil
	}

	return &r[0], nil
}

// GetRepoByName ...
func GetRepoByName(name string) (*models.RepoRecord, error) {
	o := orm.NewOrm()
	var r []models.RepoRecord
	n, err := o.Raw(`select * from repo where name = ? and deleted = 0`, name).QueryRows(&r)
	if err != nil {
		return nil, err
	}

	if n == 0 {
		return nil, nil
	}

	return &r[0], nil
}
