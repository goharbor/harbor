package main

import (
	"github.com/astaxie/beego"
	"github.com/vmware/harbor/dao"
	_ "github.com/vmware/harbor/job/imgout"
	//	"github.com/vmware/harbor/utils/log"
)

func main() {
	dao.InitDB()
	initRouters()
	beego.Run()
}
