package config

import (
	"fmt"
	"os"
	"strconv"

	"github.com/vmware/harbor/utils/log"
)

const defaultMaxWorkers int = 10

var maxJobWorkers int
var localRegURL string
var logDir string

func init() {
	maxWorkersEnv := os.Getenv("MAX_JOB_WORKERS")
	maxWorkers64, err := strconv.ParseInt(maxWorkersEnv, 10, 32)
	maxJobWorkers = int(maxWorkers64)
	if err != nil {
		log.Warningf("Failed to parse max works setting, error: %v, the default value: %d will be used", err, defaultMaxWorkers)
		maxJobWorkers = defaultMaxWorkers
	}

	localRegURL = os.Getenv("LOCAL_REGISTRY_URL")
	if len(localRegURL) == 0 {
		localRegURL = "http://registry:5000/"
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

	log.Debugf("config: maxJobWorkers: %d", maxJobWorkers)
	log.Debugf("config: localRegURL: %s", localRegURL)
	log.Debugf("config: logDir: %s", logDir)
}

func MaxJobWorkers() int {
	return maxJobWorkers
}

func LocalRegURL() string {
	return localRegURL
}

func LogDir() string {
	return logDir
}
