package job

import (
	"github.com/vmware/harbor/models"
	"github.com/vmware/harbor/utils/log"
)

func Schedule(job models.JobEntry) {
	log.Infof("job: %d will be scheduled", job.ID)
	//TODO: add support for cron string when needed.
	go run(job)
}
