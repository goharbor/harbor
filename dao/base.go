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
	"log"
	"net"
	"os"
	"strings"
	"time"

	"github.com/astaxie/beego/orm"
	_ "github.com/go-sql-driver/mysql" //register mysql driver
	_ "github.com/mattes/migrate/driver/mysql"
	"github.com/mattes/migrate/migrate"
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
	addr, port, username, password := dbConfig()
	dbStr := username + ":" + password + "@tcp(" + addr + ":" + port + ")/registry"
	log.Println(dbStr)
	ch := make(chan int, 1)
	go func() {
		var err error
		var c net.Conn
		for {
			c, err = net.Dial("tcp", addr+":"+port)
			if err == nil {
				c.Close()
				ch <- 1
			} else {
				log.Printf("failed to connect to db, retry after 2 seconds...")
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

func UpgradeDB() {
	orm.RegisterDriver("mysql", orm.DRMySQL)
	addr, port, username, password := dbConfig()

	dbStr := fmt.Sprintf("mysql://%s:%s@tcp(%s:%s)/registry",
		username, password, addr, port)
	log.Printf("upgrading DB", dbStr)

	sqlPath := os.Getenv("SQL_PATH")
	fmt.Println(sqlPath)
	if len(sqlPath) == 0 {
		sqlPath = "/go/src/github.com/vmware/harbor/sql"
	}
	errors, ok := migrate.UpSync(dbStr, sqlPath)
	if errors != nil && len(errors) > 0 {
		for _, err := range errors {
			log.Printf("db err", err)
		}
		log.Printf("can't upgrade db", errors)
		panic(-1)
	}
	if !ok {
		log.Printf("can't upgrade db")
		panic(-1)
	}
	log.Printf("DB upgraded")
}

func dbConfig() (string, string, string, string) {
	addr := os.Getenv("MYSQL_HOST")
	if len(addr) == 0 {
		addr = os.Getenv("MYSQL_PORT_3306_TCP_ADDR")
	}

	port := os.Getenv("MYSQL_PORT_3306_TCP_PORT")

	if len(port) == 0 {
		port = os.Getenv("MYSQL_PORT")
	}

	username := os.Getenv("MYSQL_USR")

	password := os.Getenv("MYSQL_ENV_MYSQL_ROOT_PASSWORD")
	if len(password) == 0 {
		password = os.Getenv("MYSQL_PWD")
	}

	return addr, port, username, password
}
