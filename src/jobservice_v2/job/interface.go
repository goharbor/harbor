// Copyright 2018 The Harbor Authors. All rights reserved.

package job

import "github.com/vmware/harbor/src/jobservice_v2/env"

//CheckOPCmdFunc is the function to check if the related operation commands
//like STOP or CANCEL is fired for the specified job. If yes, return the
//command code for job to determin if take corresponding action.
type CheckOPCmdFunc func(string) (uint, bool)

//Interface defines the related injection and run entry methods.
type Interface interface {
	//Declare how many times the job can be retried if failed.
	//
	//Return:
	// uint: the failure count allowed
	MaxFails() uint

	//Indicate whether the parameters of job are valid.
	//
	//Return:
	// error if parameters are not valid. NOTES: If no parameters needed, directly return nil.
	Validate(params map[string]interface{}) error

	//Run the business logic here.
	//The related arguments will be injected by the workerpool.
	//
	//ctx env.JobContext            : Job execution context.
	//params map[string]interface{} : parameters with key-pair style for the job execution.
	//f	CheckOPCmdFunc: check function reference.
	//
	//Returns:
	//  error if failed to run. NOTES: If job is stopped or cancelled, a specified error should be returned
	//
	Run(ctx env.JobContext, params map[string]interface{}, f CheckOPCmdFunc) error
}
