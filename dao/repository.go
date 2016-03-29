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

	//"errors"
	//"fmt"
	//"time"

	//"github.com/astaxie/beego"
	//"github.com/astaxie/beego/orm"
)

func AddRepository(project models.Repository) error {
	return nil
}

func QueryRepository(query models.Repository) ([]models.Repository, error) {
	o := orm.NewOrm()

	sql := `select name, description, project_name, icon, latest_tag ,updatedAt from repository `

	queryParam := make([]interface{}, 1)
	if query.Name != "" {
		sql += " where name like ? "
		queryParam = append(queryParam, query.Name)
	}

	sql += " order by updatedAt desc"

	var r []models.Repository
	_, err := o.Raw(sql, queryParam).QueryRows(&r)

	if err != nil {
		return nil, err
	}
	return r, nil
}

//RepositoryExists returns whether the project exists according to its name of ID.
func RepositoryExists(nameOrID interface{}) (bool, error) {
	return false, nil
}

// GetRepositoryByID ...
func GetRepositoryByID(projectID int64) (*models.Repository, error) {
	return nil, nil
}

// GetRepositoryByName ...
func GetRepositoryByName(repositoryName string) (models.Repository, error) {
	o := orm.NewOrm()

	sql := `select name, description, project_name, icon, latest_tag ,updatedAt from repository `

	queryParam := make([]interface{}, 1)
	if repositoryName == "" {
		sql += " where name = ? "
		queryParam = append(queryParam, repositoryName)
	}

	sql += " order by updatedAt desc"

	var r models.Repository
	_, err := o.Raw(sql, queryParam).QueryRows(&r)

	return r, err
}
