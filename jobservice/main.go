package main

import (
	"github.com/astaxie/beego"
	"github.com/vmware/harbor/dao"
	"github.com/vmware/harbor/job"
)

func main() {
	dao.InitDB()
	initRouters()
	job.InitWorkerPool()
	go job.Dispatch()
	beego.Run()
}
