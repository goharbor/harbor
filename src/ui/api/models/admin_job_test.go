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
package models

import (
	"log"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/vmware/harbor/src/common"
	"github.com/vmware/harbor/src/common/utils/test"
)

var adminServerTestConfig = map[string]interface{}{
	common.DefaultUIEndpoint: "test",
}

func TestMain(m *testing.M) {
	server, err := test.NewAdminserver(adminServerTestConfig)
	if err != nil {
		log.Fatalf("failed to create a mock admin server: %v", err)
	}
	defer server.Close()

}

func TestToJob(t *testing.T) {
	adminjob := &AdminJobReq{
		Name: "GC",
		Kind: "Generic",
	}

	job, err := adminjob.ToJob()
	assert.Nil(t, err)
	assert.Equal(t, job.Name, "IMAGE_GC")
	assert.Equal(t, job.Metadata.JobKind, adminjob.Kind)
}

func TestToJobErr(t *testing.T) {
	adminjob := &AdminJobReq{
		Name: "errJob",
		Kind: "Generic",
	}

	_, err := adminjob.ToJob()
	assert.NotNil(t, err)
}

func TestToJobSchdule(t *testing.T) {
	schedule := &ScheduleParam{
		Type:    "Daily",
		Offtime: 200,
	}

	adminJobReq := &AdminJobReq{
		Name:     "GC",
		Kind:     "Periodic",
		Schedule: schedule,
	}

	job, err := adminJobReq.ToJob()
	assert.Nil(t, err)
	assert.Equal(t, job.Name, "IMAGE_GC")
	assert.Equal(t, job.Metadata.JobKind, adminJobReq.Kind)
	assert.Equal(t, job.Metadata.Cron, "20 3 0 * * *")
}
