package config

import (
	"fmt"
	"os"
	"strconv"

	"github.com/vmware/harbor/utils/log"
)

const defaultMaxWorkers int = 10

var maxJobWorkers int
var localURL string
var logDir string
var uiSecret string

func init() {
	maxWorkersEnv := os.Getenv("MAX_JOB_WORKERS")
	maxWorkers64, err := strconv.ParseInt(maxWorkersEnv, 10, 32)
	maxJobWorkers = int(maxWorkers64)
	if err != nil {
		log.Warningf("Failed to parse max works setting, error: %v, the default value: %d will be used", err, defaultMaxWorkers)
		maxJobWorkers = defaultMaxWorkers
	}

	localURL = os.Getenv("HARBOR_URL")
	if len(localURL) == 0 {
		localURL = "http://registry:5000/"
	}

	logDir = os.Getenv("LOG_DIR")
	if len(logDir) == 0 {
		logDir = "/var/log"
	}

	f, err := os.Open(logDir)
	defer f.Close()
	if err != nil {
		panic(err)
	}
	finfo, err := f.Stat()
	if err != nil {
		panic(err)
	}
	if !finfo.IsDir() {
		panic(fmt.Sprintf("%s is not a direcotry", logDir))
	}

	uiSecret = os.Getenv("UI_SECRET")
	if len(uiSecret) == 0 {
		panic("UI Secret is not set")
	}

	log.Debugf("config: maxJobWorkers: %d", maxJobWorkers)
	log.Debugf("config: localHarborURL: %s", localURL)
	log.Debugf("config: logDir: %s", logDir)
	log.Debugf("config: uiSecret: ******")
}

// MaxJobWorkers ...
func MaxJobWorkers() int {
	return maxJobWorkers
}

// LocalHarborURL returns the local registry url, job service will use this URL to pull manifest and repository.
func LocalHarborURL() string {
	return localURL
}

// LogDir returns the absolute path to which the log file will be written
func LogDir() string {
	return logDir
}

// UISecret will return the value of secret cookie for jobsevice to call UI API.
func UISecret() string {
	return uiSecret
}
