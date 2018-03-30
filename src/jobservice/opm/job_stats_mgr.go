// Copyright 2018 The Harbor Authors. All rights reserved.

package opm

import "github.com/vmware/harbor/src/jobservice/models"

//JobStatsManager defines the methods to handle stats of job.
type JobStatsManager interface {
	//Start to serve
	Start()

	//Shutdown the manager
	Shutdown()

	//Save the job stats
	//Async method to retry and improve performance
	//
	//jobStats models.JobStats : the job stats to be saved
	Save(jobStats models.JobStats)

	//Get the job stats from backend store
	//Sync method as we need the data
	//
	//Returns:
	//  models.JobStats : job stats data
	//  error           : error if meet any problems
	Retrieve(jobID string) (models.JobStats, error)

	//SetJobStatus will mark the status of job to the specified one
	//Async method to retry
	SetJobStatus(jobID string, status string)

	//Send command fro the specified job
	//
	//jobID string   : ID of the being retried job
	//command string : the command applied to the job like stop/cancel
	//
	//Returns:
	//  error if it was not successfully sent
	SendCommand(jobID string, command string) error

	//CtlCommand checks if control command is fired for the specified job.
	//
	//jobID string : ID of the job
	//
	//Returns:
	//  the command if it was fired
	//  error if it was not fired yet to meet some other problems
	CtlCommand(jobID string) (string, error)

	//CheckIn message for the specified job like detailed progress info.
	//
	//jobID string   : ID of the job
	//message string : The message being checked in
	//
	CheckIn(jobID string, message string)

	//DieAt marks the failed jobs with the time they put into dead queue.
	//
	//jobID string   : ID of the job
	//message string : The message being checked in
	//
	DieAt(jobID string, dieAt int64)

	//RegisterHook is used to save the hook url or cache the url in memory.
	//
	//jobID string   : ID of job
	//hookURL string : the hook url being registered
	//isCached bool  :  to indicate if only cache the hook url
	//
	//Returns:
	//  error if meet any problems
	RegisterHook(jobID string, hookURL string, isCached bool) error

	//Mark the periodic job stats expired
	//
	//jobID string   : ID of job
	//
	//Returns:
	//  error if meet any problems
	ExpirePeriodicJobStats(jobID string) error
}
