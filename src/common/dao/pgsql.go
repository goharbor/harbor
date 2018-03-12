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

package dao

import (
	"fmt"

	"github.com/astaxie/beego/orm"
	_ "github.com/lib/pq" //register pgsql driver
	"github.com/vmware/harbor/src/common/utils"
)

type pgsql struct {
	host     string
	port     string
	usr      string
	pwd      string
	database string
	sslmode  bool
}

type pgsqlSSLMode bool

func (pm pgsqlSSLMode) String() string {
	if bool(pm) {
		return "enable"
	}
	return "disable"
}

// Name returns the name of PostgreSQL
func (p *pgsql) Name() string {
	return "PostgreSQL"
}

// String ...
func (p *pgsql) String() string {
	return fmt.Sprintf("type-%s host-%s port-%s databse-%s sslmode-%q",
		p.Name(), p.host, p.port, p.database, pgsqlSSLMode(p.sslmode))
}

// NewPQSQL returns an instance of postgres
func NewPQSQL(host string, port string, usr string, pwd string, database string, sslmode bool) Database {
	return &pgsql{
		host:     host,
		port:     port,
		usr:      usr,
		pwd:      pwd,
		database: database,
		sslmode:  sslmode,
	}
}

//Register registers pgSQL to orm with the info wrapped by the instance.
func (p *pgsql) Register(alias ...string) error {
	if err := utils.TestTCPConn(fmt.Sprintf("%s:%s", p.host, p.port), 60, 2); err != nil {
		return err
	}

	if err := orm.RegisterDriver("postgres", orm.DRPostgres); err != nil {
		return err
	}

	an := "default"
	if len(alias) != 0 {
		an = alias[0]
	}
	info := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
		p.host, p.port, p.usr, p.pwd, p.database, pgsqlSSLMode(p.sslmode))

	return orm.RegisterDataBase(an, "postgres", info)
}
