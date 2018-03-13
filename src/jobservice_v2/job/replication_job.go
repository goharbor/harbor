// Copyright 2018 The Harbor Authors. All rights reserved.

package job

import (
	"fmt"
)

//ReplicationJob is the job for replicating repositories.
type ReplicationJob struct {
	ctx       Context
	params    map[string]interface{}
	opCmdFunc CheckOPCmdFunc
}

//SetContext is implementation of same method in Interface.
func (rj *ReplicationJob) SetContext(ctx Context) {
	rj.ctx = ctx
	fmt.Printf("ReplicationJob context=%#v\n", rj.ctx)
}

//SetParams is implementation of same method in Interface.
func (rj *ReplicationJob) SetParams(params map[string]interface{}) {
	rj.params = params
	fmt.Printf("ReplicationJob args: %v\n", rj.params)
}

//SetCheckOPCmdFunc is implementation of same method in Interface.
func (rj *ReplicationJob) SetCheckOPCmdFunc(f CheckOPCmdFunc) {}

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
