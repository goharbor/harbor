// Copyright (c) 2017 VMware, Inc. All Rights Reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//    http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package main

import (
	"os"

	"github.com/astaxie/beego"
	"github.com/vmware/harbor/src/common/dao"
	"github.com/vmware/harbor/src/common/models"
	"github.com/vmware/harbor/src/common/utils/log"
	"github.com/vmware/harbor/src/jobservice/config"
	"github.com/vmware/harbor/src/jobservice/job"
)

func main() {
	log.Info("initializing configurations...")
	if err := config.Init(); err != nil {
		log.Fatalf("failed to initialize configurations: %v", err)
	}
	log.Info("configurations initialization completed")

	database, err := config.Database()
	if err != nil {
		log.Fatalf("failed to get database configurations: %v", err)
	}

	if err := dao.InitDatabase(database); err != nil {
		log.Fatalf("failed to initialize database: %v", err)
	}

	initRouters()
	if err := job.InitWorkerPools(); err != nil {
		log.Fatalf("Failed to initialize worker pools, error: %v", err)
	}
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
	rjobs, err := dao.GetRepJobByStatus(models.JobPending, models.JobRetrying, models.JobRunning)
	if err == nil {
		for _, j := range rjobs {
			rj := job.NewRepJob(j.ID)
			log.Debugf("Resuming replication job: %v", rj)
			job.Schedule(rj)
		}
	} else {
		log.Warningf("Failed to resume replication jobs, error: %v", err)
	}
	sjobs, err := dao.GetScanJobsByStatus(models.JobPending, models.JobRetrying, models.JobRunning)
	if err == nil {
		for _, j := range sjobs {
			sj := job.NewScanJob(j.ID)
			log.Debugf("Resuming scan job: %v", sj)
			job.Schedule(sj)
		}
	} else {
		log.Warningf("Failed to resume scan jobs, error: %v", err)
	}
}

func init() {
	configPath := os.Getenv("CONFIG_PATH")
	if len(configPath) != 0 {
		log.Infof("Config path: %s", configPath)
		if err := beego.LoadAppConfig("ini", configPath); err != nil {
			log.Fatalf("Failed to load config file: %s, error: %v", configPath, err)
		}
	}
}
