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

// Package utils contains methods to support security, cache, and webhook functions.
package utils

import (
	"github.com/goharbor/harbor/src/common/job"
	"github.com/goharbor/harbor/src/core/config"

	"sync"
)

var (
	cl               sync.Mutex
	jobServiceClient job.Client
)

// GetJobServiceClient returns the job service client instance.
func GetJobServiceClient() job.Client {
	cl.Lock()
	defer cl.Unlock()
	if jobServiceClient == nil {
		jobServiceClient = job.NewDefaultClient(config.InternalJobServiceURL(), config.CoreSecret())
	}
	return jobServiceClient
}
