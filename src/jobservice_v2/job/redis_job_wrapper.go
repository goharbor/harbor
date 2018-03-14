// Copyright 2018 The Harbor Authors. All rights reserved.

package job

import (
	"reflect"

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
	//Inject data
	runningJob := rj.Wrap()
	runningJob.SetContext(rj.context.JobContext)
	if runningJob.ParamsRequired() {
		if err := runningJob.SetParams(j.Args); err != nil {
			return err
		}
	}

	//TODO: Update job status to 'Running'
	err := runningJob.Run()

	//TODO:
	//If error is stopped error, update status to 'Stopped' and return nil
	//If error is cancelled error, update status to 'Cancelled' and return err

	return err
}

//Wrap returns a new (job.)Interface based on the wrapped job handler reference.
func (rj *RedisJob) Wrap() Interface {
	theType := reflect.TypeOf(rj.job)

	if theType.Kind() == reflect.Ptr {
		theType = theType.Elem()
	}

	//Crate new
	v := reflect.New(theType).Elem()
	return v.Addr().Interface().(Interface)
}
