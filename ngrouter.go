package main

import (
	"github.com/astaxie/beego"
	"github.com/vmware/harbor/controllers/ng"
)

func initNgRouters() {

	beego.SetStaticPath("ng/static", "ng")
	beego.Router("/ng", &ng.IndexController{})
	beego.Router("/ng/dashboard", &ng.DashboardController{})
	beego.Router("/ng/project", &ng.ProjectController{})
	beego.Router("/ng/repository", &ng.RepositoryController{})
	beego.Router("/ng/sign_up", &ng.SignUpController{})
	beego.Router("/ng/account_setting", &ng.AccountSettingController{})
}
