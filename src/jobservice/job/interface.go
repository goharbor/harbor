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

package job

// Interface defines the related injection and run entry methods.
type Interface interface {
	// Declare how many times the job can be retried if failed.
	//
	// Return:
	// unit: the failure count allowed. If it is set to 0, then default value 4 is used.
	MaxFails() uint

	// Tell the worker worker if retry the failed job when the fails is
	// still less that the number declared by the method 'MaxFails'.
	//
	// Returns:
	//  true for retry and false for none-retry
	ShouldRetry() bool

	// Indicate whether the parameters of job are valid.
	//
	// Return:
	// error if parameters are not valid. NOTES: If no parameters needed, directly return nil.
	Validate(params Parameters) error

	// Run the business logic here.
	// The related arguments will be injected by the workerpool.
	//
	// ctx Context                   : Job execution context.
	// params map[string]interface{} : parameters with key-pair style for the job execution.
	//
	// Returns:
	//  error if failed to run. NOTES: If job is stopped or cancelled, a specified error should be returned
	//
	Run(ctx Context, params Parameters) error
}
