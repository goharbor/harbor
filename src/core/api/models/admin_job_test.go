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
package models

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/goharbor/harbor/src/common"
	common_job "github.com/goharbor/harbor/src/common/job"
	"github.com/goharbor/harbor/src/common/utils/test"
	"github.com/goharbor/harbor/src/core/config"
	"os"
	"strings"
)

var testConfig = map[string]interface{}{
	common.DefaultCoreEndpoint: "test",
}

func TestMain(m *testing.M) {

	test.InitDatabaseFromEnv()
	config.Init()
	config.Upload(testConfig)
	os.Exit(m.Run())

}

func TestToJob(t *testing.T) {

	adminJobSchedule := AdminJobSchedule{
		Schedule: &ScheduleParam{
			Type: "Daily",
			Cron: "20 3 0 * * *",
		},
	}

	adminjob := &AdminJobReq{
		Name:             common_job.ImageGC,
		AdminJobSchedule: adminJobSchedule,
	}

	job := adminjob.ToJob()
	assert.Equal(t, job.Name, "IMAGE_GC")
	assert.Equal(t, job.Metadata.JobKind, common_job.JobKindPeriodic)
	assert.Equal(t, job.Metadata.Cron, "20 3 0 * * *")
}

func TestToJobManual(t *testing.T) {

	adminJobSchedule := AdminJobSchedule{
		Schedule: &ScheduleParam{
			Type: "Manual",
		},
	}

	adminjob := &AdminJobReq{
		AdminJobSchedule: adminJobSchedule,
		Name:             common_job.ImageGC,
	}

	job := adminjob.ToJob()
	assert.Equal(t, job.Name, "IMAGE_GC")
	assert.Equal(t, job.Metadata.JobKind, common_job.JobKindGeneric)
}

func TestIsPeriodic(t *testing.T) {

	adminJobSchedule := AdminJobSchedule{
		Schedule: &ScheduleParam{
			Type: "Daily",
			Cron: "20 3 0 * * *",
		},
	}

	adminjob := &AdminJobReq{
		AdminJobSchedule: adminJobSchedule,
	}

	isPeriodic := adminjob.IsPeriodic()
	assert.Equal(t, isPeriodic, true)
}

func TestJobKind(t *testing.T) {

	adminJobSchedule := AdminJobSchedule{
		Schedule: &ScheduleParam{
			Type: "Daily",
			Cron: "20 3 0 * * *",
		},
	}

	adminjob := &AdminJobReq{
		AdminJobSchedule: adminJobSchedule,
	}

	kind := adminjob.JobKind()
	assert.Equal(t, kind, "Periodic")

	adminJobSchedule1 := AdminJobSchedule{
		Schedule: &ScheduleParam{
			Type: "Manual",
		},
	}
	adminjob1 := &AdminJobReq{
		AdminJobSchedule: adminJobSchedule1,
	}
	kind1 := adminjob1.JobKind()
	assert.Equal(t, kind1, "Generic")
}

func TestCronString(t *testing.T) {

	adminJobSchedule := AdminJobSchedule{
		Schedule: &ScheduleParam{
			Type: "Daily",
			Cron: "20 3 0 * * *",
		},
	}

	adminjob := &AdminJobReq{
		AdminJobSchedule: adminJobSchedule,
	}
	cronStr := adminjob.CronString()
	assert.True(t, strings.EqualFold(cronStr, "{\"type\":\"Daily\",\"Cron\":\"20 3 0 * * *\"}"))
}
