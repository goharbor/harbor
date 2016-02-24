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
	"errors"
	"fmt"
	"log"

	"github.com/vmware/harbor/dao"
	"github.com/vmware/harbor/models"
	_ "github.com/vmware/harbor/opt_auth/db"
	_ "github.com/vmware/harbor/opt_auth/ldap"
	_ "github.com/vmware/harbor/routers"

	"os"

	"github.com/astaxie/beego"
)

const (
	ADMIN_USER_ID = 1
)

func updateInitPassword(userId int, password string) error {
	queryUser := models.User{UserId: userId}
	user, err := dao.GetUser(queryUser)
	if err != nil {
		log.Println("Failed to get user's initial password, userId:", userId)
		return err
	}
	if user == nil {
		log.Printf("User id: %d does not exist.", userId)
		return errors.New(fmt.Sprintf("User id: %s does not exist.", userId))
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
			log.Printf("Failed to update user encrypted password, userId: %d, err: %v", userId, err)
			return err
		}
		log.Printf("User id: %d updated its encypted password successfully.", userId)
	} else {
		log.Printf("User id: %d already has its encrypted password.", userId)
	}
	return nil
}

func main() {

	beego.BConfig.WebConfig.Session.SessionOn = true

	updateInitPassword(ADMIN_USER_ID, os.Getenv("HARBOR_ADMIN_PASSWORD"))

	beego.Run()
}
