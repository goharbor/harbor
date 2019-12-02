// Copyright Project Harbor Authors
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
	"github.com/goharbor/harbor/src/common/models"
)

var orderMap = map[string]string{
	"name":           "name asc",
	"+name":          "name asc",
	"-name":          "name desc",
	"creation_time":  "creation_time asc",
	"+creation_time": "creation_time asc",
	"-creation_time": "creation_time desc",
	"update_time":    "update_time asc",
	"+update_time":   "update_time asc",
	"-update_time":   "update_time desc",
	"pull_count":     "pull_count asc",
	"+pull_count":    "pull_count asc",
	"-pull_count":    "pull_count desc",
}

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
	if err != nil {
		return err
	}
	if num == 0 {
		return fmt.Errorf("Failed to increase repository pull count with name: %s", name)
	}
	return nil
}

// RepositoryExists returns whether the repository exists according to its name.
func RepositoryExists(name string) bool {
	o := GetOrmer()
	return o.QueryTable("repository").Filter("name", name).Exist()
}

// GetTopRepos returns the most popular repositories whose project ID is
// in projectIDs
func GetTopRepos(projectIDs []int64, n int) ([]*models.RepoRecord, error) {
	repositories := []*models.RepoRecord{}
	if len(projectIDs) == 0 {
		return repositories, nil
	}

	_, err := GetOrmer().QueryTable(&models.RepoRecord{}).
		Filter("project_id__in", projectIDs).
		OrderBy("-pull_count").
		Limit(n).
		All(&repositories)

	return repositories, err
}

// GetTotalOfRepositories ...
func GetTotalOfRepositories(query ...*models.RepositoryQuery) (int64, error) {
	sql, params := repositoryQueryConditions(query...)
	sql = `select count(*) ` + sql
	var total int64
	if err := GetOrmer().Raw(sql, params).QueryRow(&total); err != nil {
		return 0, err
	}
	return total, nil
}

// GetRepositories ...
func GetRepositories(query ...*models.RepositoryQuery) ([]*models.RepoRecord, error) {
	repositories := []*models.RepoRecord{}
	order := "name asc"
	if len(query) > 0 && query[0] != nil {
		if s, ok := orderMap[query[0].Sort]; ok {
			order = s
		}
	}

	condition, params := repositoryQueryConditions(query...)
	sql := fmt.Sprintf(`select r.repository_id, r.name, r.project_id, r.description, r.pull_count, 
	r.star_count, r.creation_time, r.update_time %s order by r.%s `, condition, order)
	if len(query) > 0 && query[0] != nil {
		page, size := query[0].Page, query[0].Size
		if size > 0 {
			sql += `limit ? `
			params = append(params, size)
			if page > 0 {
				sql += `offset ? `
				params = append(params, size*(page-1))
			}
		}
	}

	if _, err := GetOrmer().Raw(sql, params).QueryRows(&repositories); err != nil {
		return nil, err
	}

	return repositories, nil
}

func repositoryQueryConditions(query ...*models.RepositoryQuery) (string, []interface{}) {
	params := []interface{}{}
	sql := `from repository r `
	if len(query) == 0 || query[0] == nil {
		return sql, params
	}
	q := query[0]

	if q.LabelID > 0 {
		sql += `join harbor_resource_label rl on r.repository_id = rl.resource_id 
		and rl.resource_type = 'r' `
	}
	sql += `where 1=1 `

	if len(q.Name) > 0 {
		sql += `and r.name like ? `
		params = append(params, "%"+Escape(q.Name)+"%")
	}

	if len(q.ProjectIDs) > 0 {
		sql += fmt.Sprintf(`and r.project_id in ( %s ) `,
			ParamPlaceholderForIn(len(q.ProjectIDs)))
		params = append(params, q.ProjectIDs)
	}

	if len(q.ProjectName) > 0 {
		// use "like" rather than "table joining" because that
		// in integration mode the projects are saved in Admiral side
		sql += `and r.name like ? `
		params = append(params, q.ProjectName+"/%")
	}

	if q.LabelID > 0 {
		sql += `and rl.label_id = ? `
		params = append(params, q.LabelID)
	}

	return sql, params
}
