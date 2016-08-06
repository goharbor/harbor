/*
   Copyright (c) 2016 VMware, Inc. All Rights Reserved.
   Licensed under the Apache License, Version 2.0 (the "License");
   you may not use this file except in compliance with the License.
   You may obtain a copy of the License at

       http://www.apache.org/licenses/LICENSE-2.0

   Unless required by applicable law or agreed to in writing, software
   distributed under the License is distributed on an "AS IS" BASIS,
   WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, e
   ither express or implied.
   See the License for the specific language governing permissions and
   limitations under the License.
*/

package config

import (
	"os"
)

// MysqlConfig ...
type MysqlConfig struct {
	addr     string
	port     string
	username string
	password string
}

var mysqlConfig MysqlConfig
var harborAdminPwd string
var redisUrl string

func init() {

	mysqlConfig.addr = os.Getenv("MYSQL_HOST")
	mysqlConfig.port = os.Getenv("MYSQL_PORT")
	mysqlConfig.username = os.Getenv("MYSQL_USR")
	mysqlConfig.password = os.Getenv("MYSQL_PWD")
	harborAdminPwd = os.Getenv("HARBOR_ADMIN_PASSWORD")
	redisUrl = os.Getenv("_REDIS_URL")
}

// MysqlCfg return mysql configure entity
func MysqlCfg() MysqlConfig {
	return mysqlConfig
}

// MysqlAddr return mysql address
func MysqlAddr() string {
	return mysqlConfig.addr
}

// MysqlPort return mysql port
func MysqlPort() string {
	return mysqlConfig.port
}

// MysqlUserName return msyql user name
func MysqlUserName() string {
	return mysqlConfig.username
}

// MysqlPwd return mysql password
func MysqlPwd() string {
	return mysqlConfig.password
}

// HarborAdminPwd return harbor admin password
func HarborAdminPwd() string {
	return harborAdminPwd
}

// RedisUrl return redis url
func RedisUrl() {
	return redisUrl
}
