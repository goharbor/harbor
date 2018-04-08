// Copyright 2018 The Harbor Authors. All rights reserved.

package impl

import (
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/vmware/harbor/src/common/dao"
	"github.com/vmware/harbor/src/common/models"
	"github.com/vmware/harbor/src/jobservice/opm"

	"github.com/vmware/harbor/src/jobservice/errs"

	"github.com/vmware/harbor/src/jobservice/env"
)

//DemoJob is the job to demostrate the job interface.
type DemoJob struct{}

//MaxFails is implementation of same method in Interface.
func (dj *DemoJob) MaxFails() uint {
	return 3
}

//ShouldRetry ...
func (dj *DemoJob) ShouldRetry() bool {
	return true
}

//Validate is implementation of same method in Interface.
func (dj *DemoJob) Validate(params map[string]interface{}) error {
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
func (dj *DemoJob) Run(ctx env.JobContext, params map[string]interface{}) error {
	logger := ctx.GetLogger()

	defer func() {
		logger.Info("I'm finished, exit!")
		fmt.Println("I'm finished, exit!")
	}()
	fmt.Println("I'm running")
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
	//runtime error
	//var runtime_err error = nil
	//fmt.Println(runtime_err.Error())

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
	<-time.After(15 * time.Second)
	//logger.Fatal("I'm back, check if I'm stopped/cancelled")

	if cmd, ok := ctx.OPCommand(); ok {
		logger.Infof("cmd=%s\n", cmd)
		fmt.Printf("Receive OP command: %s\n", cmd)
		if cmd == opm.CtlCommandCancel {
			logger.Info("exit for receiving cancel signal")
			return errs.JobCancelledError()
		}

		logger.Info("exit for receiving stop signal")
		return errs.JobStoppedError()
	}

	fmt.Println("I'm close to end")

	return nil
}
