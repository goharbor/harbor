package dao

import (
	"github.com/vmware/harbor/models"
	"fmt"
)

func UpdateProjectById(c models.ProjectDesc) (err error) {
	o := GetOrmer()
	res, err := o.Raw("UPDATE project_desc SET name = ? WHERE project_id = ?",c.Name,c.ProjectId).Exec()
	if err == nil {
		num, _ := res.RowsAffected()
		fmt.Println("mysql row affected nums: ", num)
	}
	return err
}
