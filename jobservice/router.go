package main

import (
	api "github.com/vmware/harbor/api/jobs"

	"github.com/astaxie/beego"
)

func initRouters() {
	beego.Router("/api/replicationJobs", &api.ReplicationJob{})
	beego.Router("/api/replicationJobs/:id/log", &api.ReplicationJob{}, "get:GetLog")
	beego.Router("/api/replicationJobs/actions", &api.ReplicationJob{}, "post:HandleAction")
}
