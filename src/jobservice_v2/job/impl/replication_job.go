// Copyright 2018 The Harbor Authors. All rights reserved.
package impl

import (
	"fmt"

	"github.com/vmware/harbor/src/jobservice_v2/env"
	"github.com/vmware/harbor/src/jobservice_v2/job"
)

//ReplicationJob is the job for replicating repositories.
type ReplicationJob struct {
	ctx       env.JobContext
	params    map[string]interface{}
	opCmdFunc job.CheckOPCmdFunc
}

//SetContext is implementation of same method in Interface.
func (rj *ReplicationJob) SetContext(ctx env.JobContext) {
	rj.ctx = ctx
	fmt.Printf("ReplicationJob context=%#v\n", rj.ctx)
}

//SetParams is implementation of same method in Interface.
func (rj *ReplicationJob) SetParams(params map[string]interface{}) error {
	rj.params = params
	fmt.Printf("ReplicationJob args: %v\n", rj.params)
	return nil
}

//SetCheckOPCmdFunc is implementation of same method in Interface.
func (rj *ReplicationJob) SetCheckOPCmdFunc(f job.CheckOPCmdFunc) {}

//MaxFails is implementation of same method in Interface.
func (rj *ReplicationJob) MaxFails() uint {
	return 2
}

//ParamsRequired is implementation of same method in Interface.
func (rj *ReplicationJob) ParamsRequired() bool {
	return true
}

//Run the replication logic here.
func (rj *ReplicationJob) Run() error {
	fmt.Println("=======Replication job running=======")
	return nil
}
