/*
   Copyright (c) 2016 VMware, Inc. All Rights Reserved.
   Licensed under the Apache License, Version 2.0 (the "License");
   you may not use this file except in compliance with the License.
   You may obtain a copy of the License at

       http://www.apache.org/licenses/LICENSE-2.0

   Unless required by applicable law or agreed to in writing, software
   distributed under the License is distributed on an "AS IS" BASIS,
   WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
   See the License for the specific language governing permissions and
   limitations under the License.
*/

package main

import (
	"github.com/astaxie/beego"
	"github.com/vmware/harbor/dao"
	"github.com/vmware/harbor/job"
	"github.com/vmware/harbor/models"
	"github.com/vmware/harbor/utils/log"
)

func main() {
	dao.InitDB()
	initRouters()
	job.InitWorkerPool()
	go job.Dispatch()
	resumeJobs()
	beego.Run()
}

func resumeJobs() {
	log.Debugf("Trying to resume halted jobs...")
	err := dao.ResetRunningJobs()
	if err != nil {
		log.Warningf("Failed to reset all running jobs to pending, error: %v", err)
	}
	jobs, err := dao.GetRepJobByStatus(models.JobPending, models.JobRetrying)
	if err == nil {
		for _, j := range jobs {
			log.Debugf("Resuming job: %d", j.ID)
			job.Schedule(j.ID)
		}
	} else {
		log.Warningf("Failed to jobs to resume, error: %v", err)
	}
}
