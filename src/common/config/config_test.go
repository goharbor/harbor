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
package config

import (
	"os"
	"testing"
)

func TestEnvConfLoader(t *testing.T) {
	os.Unsetenv("KEY2")
	os.Setenv("KEY1", "V1")
	os.Setenv("KEY3", "V3")
	keys := []string{"KEY1", "KEY2"}
	ecl := EnvConfigLoader{
		keys,
	}
	m, err := ecl.Load()
	if err != nil {
		t.Errorf("Error loading the configuration via env: %v", err)
	}
	if m["KEY1"] != "V1" {
		t.Errorf("The value for key KEY1 should be V1, but infact: %s", m["KEY1"])
	}
	if len(m["KEY2"]) > 0 {
		t.Errorf("The value for key KEY2 should be emptye, but infact: %s", m["KEY2"])
	}
	if _, ok := m["KEY3"]; ok {
		t.Errorf("The KEY3 should not be in result as it's not in the initial key list")
	}
	os.Unsetenv("KEY1")
	os.Unsetenv("KEY3")
}

func TestCommonConfig(t *testing.T) {

	mysql := MySQLSetting{"registry", "root", "password", "127.0.0.1", "3306"}
	sqlite := SQLiteSetting{"file.db"}
	verify := "off"
	ext := "http://harbor"
	token := "http://token"
	loglevel := "info"

	os.Setenv("DATABASE", "")
	os.Setenv("MYSQL_DATABASE", mysql.Database)
	os.Setenv("MYSQL_USR", mysql.User)
	os.Setenv("MYSQL_PWD", mysql.Password)
	os.Setenv("MYSQL_HOST", mysql.Host)
	os.Setenv("MYSQL_PORT", mysql.Port)
	os.Setenv("SQLITE_FILE", sqlite.FilePath)
	os.Setenv("VERIFY_REMOTE_CERT", verify)
	os.Setenv("EXT_ENDPOINT", ext)
	os.Setenv("TOKEN_ENDPOINT", token)
	os.Setenv("LOG_LEVEL", loglevel)

	err := Reload()
	if err != nil {
		t.Errorf("Unexpected error when loading the configurations, error: %v", err)
	}
	if Database() != "mysql" {
		t.Errorf("Expected Database value: mysql, fact: %s", mysql)
	}
	if MySQL() != mysql {
		t.Errorf("Expected MySQL setting: %+v, fact: %+v", mysql, MySQL())
	}
	if VerifyRemoteCert() {
		t.Errorf("Expected VerifyRemoteCert: false, env var: %s, fact: %v", verify, VerifyRemoteCert())
	}
	if ExtEndpoint() != ext {
		t.Errorf("Expected ExtEndpoint: %s, fact: %s", ext, ExtEndpoint())
	}
	if TokenEndpoint() != token {
		t.Errorf("Expected TokenEndpoint: %s, fact: %s", token, TokenEndpoint())
	}
	if LogLevel() != loglevel {
		t.Errorf("Expected LogLevel: %s, fact: %s", loglevel, LogLevel())
	}
	os.Setenv("DATABASE", "sqlite")
	err = Reload()
	if err != nil {
		t.Errorf("Unexpected error when loading the configurations, error: %v", err)
	}
	if SQLite() != sqlite {
		t.Errorf("Expected SQLite setting: %+v, fact %+v", sqlite, SQLite())
	}

	os.Unsetenv("DATABASE")
	os.Unsetenv("MYSQL_DATABASE")
	os.Unsetenv("MYSQL_USR")
	os.Unsetenv("MYSQL_PWD")
	os.Unsetenv("MYSQL_HOST")
	os.Unsetenv("MYSQL_PORT")
	os.Unsetenv("SQLITE_FILE")
	os.Unsetenv("VERIFY_REMOTE_CERT")
	os.Unsetenv("EXT_ENDPOINT")
	os.Unsetenv("TOKEN_ENDPOINT")
	os.Unsetenv("LOG_LEVEL")

}
