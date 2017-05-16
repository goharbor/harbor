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
	"testing"

	"github.com/vmware/harbor/src/common/models"
)

func TestDeleteProject(t *testing.T) {
	name := "project_for_test"
	project := models.Project{
		OwnerID: currentUser.UserID,
		Name:    name,
	}

	id, err := AddProject(project)
	if err != nil {
		t.Fatalf("failed to add project: %v", err)
	}
	defer func() {
		if err := delProjPermanent(id); err != nil {
			t.Errorf("failed to clear up project %d: %v", id, err)
		}
	}()

	if err = DeleteProject(id); err != nil {
		t.Fatalf("failed to delete project: %v", err)
	}

	p := &models.Project{}
	if err = GetOrmer().Raw(`select * from project where project_id = ?`, id).
		QueryRow(p); err != nil {
		t.Fatalf("failed to get project: %v", err)
	}

	if p.Deleted != 1 {
		t.Errorf("unexpeced deleted column: %d != %d", p.Deleted, 1)
	}

	deletedName := fmt.Sprintf("%s#%d", name, id)
	if p.Name != deletedName {
		t.Errorf("unexpected name: %s != %s", p.Name, deletedName)
	}

}

func delProjPermanent(id int64) error {
	_, err := GetOrmer().QueryTable("access_log").
		Filter("ProjectID", id).
		Delete()
	if err != nil {
		return err
	}

	_, err = GetOrmer().Raw(`delete from project_member 
		where project_id = ?`, id).Exec()
	if err != nil {
		return err
	}

	_, err = GetOrmer().QueryTable("project").
		Filter("ProjectID", id).
		Delete()
	return err
}
