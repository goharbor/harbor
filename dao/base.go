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

package dao

import (
	"net"

	"github.com/vmware/harbor/utils/log"

	"os"
	"strings"
	"time"

	"github.com/astaxie/beego/orm"
	_ "github.com/go-sql-driver/mysql" //register mysql driver
)

// NonExistUserID : if a user does not exist, the ID of the user will be 0.
const NonExistUserID = 0

func isIllegalLength(s string, min int, max int) bool {
	if min == -1 {
		return (len(s) > max)
	}
	if max == -1 {
		return (len(s) <= min)
	}
	return (len(s) < min || len(s) > max)
}

func isContainIllegalChar(s string, illegalChar []string) bool {
	for _, c := range illegalChar {
		if strings.Index(s, c) >= 0 {
			return true
		}
	}
	return false
}

// GenerateRandomString generates a random string
func GenerateRandomString() (string, error) {
	o := orm.NewOrm()
	var uuid string
	err := o.Raw(`select uuid() as uuid`).QueryRow(&uuid)
	if err != nil {
		return "", err
	}
	return uuid, nil

}

//InitDB initializes the database
func InitDB() {
	orm.RegisterDriver("mysql", orm.DRMySQL)
	addr := os.Getenv("MYSQL_HOST")
	port := os.Getenv("MYSQL_PORT")
	username := os.Getenv("MYSQL_USR")
	password := os.Getenv("MYSQL_PWD")

	log.Debugf("db url: %s:%s, db user: %s", addr, port, username)
	dbStr := username + ":" + password + "@tcp(" + addr + ":" + port + ")/registry"
	ch := make(chan int, 1)
	go func() {
		var err error
		var c net.Conn
		for {
			c, err = net.DialTimeout("tcp", addr+":"+port, 20*time.Second)
			if err == nil {
				c.Close()
				ch <- 1
			} else {
				log.Errorf("failed to connect to db, retry after 2 seconds :%v", err)
				time.Sleep(2 * time.Second)
			}
		}
	}()
	select {
	case <-ch:
	case <-time.After(60 * time.Second):
		panic("Failed to connect to DB after 60 seconds")
	}
	err := orm.RegisterDataBase("default", "mysql", dbStr)
	if err != nil {
		panic(err)
	}
}
