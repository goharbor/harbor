// Copyright 2018 The Harbor Authors. All rights reserved.

package job

import (
	"github.com/gocraft/work"
	"github.com/vmware/harbor/src/jobservice_v2/env"
)

//RedisJob is a job wrapper to wrap the job.Interface to the style which can be recognized by the redis pool.
type RedisJob struct {
	job     interface{}
	context *env.Context
}

//NewRedisJob is constructor of RedisJob
func NewRedisJob(j interface{}, ctx *env.Context) *RedisJob {
	return &RedisJob{j, ctx}
}

//Run the job
func (rj *RedisJob) Run(j *work.Job) error {
	//Build job execution context
	jData := env.JobData{
		ID:   j.ID,
		Name: j.Name,
		Args: j.Args,
	}
	execContext, err := rj.context.JobContext.Build(jData)
	if err != nil {
		return err
	}

	//Inject data
	runningJob := Wrap(rj.job)
	//TODO: Update job status to 'Running'
	//TODO: Check function should be defined
	err = runningJob.Run(execContext, j.Args, nil)

	//TODO:
	//If error is stopped error, update status to 'Stopped' and return nil
	//If error is cancelled error, update status to 'Cancelled' and return err
	//Need to consider how to rm the retry option

	return err
}
