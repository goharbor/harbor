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

	"errors"
	"fmt"
	"log"
	"strings"

	"github.com/astaxie/beego/orm"
)

func AddOrUpdateRepository(repository *models.Repository) (*models.Repository, error) {
	repoFound, _ := RepositoryExists(fmt.Sprintf("%s/%s", repository.ProjectName, repository.Name))
	// find project id in case not exists
	if (*repository).Id == 0 {
		project, err := GetProjectByName(repository.ProjectName)
		if err != nil {
			return nil, err
		}
		(*repository).ProjectID = project.ProjectID
	}
	if repoFound != nil {
		log.Println("begin to update repository")
		repository.Id = repoFound.Id
		repoUpdated, err := UpdateRepository(repository)
		return repoUpdated, err
	} else {
		log.Println("begin to add repository")
		repoAdded, err := AddRepository(repository)
		return repoAdded, err
	}
	return repository, nil
}

func AddRepository(repository *models.Repository) (*models.Repository, error) {
	o := orm.NewOrm()

	p, err := o.Raw("insert into repository(name, project_name, latest_tag, project_id, user_name, created_at, updated_at) values (?, ?, ?, ?, ?, now(), now())").Prepare()
	if err != nil {
		return nil, err
	}

	r, err := p.Exec(repository.Name, repository.ProjectName, repository.LatestTag, repository.ProjectID, repository.UserName)
	if err != nil {
		return nil, err
	}

	repositoryID, err := r.LastInsertId()
	if err != nil {
		return nil, err
	}
	repository.Id = repositoryID
	return repository, nil
}

func UpdateRepoInfo(repository *models.Repository) (*models.Repository, error) {
	o := orm.NewOrm()

	p, err := o.Raw("UPDATE repository SET description =?, is_public=? , category = ? WHERE name=? AND project_name=?").Prepare()
	if err != nil {
		return nil, err
	}

	_, err = p.Exec(repository.Description, repository.IsPublic, repository.Category, repository.Name, repository.ProjectName)
	if err != nil {
		return nil, err
	}
	return repository, nil
}

func UpdateRepository(repository *models.Repository) (*models.Repository, error) {
	o := orm.NewOrm()

	p, err := o.Raw("UPDATE repository SET latest_tag=?, updated_at=now() WHERE name=? AND project_name=?").Prepare()
	if err != nil {
		return nil, err
	}

	_, err = p.Exec(repository.LatestTag, repository.Name, repository.ProjectName)
	if err != nil {
		return nil, err
	}

	return repository, nil
}

func RepositoriesUnderNamespace(namespace string) ([]models.Repository, error) {
	if namespace == "admin" {
		namespace = "library"
	}

	o := orm.NewOrm()
	sql := `SELECT name, description, project_id,  project_name, category, is_public, user_name, latest_tag, created_at, updated_at FROM repository WHERE is_public = 1 and project_name=?  ORDER BY updated_at DESC`

	var r []models.Repository
	_, err := o.Raw(sql, namespace).QueryRows(&r)
	if err != nil {
		return nil, err
	}
	return r, nil
}

//RepositoryExists returns whether the project exists according to its name of ID.
func RepositoryExists(nameOrID interface{}) (*models.Repository, error) {
	switch nameOrID.(type) {
	case int:
		repo, _ := GetRepositoryByID(nameOrID.(int64))
		if repo != nil {
			return repo, nil
		}
	case string:
		log.Println("nameOrnumber: ", nameOrID.(string))
		repo, _ := GetRepositoryByName(nameOrID.(string))
		if repo != nil {
			return repo, nil
		}
	}

	return nil, nil
}

// GetRepositoryByID ...
func GetRepositoryByID(repositoryId int64) (*models.Repository, error) {
	log.Println("GetRepositoryByID repoId: ", repositoryId)
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
	log.Println("GetRepositoryByName repoName: ", repoName)
	o := orm.NewOrm()
	projectName := strings.Split(repoName, "/")[0]
	repositoryName := strings.Split(repoName, "/")[1]
	var repositories []models.Repository

	sql := `select * from repository where project_name=?  and name = ?`
	count, err := o.Raw(sql, projectName, repositoryName).QueryRows(&repositories)
	if err != nil {
		return nil, err
	} else if count == 0 {
		return nil, errors.New("repo not found")
	} else {
		return &repositories[0], nil
	}
}
