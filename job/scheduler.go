package job

import (
	"github.com/vmware/harbor/models"
	"github.com/vmware/harbor/utils/log"
	"os"
	"strconv"
)

var lock chan bool

const defaultMaxJobs int64 = 10

func init() {
	maxJobsEnv := os.Getenv("MAX_CONCURRENT_JOB")
	maxJobs, err := strconv.ParseInt(maxJobsEnv, 10, 32)
	if err != nil {
		log.Warningf("Failed to parse max job setting, error: %v, the default value: %d will be used", err, defaultMaxJobs)
		maxJobs = defaultMaxJobs
	}
	lock = make(chan bool, maxJobs)
}
func Schedule(job models.JobEntry) {
	log.Infof("job: %d will be scheduled", job.ID)
	//TODO: add support for cron string when needed.
	go func() {
		lock <- true
		defer func() { <-lock }()
		log.Infof("running job: %d", job.ID)
		run(job)
	}()
}
