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
	"github.com/vmware/harbor/models"

	"fmt"
	"strings"

	"github.com/astaxie/beego/orm"
)

func AddOrUpdateRepository(repository *models.Repository) (*models.Repository, error) {
	exists, _ := RepositoryExists(fmt.Sprintf("%s/%s", repository.ProjectName, repository.Name))
	if !exists {
		err := AddRepository(repository)
		if err != nil {
			return nil, err
		}
	} else {
		err := UpdateRepository(repository)
		if err != nil {
			return nil, err
		}
	}
	return repository, nil
}

func AddRepository(repository *models.Repository) error {
	o := orm.NewOrm()

	p, err := o.Raw("insert into repository(name, project_name, latest_tag) values (?, ?, ?)").Prepare()
	if err != nil {
		return err
	}

	r, err := p.Exec(repository.Name, repository.ProjectName, repository.LatestTag)
	if err != nil {
		return err
	}

	repositoryID, err := r.LastInsertId()
	if err != nil {
		return err
	}
	repository.Id = repositoryID
	return nil
}

func UpdateRepository(repository *models.Repository) error {
	o := orm.NewOrm()

	p, err := o.Raw("UPDATE repository SET latest_tag=? updated_at=now() WHERE name=? AND project_name=?").Prepare()
	if err != nil {
		return err
	}

	_, err = p.Exec(repository.LatestTag, repository.Name, repository.ProjectName)
	if err != nil {
		return err
	}

	return nil
}

//RepositoryExists returns whether the project exists according to its name of ID.
func RepositoryExists(nameOrID interface{}) (bool, error) {
	switch nameOrID.(type) {
	case int:
		repo, _ := GetRepositoryByID(nameOrID.(int64))
		if repo != nil {
			return true, nil
		}
	case string:
		repo, _ := GetRepositoryByName(nameOrID.(string))
		if repo != nil {
			return true, nil
		}
	}

	return false, nil
}

// GetRepositoryByID ...
func GetRepositoryByID(repositoryId int64) (*models.Repository, error) {
	o := orm.NewOrm()
	var repositories []models.Repository
	count, err := o.Raw("SELECT * from repository where id=? ", repositoryId).QueryRows(&repositories)
	if err != nil {
		return nil, err
	} else if count == 0 {
		return nil, nil
	} else {
		return &repositories[0], nil
	}
}

// GetRepositoryByName ...
func GetRepositoryByName(repoName string) (*models.Repository, error) {
	o := orm.NewOrm()
	projectName := strings.Split(repoName, "/")[0]
	repositoryName := strings.Split(repoName, "/")[1]
	var repositories []models.Repository
	count, err := o.Raw("SELECT * from repository where project_name=? AND name=? ", projectName, repositoryName).QueryRows(&repositories)
	if err != nil {
		return nil, err
	} else if count == 0 {
		return nil, nil
	} else {
		return &repositories[0], nil
	}
}
