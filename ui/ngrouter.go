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
	beego.Router("/ng/admin_option", &ng.AdminOptionController{})
	beego.Router("/ng/forgot_password", &ng.ForgotPasswordController{})
	beego.Router("/ng/reset_password", &ng.ResetPasswordController{})
	beego.Router("/ng/search", &ng.SearchController{})

	beego.Router("/ng/log_out", &ng.CommonController{}, "get:LogOut")
	beego.Router("/ng/reset", &ng.CommonController{}, "post:ResetPassword")
	beego.Router("/ng/sendEmail", &ng.CommonController{}, "get:SendEmail")
	beego.Router("/ng/language", &ng.CommonController{}, "get:SwitchLanguage")

	beego.Router("/ng/optional_menu", &ng.OptionalMenuController{})
	beego.Router("/ng/navigation_header", &ng.NavigationHeaderController{})
}
