package main

import (
	"github.com/vmware/harbor/api"

	"github.com/astaxie/beego"
)

func initRouters() {
	beego.Router("/api/jobs/replication", &api.ReplicationJob{})
}
