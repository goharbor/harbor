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

package retention

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/goharbor/harbor/src/controller/event/operator"
	"github.com/goharbor/harbor/src/lib/log"
	"github.com/goharbor/harbor/src/pkg/scheduler"
)

func init() {
	err := scheduler.RegisterCallbackFunc(SchedulerCallback, retentionCallback)
	if err != nil {
		log.Fatalf("failed to register retention callback, %v", err)
	}
}

func retentionCallback(ctx context.Context, p string) error {
	param := &TriggerParam{}
	if err := json.Unmarshal([]byte(p), param); err != nil {
		return fmt.Errorf("failed to unmarshal the param: %v", err)
	}

	if param.Operator != "" {
		ctx = context.WithValue(ctx, operator.ContextKey{}, param.Operator)
	}
	_, err := Ctl.TriggerRetentionExec(ctx, param.PolicyID, param.Trigger, false)
	return err
}
