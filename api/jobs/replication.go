package api

import (
	"encoding/json"
	"fmt"
	"github.com/vmware/harbor/api"
	"github.com/vmware/harbor/dao"
	"github.com/vmware/harbor/job"
	"github.com/vmware/harbor/models"
	"github.com/vmware/harbor/utils/log"
	"io/ioutil"
	"net/http"
	"net/http/httputil"
	"os"
	"strconv"
	"strings"
)

type ReplicationJob struct {
	api.BaseAPI
}

type ReplicationReq struct {
	PolicyID int64 `json:"policy_id"`
}

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
	repoList, err := getRepoList(p.ProjectID)
	if err != nil {
		log.Errorf("Failed to get repository list, project id: %d, error: %v", p.ProjectID, err)
		rj.RenderError(http.StatusInternalServerError, err.Error())
		return
	}
	log.Debugf("repo list: %v", repoList)
	for _, repo := range repoList {
		j := models.RepJob{
			Repository: repo,
			PolicyID:   data.PolicyID,
			Operation:  models.RepOpTransfer,
		}
		log.Debugf("Creating job for repo: %s, policy: %d", repo, data.PolicyID)
		id, err := dao.AddRepJob(j)
		if err != nil {
			log.Errorf("Failed to insert job record, error: %v", err)
			rj.RenderError(http.StatusInternalServerError, err.Error())
			return
		}
		log.Debugf("Send job to scheduler, job id: %d", id)
		job.Schedule(id)
	}
}

// calls the api from UI to get repo list
func getRepoList(projectID int64) ([]string, error) {
	uiURL := os.Getenv("UI_URL")
	if len(uiURL) == 0 {
		uiURL = "ui"
	}
	if !strings.HasSuffix(uiURL, "/") {
		uiURL += "/"
	}
	//TODO:Use secret key instead
	uiUser := os.Getenv("UI_USR")
	if len(uiUser) == 0 {
		uiUser = "admin"
	}
	uiPwd := os.Getenv("UI_PWD")
	if len(uiPwd) == 0 {
		uiPwd = "Harbor12345"
	}
	client := &http.Client{}
	req, err := http.NewRequest("GET", uiURL+"api/repositories?project_id="+strconv.Itoa(int(projectID)), nil)
	if err != nil {
		log.Errorf("Error when creating request: %v")
		return nil, err
	}
	req.SetBasicAuth(uiUser, uiPwd)
	resp, err := client.Do(req)
	if err != nil {
		log.Errorf("Error when calling UI api to get repositories, error: %v", err)
		return nil, err
	}
	if resp.StatusCode != http.StatusOK {
		log.Errorf("Unexpected status code: %d", resp.StatusCode)
		dump, _ := httputil.DumpResponse(resp, true)
		log.Debugf("response: %q", dump)
		return nil, fmt.Errorf("Unexpected status code when getting repository list: %d", resp.StatusCode)
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Errorf("Failed to read the response body, error: %v", err)
		return nil, err
	}
	var repoList []string
	err = json.Unmarshal(body, &repoList)
	return repoList, err
}
