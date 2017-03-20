package controllers

import "github.com/astaxie/beego"

// IndexController handles request to /
type IndexController struct {
	beego.Controller
}

// Get renders the index page
func (ic *IndexController) Get() {
	ic.TplExt = "html"
	ic.TplName = "index.html"
}
