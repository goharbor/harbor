package retention

import (
	"context"
	"encoding/json"
	"fmt"

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
	_, err := Ctl.TriggerRetentionExec(ctx, param.PolicyID, param.Trigger, false)
	return err
}
