package dao

import (
	"github.com/vmware/harbor/models"
)

func AddLabel(repoLabel models.RepoLabel) (int64, error) {
	o := GetOrmer()

	sql := `insert into repo_label(repoName, label) values(?,?)`


	p,_ := o.Raw(sql).Prepare()

	defer p.Close()

	r, err := p.Exec(repoLabel.RepoName, repoLabel.Label)

	insertId,_ := r.LastInsertId()
	return insertId, err
}


func DeletelLabel(repoLabel models.RepoLabel) (int64, error) {
	o := GetOrmer()

	sql := `delete from repo_label where repoName=? and label=?`


	p,_ := o.Raw(sql).Prepare()

	defer p.Close()

	r, err := p.Exec(repoLabel.RepoName, repoLabel.Label)

	affectedRows, _ := r.RowsAffected()

	return affectedRows, err
}


func GetRepoLabels(repoName string) ([]string, error){
	o := GetOrmer()

	sql := `select lable from repo_label where repoName=?`

	var labels []string

	if _, err := o.Raw(sql, repoName).QueryRows(&labels); err != nil {
		return nil, err
	}

	return labels, nil
}


func GetRepoNames(label string) ([]string, error){
	o := GetOrmer()

	sql := `select repoName from repo_label where lable=?`

	var repoNames []string

	if _, err := o.Raw(sql, label).QueryRows(&repoNames); err != nil {
		return nil, err
	}

	return repoNames, nil
}