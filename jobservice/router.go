package main

import (
	api "github.com/vmware/harbor/api/jobs"

	"github.com/astaxie/beego"
)

func initRouters() {
	beego.Router("/api/jobs/replication", &api.ReplicationJob{})
}
