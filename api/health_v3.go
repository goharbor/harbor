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

package api

import (
	"fmt"

	"github.com/Dataman-Cloud/health_checker"
	"github.com/vmware/harbor/dao"
	"github.com/vmware/harbor/utils"
)

type HealthV3API struct {
	BaseAPI
}

func (ra *HealthV3API) Prepare() {
}

func (ra *HealthV3API) Get() {
	checker := health_checker.NewHealthChecker("Harbor")

	redisHost, redisPort := utils.RedisConfig()
	redisDsn := fmt.Sprintf("%s:%d", redisHost, redisPort)
	checker.AddCheckPoint("redis", redisDsn, nil, nil)

	addr, port, username, password := dao.DbConfig()
	mysqlDsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/registry",
		username, password, addr, port)
	checker.AddCheckPoint("mysql", mysqlDsn, nil, nil)

	ra.Data["json"] = checker.Check()
	ra.ServeJSON()
}
