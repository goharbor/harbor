// Copyright Project Harbor Authors
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

package task

import (
	"context"
	"fmt"

	"github.com/goharbor/harbor/src/jobservice/job"
)

var (
	checkInProcessorRegistry              = map[string]CheckInProcessor{}
	statusChangePostFuncRegistry          = map[string]StatusChangePostFunc{}
	executionStatusChangePostFuncRegistry = map[string]ExecutionStatusChangePostFunc{}
)

// CheckInProcessor is the processor to process the check in data which is sent by jobservice via webhook
type CheckInProcessor func(ctx context.Context, task *Task, sc *job.StatusChange) (err error)

// StatusChangePostFunc is the function called after the task status changed
type StatusChangePostFunc func(ctx context.Context, taskID int64, status string) (err error)

// ExecutionStatusChangePostFunc is the function called after the execution status changed
type ExecutionStatusChangePostFunc func(ctx context.Context, executionID int64, status string) (err error)

// RegisterCheckInProcessor registers check in processor for the specific vendor type
func RegisterCheckInProcessor(vendorType string, processor CheckInProcessor) error {
	if _, exist := checkInProcessorRegistry[vendorType]; exist {
		return fmt.Errorf("check in processor for %s already exists", vendorType)
	}
	checkInProcessorRegistry[vendorType] = processor
	return nil
}

// RegisterTaskStatusChangePostFunc registers a task status change post function for the specific vendor type
func RegisterTaskStatusChangePostFunc(vendorType string, fc StatusChangePostFunc) error {
	if _, exist := statusChangePostFuncRegistry[vendorType]; exist {
		return fmt.Errorf("the task status change post function for %s already exists", vendorType)
	}
	statusChangePostFuncRegistry[vendorType] = fc
	return nil
}

// RegisterExecutionStatusChangePostFunc registers an execution status change post function for the specific vendor type
func RegisterExecutionStatusChangePostFunc(vendorType string, fc ExecutionStatusChangePostFunc) error {
	if _, exist := executionStatusChangePostFuncRegistry[vendorType]; exist {
		return fmt.Errorf("the execution status change post function for %s already exists", vendorType)
	}
	executionStatusChangePostFuncRegistry[vendorType] = fc
	return nil
}
