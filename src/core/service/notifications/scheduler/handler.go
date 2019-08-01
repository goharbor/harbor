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

package scheduler

import (
	"encoding/json"
	"fmt"

	"github.com/goharbor/harbor/src/common/job/models"
	"github.com/goharbor/harbor/src/common/utils/log"
	"github.com/goharbor/harbor/src/core/api"
	"github.com/goharbor/harbor/src/pkg/scheduler/hook"
)

// Handler handles the scheduler requests
type Handler struct {
	api.BaseController
}

// Handle ...
func (h *Handler) Handle() {
	var data models.JobStatusChange
	// status update
	if len(data.CheckIn) == 0 {
		schedulerID, err := h.GetInt64FromPath(":id")
		if err != nil {
			log.Errorf("failed to get the schedule ID: %v", err)
			return
		}
		if err := hook.GlobalController.UpdateStatus(schedulerID, data.Status); err != nil {
			h.SendInternalServerError(fmt.Errorf("failed to update status of job %s: %v", data.JobID, err))
			return
		}
		return
	}

	// run callback function
	// just log the error message when handling check in request if got any error
	params := map[string]interface{}{}
	if err := json.Unmarshal([]byte(data.CheckIn), &params); err != nil {
		log.Errorf("failed to unmarshal parameters from check in message: %v", err)
		return
	}
	callbackFuncNameParam, exist := params["callback_func_name"]
	if !exist {
		log.Error("cannot get the parameter \"callback_func_name\" from the check in message")
		return
	}
	callbackFuncName, ok := callbackFuncNameParam.(string)
	if !ok || len(callbackFuncName) == 0 {
		log.Errorf("invalid \"callback_func_name\": %v", callbackFuncName)
		return
	}
	if err := hook.GlobalController.Run(callbackFuncName, params["params"]); err != nil {
		log.Errorf("failed to run the callback function %s: %v", callbackFuncName, err)
		return
	}
}
