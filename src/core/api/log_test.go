// Copyright 2018 Project Harbor Authors
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
package api

import (
	"fmt"
	"net/http"
	"testing"
	"time"

	"github.com/goharbor/harbor/src/common"
	"github.com/goharbor/harbor/src/common/dao"
	"github.com/goharbor/harbor/src/common/dao/project"
	"github.com/goharbor/harbor/src/common/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const (
	username             = "user_for_testing_log_api"
	repositoryName       = "repository_for_testing_log_api"
	repoTag              = "tag_for_testing_log_api"
	operation            = "op_for_test_log_api"
	apiURL               = "/api/logs"
	projectID      int64 = 1
)

func TestLogGetUnauthenticated(t *testing.T) {
	fmt.Println("Testing Log API")

	cc := &codeCheckingCase{
		request: &testingRequest{
			method: http.MethodGet,
			url:    apiURL,
		},
		code: http.StatusUnauthorized,
	}
	runCodeCheckingCases(t, cc)
}

func TestLogGetSuccessEmpty(t *testing.T) {
	request := makeRequest()
	logs := []*models.AccessLog{}
	err := handleAndParse(request, &logs)
	require.Nil(t, err)
	require.Equal(t, 0, len(logs))
}

func TestLogGetSuccessEntry(t *testing.T) {
	makeLogRecord(t, 1)
	defer eraseLogRecord()

	logs := []*models.AccessLog{}
	request := makeRequest()
	request.credential = projGuest

	err := handleAndParse(request, &logs)
	require.Nil(t, err)
	require.Equal(t, 1, len(logs))
	assert.Equal(t, projectID, logs[0].ProjectID)
	assert.Equal(t, username, logs[0].Username)
	assert.Equal(t, repositoryName, logs[0].RepoName)
	assert.Equal(t, repoTag, logs[0].RepoTag)
	assert.Equal(t, operation, logs[0].Operation)
}

func TestGuestAccess(t *testing.T) {
	projektID, err := dao.AddProject(models.Project{Name: "private", OwnerID: 1, Metadata: map[string]string{models.ProMetaPublic: "false"}})
	require.Nil(t, err)

	makeLogRecord(t, projektID)
	defer eraseLogRecord()

	userID, err := dao.Register(models.User{
		Username: "guest",
		Email:    "guest@guest.com",
		Password: "guest",
	})

	_, err = project.AddProjectMember(models.Member{
		ProjectID:  projektID,
		Role:       models.GUEST,
		EntityID:   int(userID),
		EntityType: common.UserMember,
	})

	require.Nil(t, err)

	request := makeRequest()
	request.credential = &usrInfo{Name: "guest", Passwd: "guest"}
	request.queryStruct = struct {
		Username string `url:"username"`
	}{
		Username: username,
	}
	logs := []*models.AccessLog{}

	err = handleAndParse(request, &logs)
	require.Nil(t, err)

	require.Equal(t, 0, len(logs))
}

func makeLogRecord(t *testing.T, newProjectID int64) {
	err := dao.AddAccessLog(models.AccessLog{
		ProjectID: newProjectID,
		Username:  username,
		RepoName:  repositoryName,
		RepoTag:   repoTag,
		Operation: operation,
		OpTime:    time.Now(),
	})
	require.Nil(t, err)
}

func eraseLogRecord() {
	_, _ = dao.GetOrmer().QueryTable(&models.AccessLog{}).
		Filter("username", username).Delete()
}

func makeRequest() *testingRequest {
	now := time.Now()
	return &testingRequest{
		method:     http.MethodGet,
		url:        apiURL,
		credential: nonSysAdmin,
		queryStruct: struct {
			Username       string `url:"username"`
			Repository     string `url:"repository"`
			Tag            string `url:"tag"`
			Operation      string `url:"operation"`
			BeginTimestamp int64  `url:"begin_timestamp"`
			EndTimestamp   int64  `url:"end_timestamp"`
		}{
			Username:       username,
			Repository:     repositoryName,
			Tag:            repoTag,
			Operation:      operation,
			BeginTimestamp: now.Add(-1 * time.Second).Unix(),
			EndTimestamp:   now.Add(1 * time.Second).Unix(),
		},
	}
}
