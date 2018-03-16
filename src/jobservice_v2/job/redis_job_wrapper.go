// Copyright 2018 The Harbor Authors. All rights reserved.

package job

import (
	"github.com/gocraft/work"
	"github.com/vmware/harbor/src/jobservice_v2/env"
	"github.com/vmware/harbor/src/jobservice_v2/errs"
)

//StatusChangeCallback is the func called when job status changed
type StatusChangeCallback func(jobID string, status string)

//CheckOPCmdFuncFactoryFunc is used to generate CheckOPCmdFunc func for the specified job
type CheckOPCmdFuncFactoryFunc func(jobID string) CheckOPCmdFunc

//RedisJob is a job wrapper to wrap the job.Interface to the style which can be recognized by the redis pool.
type RedisJob struct {
	job              interface{}
	context          *env.Context
	callback         StatusChangeCallback
	opCmdFuncFactory CheckOPCmdFuncFactoryFunc
}

//NewRedisJob is constructor of RedisJob
func NewRedisJob(j interface{}, ctx *env.Context, statusChangeCallback StatusChangeCallback, opCmdFuncFactory CheckOPCmdFuncFactoryFunc) *RedisJob {
	return &RedisJob{j, ctx, statusChangeCallback, opCmdFuncFactory}
}

//Run the job
func (rj *RedisJob) Run(j *work.Job) error {
	//Build job execution context
	jData := env.JobData{
		ID:        j.ID,
		Name:      j.Name,
		Args:      j.Args,
		ExtraData: make(map[string]interface{}),
	}
	jData.ExtraData["opCommandFunc"] = rj.opCmdFuncFactory(j.ID)
	execContext, err := rj.context.JobContext.Build(jData)
	if err != nil {
		return err
	}

	//Inject data
	runningJob := Wrap(rj.job)
	//Start to run
	rj.callback(j.ID, JobStatusRunning)

	//TODO: Check function should be defined
	err = runningJob.Run(execContext, j.Args)

	if err == nil {
		rj.callback(j.ID, JobStatusSuccess)
		return nil
	}

	if errs.IsJobStoppedError(err) {
		rj.callback(j.ID, JobStatusStopped)
		return nil // no need to put it into the dead queue for resume
	}

	if errs.IsJobCancelledError(err) {
		rj.callback(j.ID, JobStatusCancelled)
		return err //need to resume
	}

	rj.callback(j.ID, JobStatusError)
	return err

	//TODO:
	//Need to consider how to rm the retry option
}
