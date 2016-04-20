package job

import (
	"fmt"
	"github.com/vmware/harbor/models"
	"github.com/vmware/harbor/utils/log"
	"sync"
)

type JobRunner interface {
	Run(je models.JobEntry) error
}

var runners map[string]*JobRunner = make(map[string]*JobRunner)
var runnerLock = &sync.Mutex{}

func Register(jobType string, runner JobRunner) {
	runnerLock.Lock()
	defer runnerLock.Unlock()
	runners[jobType] = &runner
	log.Debugf("runnter for job type:%s has been registered", jobType)
}

func RunnerExists(jobType string) bool {
	_, ok := runners[jobType]
	return ok
}

func run(je models.JobEntry) error {
	runner, ok := runners[je.Type]
	if !ok {
		return fmt.Errorf("Runner for job type: %s does not exist")
	}
	(*runner).Run(je)
	return nil
}
