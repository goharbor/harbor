package dao

import (
	"github.com/vmware/harbor/models"
)

func AddLabel(repoLabel models.RepoLabel) (int64, error) {
	o := GetOrmer()

	sql := `insert into repolabel(repoName, label) values(?,?)`


	p,_ := o.Raw(sql).Prepare()

	defer p.Close()

	r, err := p.Exec(repoLabel.RepoName, repoLabel.Label)

	return r.LastInsertId(), err
}


func DeletelLabel(repoLabel models.RepoLabel) (int64, error) {
	o := GetOrmer()

	sql := `delete from repolabel where repoName=? and label=?`


	p,_ := o.Raw(sql).Prepare()

	defer p.Close()

	r, err := p.Exec(repoLabel.RepoName, repoLabel.Label)

	return r.RowsAffected(), err
}


func GetRepoLabels(repoName string) ([]string, error){
	o := GetOrmer()

	sql := `select lable from repolabel where repoName=?`

	var labels []string

	if _, err := o.Raw(sql, repoName).QueryRows(&labels); err != nil {
		return nil, err
	}

	return labels, nil
}


func GetRepoNames(label string) ([]string, error){
	o := GetOrmer()

	sql := `select repoName from repolabel where lable=?`

	var repoNames []string

	if _, err := o.Raw(sql, label).QueryRows(&repoNames); err != nil {
		return nil, err
	}

	return repoNames, nil
}