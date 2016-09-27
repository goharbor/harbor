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
	"fmt"
	"net"

	"os"
	"sync"
	"time"

	"github.com/astaxie/beego/orm"
	_ "github.com/go-sql-driver/mysql" //register mysql driver
	"github.com/vmware/harbor/utils/log"
)

// NonExistUserID : if a user does not exist, the ID of the user will be 0.
const NonExistUserID = 0

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
	// orm.Debug = true
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

var globalOrm orm.Ormer
var once sync.Once

// GetOrmer :set ormer singleton
func GetOrmer() orm.Ormer {
	once.Do(func() {
		globalOrm = orm.NewOrm()
	})
	return globalOrm
}

func paginateForRawSQL(sql string, limit, offset int64) string {
	return fmt.Sprintf("%s limit %d offset %d", sql, limit, offset)
}
