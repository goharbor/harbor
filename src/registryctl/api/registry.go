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

package api

import (
	"bytes"
	"net/http"
	"time"

	"os/exec"

	"github.com/goharbor/harbor/src/common/utils/log"
)

const (
	regConf = "/etc/registry/config.yml"
)

// GCResult ...
type GCResult struct {
	Status    bool      `json:"status"`
	Msg       string    `json:"msg"`
	StartTime time.Time `json:"starttime"`
	EndTime   time.Time `json:"endtime"`
}

// StartGC ...
func StartGC(w http.ResponseWriter, r *http.Request) {
	cmd := exec.Command("/bin/bash", "-c", "registry garbage-collect --delete-untagged=false "+regConf)
	var outBuf, errBuf bytes.Buffer
	cmd.Stdout = &outBuf
	cmd.Stderr = &errBuf

	start := time.Now()
	log.Debugf("Start to execute garbage collection...")
	if err := cmd.Run(); err != nil {
		log.Errorf("Fail to execute GC: %v, command err: %s", err, errBuf.String())
		handleInternalServerError(w)
		return
	}

	gcr := GCResult{true, outBuf.String(), start, time.Now()}
	if err := writeJSON(w, gcr); err != nil {
		log.Errorf("failed to write response: %v", err)
		return
	}
	log.Debugf("Successful to execute garbage collection...")
}
