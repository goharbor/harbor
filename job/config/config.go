/*
   Copyright (c) 2016 VMware, Inc. All Rights Reserved.
   Licensed under the Apache License, Version 2.0 (the "License");
   you may not use this file except in compliance with the License.
   You may obtain a copy of the License at

       http://www.apache.org/licenses/LICENSE-2.0

   Unless required by applicable law or agreed to in writing, software
   distributed under the License is distributed on an "AS IS" BASIS,
   WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
   See the License for the specific language governing permissions and
   limitations under the License.
*/

package config

import (
	"fmt"
	"os"
	"strconv"

	"github.com/astaxie/beego"
	"github.com/vmware/harbor/utils/log"
)

const defaultMaxWorkers int = 10

var maxJobWorkers int
var localUIURL string
var localRegURL string
var logDir string
var uiSecret string
var secretKey string
var verifyRemoteCert string

func init() {
	maxWorkersEnv := os.Getenv("MAX_JOB_WORKERS")
	maxWorkers64, err := strconv.ParseInt(maxWorkersEnv, 10, 32)
	maxJobWorkers = int(maxWorkers64)
	if err != nil {
		log.Warningf("Failed to parse max works setting, error: %v, the default value: %d will be used", err, defaultMaxWorkers)
		maxJobWorkers = defaultMaxWorkers
	}

	localRegURL = os.Getenv("REGISTRY_URL")
	if len(localRegURL) == 0 {
		localRegURL = "http://registry:5000"
	}

	localUIURL = os.Getenv("UI_URL")
	if len(localUIURL) == 0 {
		localUIURL = "http://ui"
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

	verifyRemoteCert = os.Getenv("VERIFY_REMOTE_CERT")
	if len(verifyRemoteCert) == 0 {
		verifyRemoteCert = "on"
	}

	configPath := os.Getenv("CONFIG_PATH")
	if len(configPath) != 0 {
		log.Infof("Config path: %s", configPath)
		beego.LoadAppConfig("ini", configPath)
	}

	secretKey = os.Getenv("SECRET_KEY")
	if len(secretKey) != 16 {
		panic("The length of secretkey has to be 16 characters!")
	}

	log.Debugf("config: maxJobWorkers: %d", maxJobWorkers)
	log.Debugf("config: localUIURL: %s", localUIURL)
	log.Debugf("config: localRegURL: %s", localRegURL)
	log.Debugf("config: verifyRemoteCert: %s", verifyRemoteCert)
	log.Debugf("config: logDir: %s", logDir)
	log.Debugf("config: uiSecret: ******")
}

// MaxJobWorkers ...
func MaxJobWorkers() int {
	return maxJobWorkers
}

// LocalUIURL returns the local ui url, job service will use this URL to call API hosted on ui process
func LocalUIURL() string {
	return localUIURL
}

// LocalRegURL returns the local registry url, job service will use this URL to pull image from the registry
func LocalRegURL() string {
	return localRegURL
}

// LogDir returns the absolute path to which the log file will be written
func LogDir() string {
	return logDir
}

// UISecret will return the value of secret cookie for jobsevice to call UI API.
func UISecret() string {
	return uiSecret
}

// SecretKey will return the secret key for encryption/decryption password in target.
func SecretKey() string {
	return secretKey
}

// VerifyRemoteCert return the flag to tell jobservice whether or not verify the cert of remote registry
func VerifyRemoteCert() bool {
	return verifyRemoteCert != "off"
}
