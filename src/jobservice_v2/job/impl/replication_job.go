// Copyright 2018 The Harbor Authors. All rights reserved.

package impl

import (
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/vmware/harbor/src/common/dao"
	"github.com/vmware/harbor/src/common/models"
	"github.com/vmware/harbor/src/jobservice_v2/opm"

	"github.com/vmware/harbor/src/jobservice_v2/errs"

	"github.com/vmware/harbor/src/jobservice_v2/env"
)

//ReplicationJob is the job for replicating repositories.
type ReplicationJob struct{}

//MaxFails is implementation of same method in Interface.
func (rj *ReplicationJob) MaxFails() uint {
	return 3
}

//ShouldRetry ...
func (rj *ReplicationJob) ShouldRetry() bool {
	return true
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
	logger := ctx.GetLogger()

	defer func() {
		logger.Info("I'm finished, exit!")
		fmt.Println("I'm finished, exit!")
	}()
	logger.Info("=======Replication job running=======")
	logger.Infof("params: %#v\n", params)
	logger.Infof("context: %#v\n", ctx)
	if v, ok := ctx.Get("email_from"); ok {
		fmt.Printf("Get prop form context: email_from=%s\n", v)
	}
	if u, err := dao.GetUser(models.User{}); err == nil {
		fmt.Printf("u=%#+v\n", u)
	}

	/*if 1 != 0 {
		return errors.New("I suicide")
	}*/

	logger.Info("check in 30%")
	ctx.Checkin("30%")
	time.Sleep(2 * time.Second)
	logger.Warning("check in 60%")
	ctx.Checkin("60%")
	time.Sleep(2 * time.Second)
	logger.Debug("check in 100%")
	ctx.Checkin("100%")
	time.Sleep(1 * time.Second)

	//HOLD ON FOR A WHILE
	logger.Error("Holding for 20 sec")
	<-time.After(10 * time.Second)
	//logger.Fatal("I'm back, check if I'm stopped/cancelled")

	if cmd, ok := ctx.OPCommand(); ok {
		logger.Infof("cmd=%s\n", cmd)
		if cmd == opm.CtlCommandCancel {
			logger.Info("exit for receiving cancel signal")
			return errs.JobCancelledError()
		}

		logger.Info("exit for receiving stop signal")
		return errs.JobStoppedError()
	}

	fmt.Println("I'm here")

	return nil
}
