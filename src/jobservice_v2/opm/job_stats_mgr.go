// Copyright 2018 The Harbor Authors. All rights reserved.

package opm

import "github.com/vmware/harbor/src/jobservice_v2/models"

//JobStatsManager defines the methods to handle stats of job.
type JobStatsManager interface {
	//Start to serve
	Start()

	//Stop to serve
	Stop()

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
}
