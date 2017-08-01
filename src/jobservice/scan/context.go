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

package scan

import (
	"github.com/vmware/harbor/src/common/models"
	"github.com/vmware/harbor/src/common/utils/clair"
	"github.com/vmware/harbor/src/common/utils/log"
)

const (
	// StateInitialize in this state the handler will initialize the job context.
	StateInitialize = "initialize"
	// StateScanLayer in this state the handler will POST layer  of clair to scan layer by layer of the image.
	StateScanLayer = "scanlayer"
	// StateSummarize in this state, the layers are scanned by clair it will call clair API to update vulnerability overview in Harbor DB. After this state, the job is finished.
	StateSummarize = "summarize"
)

//JobContext is for sharing data across handlers in a execution of a scan job.
type JobContext struct {
	JobID      int64
	Repository string
	Tag        string
	Digest     string
	//The array of data object to set as request body for layer scan.
	layers  []models.ClairLayer
	current int
	//token for accessing the registry
	token       string
	clairClient *clair.Client
	Logger      *log.Logger
}
