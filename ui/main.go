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
	"fmt"
	"os"

	log "github.com/vmware/harbor/utils/log"

	"github.com/vmware/harbor/api"
	_ "github.com/vmware/harbor/auth/db"
	_ "github.com/vmware/harbor/auth/ldap"
	"github.com/vmware/harbor/dao"
	"github.com/vmware/harbor/models"

	"github.com/astaxie/beego"
	_ "github.com/astaxie/beego/session/redis"
)

const (
	adminUserID = 1
)

func updateInitPassword(userID int, password string) error {
	queryUser := models.User{UserID: userID}
	user, err := dao.GetUser(queryUser)
	if err != nil {
		return fmt.Errorf("Failed to get user, userID: %d %v", userID, err)
	}
	if user == nil {
		return fmt.Errorf("User id: %d does not exist.", userID)
	}
	if user.Salt == "" {
		salt, err := dao.GenerateRandomString()
		if err != nil {
			return fmt.Errorf("Failed to generate salt for encrypting password, %v", err)
		}

		user.Salt = salt
		user.Password = password
		err = dao.ChangeUserPassword(*user)
		if err != nil {
			return fmt.Errorf("Failed to update user encrypted password, userID: %d, err: %v", userID, err)
		}

		log.Infof("User id: %d updated its encypted password successfully.", userID)
	} else {
		log.Infof("User id: %d already has its encrypted password.", userID)
	}
	return nil
}

func main() {

	beego.BConfig.WebConfig.Session.SessionOn = true
	//TODO
	redisURL := os.Getenv("_REDIS_URL")
	if len(redisURL) > 0 {
		beego.BConfig.WebConfig.Session.SessionProvider = "redis"
		beego.BConfig.WebConfig.Session.SessionProviderConfig = redisURL
	}
	//
	beego.AddTemplateExt("htm")
	dao.InitDB()
	if err := updateInitPassword(adminUserID, os.Getenv("HARBOR_ADMIN_PASSWORD")); err != nil {
		log.Error(err)
	}
	initRouters()
	if err := api.SyncRegistry(); err != nil {
		log.Error(err)
	}
	beego.Run()
}
