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

package demojob

import (
	"time"

	"github.com/vmware/harbor/src/jobservice/env"
	"github.com/vmware/harbor/src/jobservice/logger"
)

// GarbageCollector is the struct to run registry's garbage collection
type GarbageCollector struct {
	logger logger.Interface
}

// MaxFails implements the interface in job/Interface
func (gc *GarbageCollector) MaxFails() uint {
	return 1
}

// ShouldRetry implements the interface in job/Interface
func (gc *GarbageCollector) ShouldRetry() bool {
	return false
}

// Validate implements the interface in job/Interface
func (gc *GarbageCollector) Validate(params map[string]interface{}) error {
	return nil
}

// Run implements the interface in job/Interface
func (gc *GarbageCollector) Run(ctx env.JobContext, params map[string]interface{}) error {
	if err := gc.init(ctx); err != nil {
		return err
	}
	gc.logger.Infof("success to run gc in job.")
	time.Sleep(20 * time.Second)
	gc.logger.Infof("doing...")
	time.Sleep(20 * time.Second)
	gc.logger.Infof("done gc in job.")
	return nil
}

func (gc *GarbageCollector) init(ctx env.JobContext) error {
	gc.logger = ctx.GetLogger()
	return nil
}
