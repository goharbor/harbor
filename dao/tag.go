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
	"log"
	"strings"

	"github.com/astaxie/beego/orm"
)

func AddOrUpdateTag(tag *models.Tag) (*models.Tag, error) {
	log.Println("inseting ")
	exists, _ := TagExists(fmt.Sprintf("%s/%s:%s", tag.ProjectName, tag.RepositoryName, tag.Version))
	if !exists {
		log.Println("exist ")
		err := AddTag(tag)
		if err != nil {
			log.Println(err)
			return nil, err
		}
	} else {
		log.Println("existxxx not ")

		err := UpdateTag(tag)
		if err != nil {
			return nil, err
		}
	}
	return tag, nil
}

func AddTag(tag *models.Tag) error {
	o := orm.NewOrm()

	p, err := o.Raw("insert into tag(project_id, repository_id, version, created_at, updated_at) values (?, ?, ?, now(), now())").Prepare()
	if err != nil {
		return err
	}

	r, err := p.Exec(tag.ProjectID, tag.RepositoryID, tag.Version)
	if err != nil {
		return err
	}

	tagID, err := r.LastInsertId()
	if err != nil {
		return err
	}
	tag.Id = tagID
	return nil
}

func UpdateTag(tag *models.Tag) error {
	o := orm.NewOrm()

	p, err := o.Raw("UPDATE tag SET updated_at=now() WHERE name=? AND repository_id=?").Prepare()
	if err != nil {
		return err
	}

	_, err = p.Exec(tag.Version, tag.RepositoryID)
	if err != nil {
		return err
	}

	return nil
}

//TagExists returns whether the project exists according to its name of ID.
func TagExists(nameOrID interface{}) (bool, error) {
	switch nameOrID.(type) {
	case int:
		repo, _ := GetTagByID(nameOrID.(int64))
		if repo != nil {
			return true, nil
		}
	case string:
		repo, _ := GetTagByName(nameOrID.(string))
		if repo != nil {
			return true, nil
		}
	}

	return false, nil
}

// GetTagByID ...
func GetTagByID(tagId int64) (*models.Tag, error) {
	o := orm.NewOrm()
	var tags []models.Tag
	count, err := o.Raw("SELECT * from tag where id=? ", tagId).QueryRows(&tags)
	if err != nil {
		return nil, err
	} else if count == 0 {
		return nil, nil
	} else {
		return &tags[0], nil
	}
}

// GetTagByName ...
func GetTagByName(str string) (*models.Tag, error) {
	o := orm.NewOrm()
	projectName := strings.Split(str, "/")[0]
	repoTagName := strings.Split(str, "/")[1]
	repositoryName := strings.Split(repoTagName, ":")[0]
	tagName := strings.Split(repoTagName, ":")[1]
	var tags []models.Tag
	count, err := o.Raw("SELECT * from tag where project_name=? AND repository_name = ? AND name=? ", projectName, repositoryName, tagName).QueryRows(&tags)
	if err != nil {
		return nil, err
	} else if count == 0 {
		return nil, nil
	} else {
		return &tags[0], nil
	}
}
