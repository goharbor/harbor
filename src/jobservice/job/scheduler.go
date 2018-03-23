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

package job

import (
	"time"

	"github.com/vmware/harbor/src/common/utils/log"
)

var jobQueue = make(chan Job)

// Schedule put a job id into job queue.
func Schedule(j Job) {
	jobQueue <- j
}

// Reschedule is called by statemachine to retry a job
func Reschedule(j Job) {
	log.Debugf("Job %v will be rescheduled in 5 minutes", j)
	time.Sleep(5 * time.Minute)
	log.Debugf("Rescheduling job %v", j)
	Schedule(j)
}
