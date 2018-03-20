// Copyright 2018 The Harbor Authors. All rights reserved.

package impl

import (
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/vmware/harbor/src/jobservice_v2/errs"

	"github.com/vmware/harbor/src/jobservice_v2/env"
)

//ReplicationJob is the job for replicating repositories.
type ReplicationJob struct{}

//MaxFails is implementation of same method in Interface.
func (rj *ReplicationJob) MaxFails() uint {
	return 2
}

//Validate is implementation of same method in Interface.
func (rj *ReplicationJob) Validate(params map[string]interface{}) error {
	if params == nil || len(params) == 0 {
		return errors.New("parameters required for replication job")
	}
	name, ok := params["image"]
	if !ok {
		return errors.New("missing parameter 'image'")
	}

	if !strings.HasPrefix(name.(string), "demo") {
		return fmt.Errorf("expected '%s' but got '%s'", "demo steven", name)
	}

	return nil
}

//Run the replication logic here.
func (rj *ReplicationJob) Run(ctx env.JobContext, params map[string]interface{}) error {
	defer func() {
		fmt.Println("I'm finished, exit!")
	}()
	fmt.Println("=======Replication job running=======")
	fmt.Printf("params: %#v\n", params)
	fmt.Printf("context: %#v\n", ctx)

	//HOLD ON FOR A WHILE
	fmt.Println("Holding for 20 sec")
	<-time.After(20 * time.Second)
	fmt.Println("I'm back, check if I'm stopped")

	if cmd, ok := ctx.OPCommand(); ok {
		fmt.Printf("cmd=%s\n", cmd)
		return errs.JobStoppedError()
	}

	return nil
}
