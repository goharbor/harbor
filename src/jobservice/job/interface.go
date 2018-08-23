// Copyright 2018 The Harbor Authors. All rights reserved.

package job

import "github.com/goharbor/harbor/src/jobservice/env"

//CheckOPCmdFunc is the function to check if the related operation commands
//like STOP or CANCEL is fired for the specified job. If yes, return the
//command code for job to determine if take corresponding action.
type CheckOPCmdFunc func() (string, bool)

//CheckInFunc is designed for job to report more detailed progress info
type CheckInFunc func(message string)

//Interface defines the related injection and run entry methods.
type Interface interface {
	//Declare how many times the job can be retried if failed.
	//
	//Return:
	// uint: the failure count allowed. If it is set to 0, then default value 4 is used.
	MaxFails() uint

	//Tell the worker pool if retry the failed job when the fails is
	//still less that the number declared by the method 'MaxFails'.
	//
	//Returns:
	//  true for retry and false for none-retry
	ShouldRetry() bool

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
	//
	//Returns:
	//  error if failed to run. NOTES: If job is stopped or cancelled, a specified error should be returned
	//
	Run(ctx env.JobContext, params map[string]interface{}) error
}
