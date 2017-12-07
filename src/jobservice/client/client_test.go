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

package client

import (
	"encoding/json"
	"net/http"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/vmware/harbor/src/common/utils/test"
	"github.com/vmware/harbor/src/jobservice/api"
)

var url string

func TestMain(m *testing.M) {
	requestMapping := []*test.RequestHandlerMapping{
		&test.RequestHandlerMapping{
			Method:  http.MethodPost,
			Pattern: "/api/jobs/replication",
			Handler: func(w http.ResponseWriter, r *http.Request) {
				replication := &api.ReplicationReq{}
				if err := json.NewDecoder(r.Body).Decode(replication); err != nil {
					w.WriteHeader(http.StatusInternalServerError)
				}
			},
		},
	}
	server := test.NewServer(requestMapping...)
	defer server.Close()

	url = server.URL

	os.Exit(m.Run())
}

func TestSubmitReplicationJob(t *testing.T) {
	client := NewDefaultClient(url)
	err := client.SubmitReplicationJob(&api.ReplicationReq{})
	assert.Nil(t, err)
}
