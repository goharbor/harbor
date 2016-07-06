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
	"log"
	"os"

	"github.com/astaxie/beego"
	_ "github.com/astaxie/beego/session/redis"
	_ "github.com/vmware/harbor/auth/db"
	"github.com/vmware/harbor/dao"
	"github.com/vmware/harbor/git"
	"github.com/vmware/harbor/models"
	_ "github.com/vmware/harbor/routers"
)

const (
	adminUserID = 1
)

func updateInitPassword(userID int, password string) error {
	queryUser := models.User{UserID: userID}
	user, err := dao.GetUser(queryUser)
	if err != nil {
		log.Println("Failed to get user, userID:", userID)
		return err
	}
	if user == nil {
		log.Printf("User id: %d does not exist.", userID)
		return fmt.Errorf("User id: %d does not exist.", userID)
	} else if user.Salt == "" {
		salt, err := dao.GenerateRandomString()
		if err != nil {
			log.Printf("Failed to generate salt for encrypting password, %v", err)
			return err
		}
		user.Salt = salt
		user.Password = password
		err = dao.ChangeUserPassword(*user)
		if err != nil {
			log.Printf("Failed to update user encrypted password, userID: %d, err: %v", userID, err)
			return err
		}
		log.Printf("User id: %d updated its encypted password successfully.", userID)
	} else {
		log.Printf("User id: %d already has its encrypted password.", userID)
	}
	return nil
}
func modCatalogTimer() {
	log.Println("start init git client and clone the uri")
	client := InitClient()
	log.Println("start git pull timer")
	go PullTimer(client)
}
func main() {
	beego.BConfig.WebConfig.Session.SessionOn = true
	beego.BConfig.WebConfig.Session.SessionProvider = "redis"
	beego.BConfig.WebConfig.Session.SessionProviderConfig = fmt.Sprintf("%s:%s,20", os.Getenv("REDIS_HOST"), os.Getenv("REDIS_PORT"))

	dao.InitDB()
	dao.UpgradeDB()
	updateInitPassword(adminUserID, os.Getenv("HARBOR_ADMIN_PASSWORD"))

	git.InitParam()
	config := git.Pairs()
	if config.Storemethod == "git" {
		modCatalogTimer()
	}

	beego.Run()
}
