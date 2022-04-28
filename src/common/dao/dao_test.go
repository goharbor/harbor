// Copyright Project Harbor Authors
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
	"context"
	"os"
	"testing"

	"github.com/beego/beego/orm"
	"github.com/goharbor/harbor/src/common/models"
	"github.com/goharbor/harbor/src/lib/log"
	libOrm "github.com/goharbor/harbor/src/lib/orm"
	"github.com/goharbor/harbor/src/pkg/user"
)

var testCtx context.Context

func execUpdate(o orm.Ormer, sql string, params ...interface{}) error {
	p, err := o.Raw(sql).Prepare()
	if err != nil {
		return err
	}
	defer p.Close()
	_, err = p.Exec(params...)
	if err != nil {
		return err
	}
	return nil
}

func cleanByUser(username string) {
	var err error

	o := GetOrmer()
	o.Begin()

	err = execUpdate(o, `delete
		from project_member
		where entity_id = (
			select user_id
			from harbor_user
			where username = ?
		) `, username)
	if err != nil {
		o.Rollback()
		log.Error(err)
	}

	err = execUpdate(o, `delete
		from project_member
		where project_id = (
			select project_id
			from project
			where name = ?
		)`, projectName)
	if err != nil {
		o.Rollback()
		log.Error(err)
	}

	err = execUpdate(o, `delete from project where name = ?`, projectName)
	if err != nil {
		o.Rollback()
		log.Error(err)
	}

	err = execUpdate(o, `delete from harbor_user where username = ?`, username)
	if err != nil {
		o.Rollback()
		log.Error(err)
	}
	err = execUpdate(o, `delete from replication_policy where id < 99`)
	if err != nil {
		log.Error(err)
	}
	err = execUpdate(o, `delete from registry where id < 99`)
	if err != nil {
		log.Error(err)
	}
	o.Commit()
}

const username string = "Tester01"
const password string = "Abc12345"
const projectName string = "test_project"

func TestMain(m *testing.M) {
	databases := []string{"postgresql"}
	for _, database := range databases {
		log.Infof("run test cases for database: %s", database)
		result := 1
		switch database {
		case "postgresql":
			PrepareTestForPostgresSQL()
		default:
			log.Fatalf("invalid database: %s", database)
		}
		testCtx = libOrm.Context()
		result = testForAll(m)

		if result != 0 {
			os.Exit(result)
		}
	}
}

func testForAll(m *testing.M) int {
	cleanByUser(username)
	// TODO: remove the code for populating data after the record is not needed.
	ctx := libOrm.Context()
	_, err := user.Mgr.Create(ctx, &models.User{
		Username: username,
		Email:    "tester01@vmware.com",
		Password: password,
		Realname: "tester01",
		Comment:  "register",
	})
	if err != nil {
		log.Errorf("Error occurred when creating user: %v", err)
		return 1
	}

	rc := m.Run()
	clearAll()
	return rc
}

func clearAll() {
	tables := []string{"project_member",
		"project_metadata", "repository", "replication_policy",
		"registry", "project", "harbor_user"}
	for _, t := range tables {
		if err := ClearTable(t); err != nil {
			log.Errorf("Failed to clear table: %s,error: %v", t, err)
		}
	}
}

var targetID, policyID, policyID2, policyID3, jobID, jobID2, jobID3 int64

func TestGetOrmer(t *testing.T) {
	o := GetOrmer()
	if o == nil {
		t.Errorf("Error get ormer.")
	}
}
