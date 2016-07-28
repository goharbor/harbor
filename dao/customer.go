package dao

import (
	"github.com/vmware/harbor/models"
	"fmt"
	"github.com/vmware/harbor/utils/log"
)


func AddCustomer(m models.Customer) (res bool , err error) {
	o := GetOrmer()

	sql := `select * from customer where name= ?`
	type dummy struct{}
	var d []dummy
	_, err = o.Raw(sql, m.Name).QueryRows(&d)
	if len(d) != 0 {
		return false, err
	}

	sql = `insert into customer(name,tag) values(?, ?)`
	p,_ := o.Raw(sql).Prepare()
	defer p.Close()

	_,err = p.Exec(m.Name,m.Tag )

	return true,err
}

func GetCustomerById(id int) (*models.Customer, error) {
	o := GetOrmer()

	p := models.Customer{}
	err := o.Raw("select * from customer where id = ?", id).QueryRow(&p)

	if err != nil {
		return nil, err
	}

	log.Infof("user: %+v", p)

	return &p, nil
}

func GetProjectAllCustomer(projectId string) ([]models.Customer, error) {
	o := GetOrmer()

	var customer []models.Customer
	if projectId == "" {
		//返回客户列表
		if _, err := o.Raw("select * from customer").QueryRows(&customer); err != nil {
			return nil, err
		}
	}else{
		p := models.Project{}
		sql := `select * from project where project_id = ?`
		if err := o.Raw(sql,projectId).QueryRow(&p); err != nil {
			return nil, err
		}
		projectName := p.Name+"%"

		log.Infof("res: %+v", projectName)
		//返回项目的客户列表
		// select * from customer where tag in (select label from repo_label where repoName like 'library%' group by label)
		sql = `select * from customer where name in (select label from repo_label
		 where repoName like ? group by label)`

		if _, err := o.Raw(sql,projectName).QueryRows(&customer); err != nil {
			return nil, err
		}
	}
	return customer, nil
}


func UpdateCustomerById(c models.Customer) (err error) {
	o := GetOrmer()
	res, err := o.Raw("UPDATE customer SET name = ?,tag =? WHERE id = ?",c.Name,c.Tag,c.Id).Exec()
	if err == nil {
		num, _ := res.RowsAffected()
		fmt.Println("mysql row affected nums: ", num)
	}
	return err
}

func DeleteCustomer(id int) (err error) {
	o := GetOrmer()
	res, err := o.Raw("delete from customer where id = ?",id).Exec()
	if err == nil {
		num, _ := res.RowsAffected()
		fmt.Println("mysql row affected nums: ", num)
	}
	return err
}

//获取客户数量
func GetCustomerCount() (count int) {
	o := GetOrmer()
	p := []models.Customer{}
	o.Raw("select * from customer").QueryRows(&p)
	return len(p)
}

//增加客户项目过滤
func GetCustomerRepoList(repoList []string, tag string) ([]string) {
	o := GetOrmer()

	var resp []string

	sql := "select * from repo_label where repoName=? and label=?"

	p := models.RepoLabel{}

	for _, r := range repoList {

		if err := o.Raw(sql,r,tag).QueryRow(&p); err != nil {
			continue
		}
		resp = append(resp, r)
	}
	return resp
}
