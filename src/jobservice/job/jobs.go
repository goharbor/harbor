// Copyright (c) 2017 VMware, Inc. All Rights Reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//    http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package job

import (
	"github.com/vmware/harbor/src/common/dao"
	uti "github.com/vmware/harbor/src/common/utils"
	"github.com/vmware/harbor/src/jobservice/config"

	"fmt"
	"path/filepath"
)

// Type is for job Type
type Type int

const (
	// ReplicationType is the Type to identify a replication job.
	ReplicationType Type = iota
	// ScanType is the Type to identify a image scanning job.
	ScanType
)

func (t Type) String() string {
	if ReplicationType == t {
		return "Replication"
	} else if ScanType == t {
		return "Scan"
	} else {
		return "Unknown"
	}
}

//Job is abstraction for image replication and image scan jobs.
type Job interface {
	//ID returns the id of the job
	ID() int64
	Type() Type
	LogPath() string
	UpdateStatus(status string) error
	Init() error
	//Parm() interface{}
}

// RepJobParm wraps the parm of a replication job
type RepJobParm struct {
	LocalRegURL    string
	TargetURL      string
	TargetUsername string
	TargetPassword string
	Repository     string
	Tags           []string
	Enabled        int
	Operation      string
	Insecure       bool
}

// RepJob implements Job interface, represents a replication job.
type RepJob struct {
	id   int64
	parm *RepJobParm
}

// ID returns the ID of the replication job
func (rj *RepJob) ID() int64 {
	return rj.id
}

// Type returns the type of the replication job, it should always be ReplicationType
func (rj *RepJob) Type() Type {
	return ReplicationType
}

// LogPath returns the absolute path of the particular replication job.
func (rj *RepJob) LogPath() string {
	return GetJobLogPath(config.LogDir(), rj.id)
}

// UpdateStatus ...
func (rj *RepJob) UpdateStatus(status string) error {
	return dao.UpdateRepJobStatus(rj.id, status)
}

// String ...
func (rj *RepJob) String() string {
	return fmt.Sprintf("{JobID: %d, JobType: %v}", rj.ID(), rj.Type())
}

// Init prepares parm for the replication job
func (rj *RepJob) Init() error {
	//init parms
	job, err := dao.GetRepJob(rj.id)
	if err != nil {
		return fmt.Errorf("Failed to get job, error: %v", err)
	}
	if job == nil {
		return fmt.Errorf("The job doesn't exist in DB, job id: %d", rj.id)
	}
	policy, err := dao.GetRepPolicy(job.PolicyID)
	if err != nil {
		return fmt.Errorf("Failed to get policy, error: %v", err)
	}
	if policy == nil {
		return fmt.Errorf("The policy doesn't exist in DB, policy id:%d", job.PolicyID)
	}

	regURL, err := config.LocalRegURL()
	if err != nil {
		return err
	}
	verify, err := config.VerifyRemoteCert()
	if err != nil {
		return err
	}
	rj.parm = &RepJobParm{
		LocalRegURL: regURL,
		Repository:  job.Repository,
		Tags:        job.TagList,
		Enabled:     policy.Enabled,
		Operation:   job.Operation,
		Insecure:    !verify,
	}
	if policy.Enabled == 0 {
		//worker will cancel this job
		return nil
	}
	target, err := dao.GetRepTarget(policy.TargetID)
	if err != nil {
		return fmt.Errorf("Failed to get target, error: %v", err)
	}
	if target == nil {
		return fmt.Errorf("The target doesn't exist in DB, target id: %d", policy.TargetID)
	}
	rj.parm.TargetURL = target.URL
	rj.parm.TargetUsername = target.Username
	pwd := target.Password

	if len(pwd) != 0 {
		key, err := config.SecretKey()
		if err != nil {
			return err
		}
		pwd, err = uti.ReversibleDecrypt(pwd, key)
		if err != nil {
			return fmt.Errorf("failed to decrypt password: %v", err)
		}
	}

	rj.parm.TargetPassword = pwd
	return nil
}

// NewRepJob returns a pointer to RepJob which implements the Job interface.
// Given API only gets the id, it will call this func to get a instance that can be manuevered by state machine.
func NewRepJob(id int64) *RepJob {
	return &RepJob{id: id}
}

//ScanJob implements the Job interface, representing a job for scanning image.
type ScanJob struct {
	id   int64
	parm *ScanJobParm
}

//ScanJobParm wraps the parms of a image scan job.
type ScanJobParm struct {
	Repository string
	Tag        string
	Digest     string
}

//ID returns the id of the scan
func (sj *ScanJob) ID() int64 {
	return sj.id
}

//Type always return ScanType
func (sj *ScanJob) Type() Type {
	return ScanType
}

//LogPath returns the absolute path of the log file for the job, log files for scan job will be put in a sub folder of base log path.
func (sj *ScanJob) LogPath() string {
	return GetJobLogPath(filepath.Join(config.LogDir(), "scan_job"), sj.id)
}

//String ...
func (sj *ScanJob) String() string {
	return fmt.Sprintf("{JobID: %d, JobType: %v}", sj.ID(), sj.Type())
}

//UpdateStatus ...
func (sj *ScanJob) UpdateStatus(status string) error {
	return dao.UpdateScanJobStatus(sj.id, status)
}

//Init query the DB and populate the information of the image to scan in the parm of this job.
func (sj *ScanJob) Init() error {
	job, err := dao.GetScanJob(sj.id)
	if err != nil {
		return fmt.Errorf("Failed to get job, error: %v", err)
	}
	if job == nil {
		return fmt.Errorf("The job doesn't exist in DB, job id: %d", sj.id)
	}
	sj.parm = &ScanJobParm{
		Repository: job.Repository,
		Tag:        job.Tag,
		Digest:     job.Digest,
	}
	return nil
}

//NewScanJob creates a instance of ScanJob by id.
func NewScanJob(id int64) *ScanJob {
	return &ScanJob{id: id}
}
