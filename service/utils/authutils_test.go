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

package utils

import (
	"fmt"
	"log"
	"os"
	"testing"

	_ "github.com/astaxie/beego/orm"
	"github.com/vmware/harbor/dao"
	_ "github.com/vmware/harbor/models"
)

const username string = "Tester01"
const service string = "TesterService"
const scope string = ""

func TestMain(m *testing.M) {

	dbHost := os.Getenv("DB_HOST")
	if len(dbHost) == 0 {
		log.Fatalf("environment variable DB_HOST is not set")
	}
	dbUser := os.Getenv("DB_USR")
	if len(dbUser) == 0 {
		log.Fatalf("environment variable DB_USR is not set")
	}
	dbPort := os.Getenv("DB_PORT")
	if len(dbPort) == 0 {
		log.Fatalf("environment variable DB_PORT is not set")
	}
	dbPassword := os.Getenv("DB_PWD")

	fmt.Printf("DB_HOST: %s, DB_USR: %s, DB_PORT: %s, DB_PWD: %s\n", dbHost, dbUser, dbPort, dbPassword)

	os.Setenv("MYSQL_PORT_3306_TCP_ADDR", dbHost)
	os.Setenv("MYSQL_PORT_3306_TCP_PORT", dbPort)
	os.Setenv("MYSQL_USR", dbUser)
	os.Setenv("MYSQL_PWD", dbPassword)
	os.Setenv("AUTH_MODE", "db_auth")
	dao.InitDB()
	os.Exit(m.Run())

}

func TestGenTokenForUI(t *testing.T) {

	token, err := GenTokenForUI(username, service, scope)
	if err != nil {
		t.Errorf("Error occurred in GenTokenForUI: %v", err)
	}
	log.Printf("Token: %v", token)
}
