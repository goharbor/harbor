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

//go:generate mockery -dir ./dao -name TaskDAO -output . -outpkg task -filename mock_task_dao_test.go -structname mockTaskDAO
//go:generate mockery -dir ./dao -name ExecutionDAO -output . -outpkg task -filename mock_execution_dao_test.go -structname mockExecutionDAO
// Need to modify the generated mock code manually to avoid the compile error: https://github.com/vektra/mockery/issues/293
/*
func (_m *mockTaskManager) Create(ctx context.Context, executionID int64, job *Job, extraAttrs ...map[string]interface{}) (int64, error) {
	_va := make([]interface{}, len(extraAttrs))
	for _i := range extraAttrs {
		_va[_i] = extraAttrs[_i]
	}
	var _ca []interface{}
	_ca = append(_ca, ctx, executionID, job)
	_ca = append(_ca, _va...)
	ret := _m.Called(_ca...)
	...
*/
//go:generate mockery -name Manager -output . -outpkg task -filename mock_task_manager_test.go -structname mockTaskManager -inpkg
//go:generate mockery -dir ../../common/job -name Client -output . -outpkg task -filename mock_jobservice_client_test.go -structname mockJobserviceClient
