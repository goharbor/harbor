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
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httputil"
	"strconv"

	"github.com/vmware/harbor/api"
	"github.com/vmware/harbor/dao"
	"github.com/vmware/harbor/job"
	"github.com/vmware/harbor/job/config"
	"github.com/vmware/harbor/job/utils"
	"github.com/vmware/harbor/models"
	"github.com/vmware/harbor/utils/log"
)

// ReplicationJob handles /api/replicationJobs /api/replicationJobs/:id/log
// /api/replicationJobs/actions
type ReplicationJob struct {
	api.BaseAPI
}

// ReplicationReq holds informations of request for /api/replicationJobs
type ReplicationReq struct {
	PolicyID  int64    `json:"policy_id"`
	Repo      string   `json:"repository"`
	Operation string   `json:"operation"`
	TagList   []string `json:"tags"`
}

// Prepare ...
func (rj *ReplicationJob) Prepare() {
	rj.authenticate()
}

func (rj *ReplicationJob) authenticate() {
	cookie, err := rj.Ctx.Request.Cookie(models.UISecretCookie)
	if err != nil && err != http.ErrNoCookie {
		log.Errorf("failed to get cookie %s: %v", models.UISecretCookie, err)
		rj.CustomAbort(http.StatusInternalServerError, "")
	}

	if err == http.ErrNoCookie {
		rj.CustomAbort(http.StatusUnauthorized, "")
	}

	if cookie.Value != config.UISecret() {
		rj.CustomAbort(http.StatusForbidden, "")
	}
}

// Post creates replication jobs according to the policy.
func (rj *ReplicationJob) Post() {
	var data ReplicationReq
	rj.DecodeJSONReq(&data)
	log.Debugf("data: %+v", data)
	p, err := dao.GetRepPolicy(data.PolicyID)
	if err != nil {
		log.Errorf("Failed to get policy, error: %v", err)
		rj.RenderError(http.StatusInternalServerError, fmt.Sprintf("Failed to get policy, id: %d", data.PolicyID))
		return
	}
	if p == nil {
		log.Errorf("Policy not found, id: %d", data.PolicyID)
		rj.RenderError(http.StatusNotFound, fmt.Sprintf("Policy not found, id: %d", data.PolicyID))
		return
	}
	if len(data.Repo) == 0 { // sync all repositories
		repoList, err := getRepoList(p.ProjectID)
		if err != nil {
			log.Errorf("Failed to get repository list, project id: %d, error: %v", p.ProjectID, err)
			rj.RenderError(http.StatusInternalServerError, err.Error())
			return
		}
		log.Debugf("repo list: %v", repoList)
		for _, repo := range repoList {
			err := rj.addJob(repo, data.PolicyID, models.RepOpTransfer)
			if err != nil {
				log.Errorf("Failed to insert job record, error: %v", err)
				rj.RenderError(http.StatusInternalServerError, err.Error())
				return
			}
		}
	} else { // sync a single repository
		var op string
		if len(data.Operation) > 0 {
			op = data.Operation
		} else {
			op = models.RepOpTransfer
		}
		err := rj.addJob(data.Repo, data.PolicyID, op, data.TagList...)
		if err != nil {
			log.Errorf("Failed to insert job record, error: %v", err)
			rj.RenderError(http.StatusInternalServerError, err.Error())
			return
		}
	}
}

func (rj *ReplicationJob) addJob(repo string, policyID int64, operation string, tags ...string) error {
	j := models.RepJob{
		Repository: repo,
		PolicyID:   policyID,
		Operation:  operation,
		TagList:    tags,
	}
	log.Debugf("Creating job for repo: %s, policy: %d", repo, policyID)
	id, err := dao.AddRepJob(j)
	if err != nil {
		return err
	}
	log.Debugf("Send job to scheduler, job id: %d", id)
	job.Schedule(id)
	return nil
}

// RepActionReq holds informations of request for /api/replicationJobs/actions
type RepActionReq struct {
	PolicyID int64  `json:"policy_id"`
	Action   string `json:"action"`
}

// HandleAction supports some operations to all the jobs of one policy
func (rj *ReplicationJob) HandleAction() {
	var data RepActionReq
	rj.DecodeJSONReq(&data)
	//Currently only support stop action
	if data.Action != "stop" {
		log.Errorf("Unrecognized action: %s", data.Action)
		rj.RenderError(http.StatusBadRequest, fmt.Sprintf("Unrecongized action: %s", data.Action))
		return
	}
	jobs, err := dao.GetRepJobToStop(data.PolicyID)
	if err != nil {
		log.Errorf("Failed to get jobs to stop, error: %v", err)
		rj.RenderError(http.StatusInternalServerError, "Faild to get jobs to stop")
		return
	}
	var jobIDList []int64
	for _, j := range jobs {
		jobIDList = append(jobIDList, j.ID)
	}
	job.WorkerPool.StopJobs(jobIDList)
}

// GetLog gets logs of the job
func (rj *ReplicationJob) GetLog() {
	idStr := rj.Ctx.Input.Param(":id")
	jid, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		log.Errorf("Error parsing job id: %s, error: %v", idStr, err)
		rj.RenderError(http.StatusBadRequest, "Invalid job id")
		return
	}
	logFile := utils.GetJobLogPath(jid)
	rj.Ctx.Output.Download(logFile)
}

// calls the api from UI to get repo list
func getRepoList(projectID int64) ([]string, error) {
	/*
		uiUser := os.Getenv("UI_USR")
		if len(uiUser) == 0 {
			uiUser = "admin"
		}
		uiPwd := os.Getenv("UI_PWD")
		if len(uiPwd) == 0 {
			uiPwd = "Harbor12345"
		}
	*/
	uiURL := config.LocalUIURL()
	client := &http.Client{}
	req, err := http.NewRequest("GET", uiURL+"/api/repositories?project_id="+strconv.Itoa(int(projectID)), nil)
	if err != nil {
		log.Errorf("Error when creating request: %v", err)
		return nil, err
	}
	//req.SetBasicAuth(uiUser, uiPwd)
	req.AddCookie(&http.Cookie{Name: models.UISecretCookie, Value: config.UISecret()})
	//dump, err := httputil.DumpRequest(req, true)
	//log.Debugf("req: %q", dump)
	resp, err := client.Do(req)
	if err != nil {
		log.Errorf("Error when calling UI api to get repositories, error: %v", err)
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		log.Errorf("Unexpected status code: %d", resp.StatusCode)
		dump, _ := httputil.DumpResponse(resp, true)
		log.Debugf("response: %q", dump)
		return nil, fmt.Errorf("Unexpected status code when getting repository list: %d", resp.StatusCode)
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Errorf("Failed to read the response body, error: %v", err)
		return nil, err
	}
	var repoList []string
	err = json.Unmarshal(body, &repoList)
	return repoList, err
}
