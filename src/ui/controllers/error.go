package controllers

import (
	"github.com/astaxie/beego"
)

type ErrorController struct {
	beego.Controller
}

func (ec *ErrorController) Error404() {
	ec.Data["content"] = "page not found"
	ec.TplName = "404.tpl"
}
