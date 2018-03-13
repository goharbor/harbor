// Copyright 2018 The Harbor Authors. All rights reserved.

package job

//CheckOPCmdFunc is the function to check if the related operation commands
//like STOP or CANCEL is fired for the specified job. If yes, return the
//command code for job to determin if take corresponding action.
type CheckOPCmdFunc func(string) (uint, bool)

//Interface defines the related injection and run entry methods.
type Interface interface {
	//SetContext used to inject the job context if needed.
	//
	//ctx	Context: Job execution context.
	SetContext(ctx Context)

	//Pass parameters via this method if have.
	//
	//params	map[string]interface{}: parameters with key-pair style for the job execution.
	SetParams(params map[string]interface{})

	//Inject the func into the job for OP command check.
	//
	//f	CheckOPCmdFunc: check function reference.
	SetCheckOPCmdFunc(f CheckOPCmdFunc)

	//Declare how many times the job can be retried if failed.
	//
	//Return:
	// uint: the failure count allowed
	MaxFails() uint

	//Indicate whether the job needs parameters or not
	//
	//Return:
	// true if required (parameter will be pre-validated and 'SetParams' will be called)
	// false if no parameters needed (no check and 'SetParams' will not be called)
	ParamsRequired() bool

	//Run the business logic here.
	Run() error
}
