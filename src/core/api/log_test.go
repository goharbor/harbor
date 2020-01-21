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

	"github.com/goharbor/harbor/src/common/dao"
	"github.com/goharbor/harbor/src/common/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLogGet(t *testing.T) {
	fmt.Println("Testing Log API")

	var projectID int64 = 1
	username := "user_for_testing_log_api"
	repository := "repository_for_testing_log_api"
	tag := "tag_for_testing_log_api"
	operation := "op_for_test_log_api"
	now := time.Now()
	err := dao.AddAccessLog(models.AccessLog{
		ProjectID: projectID,
		Username:  username,
		RepoName:  repository,
		RepoTag:   tag,
		Operation: operation,
		OpTime:    now,
	})
	require.Nil(t, err)
	defer dao.GetOrmer().QueryTable(&models.AccessLog{}).
		Filter("username", username).Delete()

	url := "/api/logs"
	// 401
	cc := &codeCheckingCase{
		request: &testingRequest{
			method: http.MethodGet,
			url:    url,
		},
		code: http.StatusUnauthorized,
	}
	runCodeCheckingCases(t, cc)

	// 200, empty log list
	c := &testingRequest{
		method:     http.MethodGet,
		url:        url,
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
			Repository:     repository,
			Tag:            tag,
			Operation:      operation,
			BeginTimestamp: now.Add(-1 * time.Second).Unix(),
			EndTimestamp:   now.Add(1 * time.Second).Unix(),
		},
	}
	logs := []*models.AccessLog{}
	err = handleAndParse(c, &logs)
	require.Nil(t, err)
	require.Equal(t, 0, len(logs))

	// 200
	c.credential = projGuest
	err = handleAndParse(c, &logs)
	require.Nil(t, err)
	require.Equal(t, 1, len(logs))
	assert.Equal(t, projectID, logs[0].ProjectID)
	assert.Equal(t, username, logs[0].Username)
	assert.Equal(t, repository, logs[0].RepoName)
	assert.Equal(t, tag, logs[0].RepoTag)
	assert.Equal(t, operation, logs[0].Operation)

	// Limited Guest 200 && no logs
	c.credential = projLimitedGuest
	err = handleAndParse(c, &logs)
	require.Nil(t, err)
	require.Equal(t, 0, len(logs))
}
