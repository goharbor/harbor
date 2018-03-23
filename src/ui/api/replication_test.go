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
package api

import (
	"fmt"
	"net/http"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/vmware/harbor/src/common/dao"
	"github.com/vmware/harbor/src/common/models"
	"github.com/vmware/harbor/src/replication"
	api_models "github.com/vmware/harbor/src/ui/api/models"
)

const (
	replicationAPIBaseURL = "/api/replications"
)

func TestReplicationAPIPost(t *testing.T) {
	targetID, err := dao.AddRepTarget(
		models.RepTarget{
			Name:     "test_replication_target",
			URL:      "127.0.0.1",
			Username: "username",
			Password: "password",
		})
	require.Nil(t, err)
	defer dao.DeleteRepTarget(targetID)

	policyID, err := dao.AddRepPolicy(
		models.RepPolicy{
			Name:      "test_replication_policy",
			ProjectID: 1,
			TargetID:  targetID,
			Trigger:   fmt.Sprintf("{\"kind\":\"%s\"}", replication.TriggerKindManual),
		})
	require.Nil(t, err)
	defer dao.DeleteRepPolicy(policyID)

	cases := []*codeCheckingCase{
		// 401
		&codeCheckingCase{
			request: &testingRequest{
				method: http.MethodPost,
				url:    replicationAPIBaseURL,
				bodyJSON: &api_models.Replication{
					PolicyID: policyID,
				},
			},
			code: http.StatusUnauthorized,
		},
		// 404
		&codeCheckingCase{
			request: &testingRequest{
				method: http.MethodPost,
				url:    replicationAPIBaseURL,
				bodyJSON: &api_models.Replication{
					PolicyID: 10000,
				},
				credential: admin,
			},
			code: http.StatusNotFound,
		},
		// 200
		&codeCheckingCase{
			request: &testingRequest{
				method: http.MethodPost,
				url:    replicationAPIBaseURL,
				bodyJSON: &api_models.Replication{
					PolicyID: policyID,
				},
				credential: admin,
			},
			code: http.StatusOK,
		},
	}

	runCodeCheckingCases(t, cases...)
}
