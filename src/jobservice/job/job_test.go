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
	"fmt"
	"os"
	"strconv"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/vmware/harbor/src/common"
	"github.com/vmware/harbor/src/common/dao"
	"github.com/vmware/harbor/src/common/models"
	"github.com/vmware/harbor/src/common/utils/log"
	"github.com/vmware/harbor/src/common/utils/test"
	"github.com/vmware/harbor/src/jobservice/config"
)

var repJobID, scanJobID int64

func TestMain(m *testing.M) {
	//Init config...
	conf := test.GetDefaultConfigMap()
	if len(os.Getenv("MYSQL_HOST")) > 0 {
		conf[common.MySQLHost] = os.Getenv("MYSQL_HOST")
	}
	if len(os.Getenv("MYSQL_PORT")) > 0 {
		p, err := strconv.Atoi(os.Getenv("MYSQL_PORT"))
		if err != nil {
			panic(err)
		}
		conf[common.MySQLPort] = p
	}
	if len(os.Getenv("MYSQL_USR")) > 0 {
		conf[common.MySQLUsername] = os.Getenv("MYSQL_USR")
	}
	if len(os.Getenv("MYSQL_PWD")) > 0 {
		conf[common.MySQLPassword] = os.Getenv("MYSQL_PWD")
	}

	server, err := test.NewAdminserver(conf)
	if err != nil {
		log.Fatalf("failed to create a mock admin server: %v", err)
	}
	defer server.Close()
	if err := os.Setenv("ADMINSERVER_URL", server.URL); err != nil {
		log.Fatalf("failed to set env %s: %v", "ADMINSERVER_URL", err)
	}
	secretKeyPath := "/tmp/secretkey"
	_, err = test.GenerateKey(secretKeyPath)
	if err != nil {
		log.Fatalf("failed to generate secret key: %v", err)
	}
	defer os.Remove(secretKeyPath)
	if err := os.Setenv("KEY_PATH", secretKeyPath); err != nil {
		log.Fatalf("failed to set env %s: %v", "KEY_PATH", err)
	}
	if err := config.Init(); err != nil {
		log.Fatalf("failed to initialize configurations: %v", err)
	}
	dbSetting, err := config.Database()
	if err != nil {
		log.Fatalf("failed to get db configurations: %v", err)
	}
	if err := dao.InitDatabase(dbSetting); err != nil {
		log.Fatalf("failed to initialised databse, error: %v", err)
	}
	//prepare data
	if err := prepareRepJobData(); err != nil {
		log.Fatalf("failed to initialised databse, error: %v", err)
	}
	if err := prepareScanJobData(); err != nil {
		log.Fatalf("failed to initialised databse, error: %v", err)
	}
	rc := m.Run()
	clearRepJobData()
	clearScanJobData()
	if rc != 0 {
		os.Exit(rc)
	}
}

func TestRepJob(t *testing.T) {
	rj := NewRepJob(repJobID)
	assert := assert.New(t)
	err := rj.Init()
	assert.Nil(err)
	assert.Equal(repJobID, rj.ID())
	assert.Equal(ReplicationType, rj.Type())
	p := fmt.Sprintf("/var/log/jobs/job_%d.log", repJobID)
	assert.Equal(p, rj.LogPath())
	err = rj.UpdateStatus(models.JobRetrying)
	assert.Nil(err)
	j, err := dao.GetRepJob(repJobID)
	assert.Equal(models.JobRetrying, j.Status)
	assert.False(rj.parm.Insecure)
	rj2 := NewRepJob(99999)
	err = rj2.Init()
	assert.NotNil(err)
}

func TestScanJob(t *testing.T) {
	sj := NewScanJob(scanJobID)
	assert := assert.New(t)
	err := sj.Init()
	assert.Nil(err)
	assert.Equal(scanJobID, sj.ID())
	assert.Equal(ScanType, sj.Type())
	p := fmt.Sprintf("/var/log/jobs/scan_job/job_%d.log", scanJobID)
	assert.Equal(p, sj.LogPath())
	err = sj.UpdateStatus(models.JobRetrying)
	assert.Nil(err)
	j, err := dao.GetScanJob(scanJobID)
	assert.Equal(models.JobRetrying, j.Status)
	assert.Equal("sha256:0204dc6e09fa57ab99ac40e415eb637d62c8b2571ecbbc9ca0eb5e2ad2b5c56f", sj.parm.Digest)
	sj2 := NewScanJob(99999)
	err = sj2.Init()
	assert.NotNil(err)
}

func TestStatusUpdater(t *testing.T) {
	assert := assert.New(t)
	rj := NewRepJob(repJobID)
	su := &StatusUpdater{rj, models.JobFinished}
	su.Enter()
	su.Exit()
	j, err := dao.GetRepJob(repJobID)
	assert.Nil(err)
	assert.Equal(models.JobFinished, j.Status)
}

func prepareRepJobData() error {
	if err := clearRepJobData(); err != nil {
		return err
	}
	regURL, err := config.LocalRegURL()
	if err != nil {
		return err
	}
	target := models.RepTarget{
		Name:     "name",
		URL:      regURL,
		Username: "username",
		Password: "password",
	}

	targetID, err := dao.AddRepTarget(target)
	if err != nil {
		return err
	}
	policy := models.RepPolicy{
		ProjectID:   1,
		TargetID:    targetID,
		Description: "whatever",
		Name:        "mypolicy",
	}
	policyID, err := dao.AddRepPolicy(policy)
	if err != nil {
		return err
	}
	job := models.RepJob{
		Repository: "library/ubuntu",
		PolicyID:   policyID,
		Operation:  "transfer",
		TagList:    []string{"12.01", "14.04", "latest"},
	}
	id, err := dao.AddRepJob(job)
	if err != nil {
		return err
	}
	repJobID = id
	return nil
}

func clearRepJobData() error {
	if err := dao.ClearTable(models.RepJobTable); err != nil {
		return err
	}
	if err := dao.ClearTable(models.RepPolicyTable); err != nil {
		return err
	}
	return dao.ClearTable(models.RepTargetTable)
}

func prepareScanJobData() error {
	if err := clearScanJobData(); err != nil {
		return err
	}
	sj := models.ScanJob{
		Status:     models.JobPending,
		Repository: "library/ubuntu",
		Tag:        "15.10",
		Digest:     "sha256:0204dc6e09fa57ab99ac40e415eb637d62c8b2571ecbbc9ca0eb5e2ad2b5c56f",
	}
	id, err := dao.AddScanJob(sj)
	if err != nil {
		return err
	}
	scanJobID = id
	return nil
}

func clearScanJobData() error {
	if err := dao.ClearTable(models.ScanJobTable); err != nil {
		return err
	}
	return dao.ClearTable(models.ScanOverviewTable)
}
