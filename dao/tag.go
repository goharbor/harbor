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
	"log"

	"github.com/astaxie/beego/orm"
	"github.com/vmware/harbor/models"
)

func AddOrUpdateTag(tag *models.Tag) (*models.Tag, error) {
	exists, err := TagExists(tag.ProjectID, tag.RepositoryID, tag.Version)
	if err != nil {
		log.Println(err)
		return nil, err
	}
	if exists {
		log.Println("exists")
		err := UpdateTag(tag)
		if err != nil {
			return nil, err
		}
	} else {
		log.Println("not exist ")
		err := AddTag(tag)
		if err != nil {
			log.Println(err)
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
func TagExists(projectID int64, repositoryID int64, version string) (bool, error) {
	tags, err := GetTagByVersion(projectID, repositoryID, version)
	if err != nil {
		return false, err
	}
	if tags == nil {
		return false, nil
	}
	return true, nil
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

// getTags
func TagsUnderNamespaceAndRepo(namespaceAndRepo string) ([]models.Tag, error) {
	repository, err := GetRepositoryByName(namespaceAndRepo)
	if err != nil {
		return nil, err
	}
	o := orm.NewOrm()
	sql := `select * from tag where project_id=? and repository_id = ? order by id desc`

	var r []models.Tag
	_, err = o.Raw(sql, repository.ProjectID, repository.Id).QueryRows(&r)
	if err != nil {
		return nil, err
	}
	return r, nil
}

// GetTagByVersion...
func GetTagByVersion(projectID int64, repositoryID int64, version string) (*models.Tag, error) {
	o := orm.NewOrm()
	var tags []models.Tag
	count, err := o.Raw("SELECT * from tag where project_id = ? and repository_id = ? and version=? ", projectID, repositoryID, version).QueryRows(&tags)
	log.Println("tagcount: ", count)
	if err != nil {
		return nil, err
	} else if count == 0 {
		return nil, nil
	} else {
		return &tags[0], nil
	}
}
