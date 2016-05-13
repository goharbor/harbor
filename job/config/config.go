package config

import (
	"github.com/vmware/harbor/utils/log"
	"os"
	"strconv"
)

const defaultMaxWorkers int = 10

var maxJobWorkers int
var localRegURL string

func init() {
	maxWorkersEnv := os.Getenv("MAX_JOB_WORKERS")
	maxWorkers64, err := strconv.ParseInt(maxWorkersEnv, 10, 32)
	maxJobWorkers = int(maxWorkers64)
	if err != nil {
		log.Warningf("Failed to parse max works setting, error: %v, the default value: %d will be used", err, defaultMaxWorkers)
		maxJobWorkers = defaultMaxWorkers
	}

	localRegURL := os.Getenv("LOCAL_REGISTRY_URL")
	if len(localRegURL) == 0 {
		localRegURL = "http://registry:5000/"
	}

	log.Debugf("config: maxJobWorkers: %d", maxJobWorkers)
	log.Debugf("config: localRegURL: %s", localRegURL)
}

func MaxJobWorkers() int {
	return maxJobWorkers
}

func LocalRegURL() string {
	return localRegURL
}
