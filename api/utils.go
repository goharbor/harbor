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

package api

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"strings"

	"github.com/vmware/harbor/dao"
	"github.com/vmware/harbor/models"
	"github.com/vmware/harbor/utils/log"
)

func checkProjectPermission(userID int, projectID int64) bool {
	roles, err := listRoles(userID, projectID)
	if err != nil {
		log.Errorf("error occurred in getProjectPermission: %v", err)
		return false
	}
	return len(roles) > 0
}

func hasProjectAdminRole(userID int, projectID int64) bool {
	roles, err := listRoles(userID, projectID)
	if err != nil {
		log.Errorf("error occurred in getProjectPermission: %v", err)
		return false
	}

	for _, role := range roles {
		if role.RoleID == models.PROJECTADMIN {
			return true
		}
	}

	return false
}

//sysadmin has all privileges to all projects
func listRoles(userID int, projectID int64) ([]models.Role, error) {
	roles := make([]models.Role, 0, 1)
	isSysAdmin, err := dao.IsAdminRole(userID)
	if err != nil {
		log.Errorf("failed to determine whether the user %d is system admin: %v", userID, err)
		return roles, err
	}
	if isSysAdmin {
		role, err := dao.GetRoleByID(models.PROJECTADMIN)
		if err != nil {
			log.Errorf("failed to get role %d: %v", models.PROJECTADMIN, err)
			return roles, err
		}
		roles = append(roles, *role)
		return roles, nil
	}

	rs, err := dao.GetUserProjectRoles(userID, projectID)
	if err != nil {
		log.Errorf("failed to get user %d 's roles for project %d: %v", userID, projectID, err)
		return roles, err
	}
	roles = append(roles, rs...)
	return roles, nil
}

func checkUserExists(name string) int {
	u, err := dao.GetUser(models.User{Username: name})
	if err != nil {
		log.Errorf("Error occurred in GetUser, error: %v", err)
		return 0
	}
	if u != nil {
		return u.UserID
	}
	return 0
}

// TriggerReplication triggers the replication according to the policy
func TriggerReplication(policyID int64, repository string,
	tags []string, operation string) error {
	data := struct {
		PolicyID  int64    `json:"policy_id"`
		Repo      string   `json:"repository"`
		Operation string   `json:"operation"`
		TagList   []string `json:"tags"`
	}{
		PolicyID:  policyID,
		Repo:      repository,
		TagList:   tags,
		Operation: operation,
	}

	b, err := json.Marshal(&data)
	if err != nil {
		return err
	}

	url := buildReplicationURL()

	resp, err := http.DefaultClient.Post(url, "application/json", bytes.NewBuffer(b))
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusOK {
		return nil
	}

	b, err = ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	return fmt.Errorf("%d %s", resp.StatusCode, string(b))
}

// GetPoliciesByRepository returns policies according the repository
func GetPoliciesByRepository(repository string) ([]*models.RepPolicy, error) {
	repository = strings.TrimSpace(repository)
	repository = strings.TrimRight(repository, "/")
	projectName := repository[:strings.LastIndex(repository, "/")]

	project, err := dao.GetProjectByName(projectName)
	if err != nil {
		return nil, err
	}

	policies, err := dao.GetRepPolicyByProject(project.ProjectID)
	if err != nil {
		return nil, err
	}

	return policies, nil
}

// TriggerReplicationByRepository triggers the replication according to the repository
func TriggerReplicationByRepository(repository string, tags []string, operation string) {
	policies, err := GetPoliciesByRepository(repository)
	if err != nil {
		log.Errorf("failed to get policies for repository %s: %v", repository, err)
		return
	}

	for _, policy := range policies {
		if policy.Enabled == 0 {
			continue
		}
		if err := TriggerReplication(policy.ID, repository, tags, operation); err != nil {
			log.Errorf("failed to trigger replication of policy %d for %s: %v", policy.ID, repository, err)
		} else {
			log.Infof("replication of policy %d for %s triggered", policy.ID, repository)
		}
	}
}

func postReplicationAction(policyID int64, acton string) error {
	data := struct {
		PolicyID int64  `json:"policy_id"`
		Action   string `json:"action"`
	}{
		PolicyID: policyID,
		Action:   acton,
	}

	b, err := json.Marshal(&data)
	if err != nil {
		return err
	}

	url := buildReplicationActionURL()

	resp, err := http.DefaultClient.Post(url, "application/json", bytes.NewBuffer(b))
	if err != nil {
		return err
	}

	defer resp.Body.Close()

	if resp.StatusCode == http.StatusOK {
		return nil
	}

	b, err = ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	return fmt.Errorf("%d %s", resp.StatusCode, string(b))
}

func buildReplicationURL() string {
	url := getJobServiceURL()
	return fmt.Sprintf("%s/api/jobs/replication", url)
}

func buildJobLogURL(jobID string) string {
	url := getJobServiceURL()
	return fmt.Sprintf("%s/api/jobs/replication/%s/log", url, jobID)
}

func buildReplicationActionURL() string {
	url := getJobServiceURL()
	return fmt.Sprintf("%s/api/jobs/replication/actions", url)
}

func getJobServiceURL() string {
	url := os.Getenv("JOB_SERVICE_URL")
	url = strings.TrimSpace(url)
	url = strings.TrimRight(url, "/")

	if len(url) == 0 {
		url = "http://jobservice"
	}

	return url
}
