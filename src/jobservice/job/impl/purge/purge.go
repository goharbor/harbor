// Copyright Project Harbor Authors
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

package purge

import (
	"os"
	"strings"

	"github.com/goharbor/harbor/src/common"
	"github.com/goharbor/harbor/src/jobservice/job"
	"github.com/goharbor/harbor/src/pkg/audit"
	"github.com/goharbor/harbor/src/pkg/auditext"
)

// Job defines the purge job
type Job struct {
	retentionHour     int
	includeEventTypes []string
	dryRun            bool
	auditMgr          audit.Manager
	auditExtMgr       auditext.Manager
}

// MaxFails is implementation of same method in Interface.
func (j *Job) MaxFails() uint {
	return 1
}

// MaxCurrency is implementation of same method in Interface.
func (j *Job) MaxCurrency() uint {
	return 1
}

// ShouldRetry ...
func (j *Job) ShouldRetry() bool {
	return false
}

// Validate is implementation of same method in Interface.
func (j *Job) Validate(_ job.Parameters) error {
	return nil
}

func (j *Job) parseParams(params job.Parameters) {
	if len(params) == 0 {
		return
	}
	retHr, exist := params[common.PurgeAuditRetentionHour]
	if !exist {
		return
	}
	if rh, ok := retHr.(int); ok {
		j.retentionHour = rh
	} else if rh, ok := retHr.(float64); ok {
		j.retentionHour = int(rh)
	}

	dryRun, exist := params[common.PurgeAuditDryRun]
	if exist {
		if dryRun, ok := dryRun.(bool); ok && dryRun {
			j.dryRun = dryRun
		}
	}

	j.includeEventTypes = []string{}
	operations, exist := params[common.PurgeAuditIncludeEventTypes]
	if exist {
		if includeOps, ok := operations.(string); ok {
			if len(includeOps) > 0 {
				j.includeEventTypes = strings.Split(includeOps, ",")
			}
		}
	}
	// UT will use the mock mgr
	if os.Getenv("UTTEST") != "true" {
		j.auditMgr = audit.Mgr
		j.auditExtMgr = auditext.Mgr
	}
}

// Run the purge logic here.
func (j *Job) Run(ctx job.Context, params job.Parameters) error {
	logger := ctx.GetLogger()
	logger.Info("Purge audit job start")
	logger.Infof("job parameters %+v", params)
	if j.shouldStop(ctx) {
		logger.Info("received the stop signal, stop the purge job")
		return nil
	}
	j.parseParams(params)
	ormCtx := ctx.SystemContext()
	if j.retentionHour == -1 || j.retentionHour == 0 {
		logger.Infof("quit purge job, retentionHour:%v ", j.retentionHour)
		return nil
	}
	// cap the retentionHour
	if j.retentionHour > common.MaxAuditRetentionHour {
		j.retentionHour = common.MaxAuditRetentionHour
	}
	n, err := j.auditMgr.Purge(ormCtx, j.retentionHour, j.includeEventTypes, j.dryRun)
	if err != nil {
		logger.Errorf("failed to purge audit log, error: %v", err)
		return err
	}
	n2, err2 := j.auditExtMgr.Purge(ormCtx, j.retentionHour, j.includeEventTypes, j.dryRun)
	if err2 != nil {
		logger.Errorf("failed to purge audit log ext, error: %v", err2)
		return err2
	}
	logger.Infof("Purge operation parameter, retention_hour:%v, include_event_types:%v, dry_run:%v",
		j.retentionHour, j.includeEventTypes, j.dryRun)
	if j.dryRun {
		logger.Infof("[DRYRUN]Purged %d rows of audit logs", n)
		logger.Infof("[DRYRUN]Purged %d rows of audit log exts", n2)
	} else {
		logger.Infof("Purged %d rows of audit logs", n)
		logger.Infof("Purged %d rows of audit log exts", n2)
	}

	// Successfully exit
	return nil
}

func (j *Job) shouldStop(ctx job.Context) bool {
	opCmd, exit := ctx.OPCommand()
	if exit && opCmd.IsStop() {
		return true
	}
	return false
}
