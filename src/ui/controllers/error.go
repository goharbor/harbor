package controllers

import (
	"github.com/astaxie/beego"
)

// ErrorController handles beego error pages
type ErrorController struct {
	beego.Controller
}

// Error404 renders the 404 page
func (ec *ErrorController) Error404() {
	ec.Data["content"] = "page not found"
	ec.TplName = "404.tpl"
}
