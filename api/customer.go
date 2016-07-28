package api

import (
	"github.com/vmware/harbor/models"
	"strconv"
	"net/http"
	"github.com/vmware/harbor/dao"
	"github.com/vmware/harbor/utils/log"
	"fmt"
	"regexp"
)

type CustomerController struct {
	BaseAPI
}

type CustomerReq struct {
	Id int 					`json:"id"`
	Name string 		`json:"name"`
	Tag  string   	`json:"tag"`
}


const customerNameMaxLen int = 30
const customerNameMinLen int = 4
/*
	Method: Post
 	https://registry.51yixiao.com/api/customer
	param: name 客户中文名
	param: tag 客户标签
 */

func (c *CustomerController) PostCustomer() {

	var req CustomerReq
	c.DecodeJSONReq(&req)

	name := req.Name
	if len(name) == 0 {
		c.CustomAbort(http.StatusBadRequest, "name is nil")
	}

	err := validateCustomerReq(name)
	if err != nil {
		log.Errorf("Invalid customer request, error: %v", err)
		c.RenderError(http.StatusBadRequest, fmt.Sprintf("invalid request: %v", err))
		return
	}

	tag := req.Tag
	if len(tag) == 0 {
		c.CustomAbort(http.StatusBadRequest, "tag is nil")
	}

	customer := models.Customer{Name: name, Tag: tag}
	if res,err := dao.AddCustomer(customer); err == nil {
		//StatusCreated
		if res == true {
			c.CustomAbort(http.StatusCreated, "add success")
		}else{
			c.CustomAbort(http.StatusConflict, "customer is exist ")
		}
		c.Data["json"] = res
	}else{
		c.CustomAbort(http.StatusInternalServerError,"Failed to insert customer to db")
	}

	c.ServeJSON()
}

/*
	Method: Get
 	https://registry.51yixiao.com/api/customer/3
	ret: 返回客户信息
 */
func (c *CustomerController) GetOneCustomer() {
	idStr := c.Ctx.Input.Param(":id")

	log.Infof("idStr: %+v",idStr )
	id, _ := strconv.Atoi(idStr)

	v, err := dao.GetCustomerById(id)
	if err != nil {
		c.Data["json"] = err.Error()
	} else {
		c.Data["json"] = v
	}
	c.ServeJSON()
}

/*
	Method: Post
 	https://registry.51yixiao.com/api/customer
	param: project_name 项目名称
	ret: 返回客户列表
 */
func (c *CustomerController) GetListCustomer() {
	projectId := c.GetString("project_id")
	log.Infof("projectID: %+v",projectId )
	l, err := dao.GetProjectAllCustomer(projectId)
	if err != nil {
		c.Data["json"] = err.Error()
	} else {
		c.Data["json"] = l
	}
	log.Infof("user: %+v",l)
	c.ServeJSON()
}

/*
 Method: Put
 https://registry.51yixiao.com/api/customer/3
 param: name 招商银行
 param: tag CMM
 */
func (c *CustomerController) UpdateCustomer() {

	var req CustomerReq
	c.DecodeJSONReq(&req)

	m := models.Customer{}
	m.Id = req.Id
	m.Name = req.Name
	m.Tag = req.Tag

	log.Infof("user: %+v",m)

	if err := dao.UpdateCustomerById(m); err == nil {
			c.Data["json"] = "OK"
	} else {
		c.Data["json"] = err.Error()
	}

	c.ServeJSON()
}

/*
Method: Delete
https://registry.51yixiao.com/api/customer/3
 */
func (c *CustomerController) DeleteCustomer() {
	idStr := c.Ctx.Input.Param(":id")
	id, _ := strconv.Atoi(idStr)
	if err := dao.DeleteCustomer(id); err == nil {
		c.Data["json"] = "OK"
	} else {
		c.Data["json"] = err.Error()
	}
	c.ServeJSON()
}

func validateCustomerReq(customer_name string) error {

	if isIllegalLength(customer_name, customerNameMinLen, customerNameMaxLen) {
		return fmt.Errorf("Customer name is illegal in length. (greater than 4 or less than 30)")
	}
	validName := regexp.MustCompile(`^[a-z0-9](?:-*[a-z0-9])*(?:[._][a-z0-9](?:-*[a-z0-9])*)*$`)
	legal := validName.MatchString(customer_name)
	if !legal {
		return fmt.Errorf("Customer name is not in lower case or contains illegal characters!")
	}
	return nil
}