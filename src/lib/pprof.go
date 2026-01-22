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

package lib

import (
	"net/http"
	// import pprof
	_ "net/http/pprof"
	"os"
	"strings"

	"github.com/goharbor/harbor/src/lib/log"
)

// Start starts a pprof server in a goroutine if PPROF_ENABLED is "true" (case-insensitive).
// The address can be set via the PPROF_ADDR environment variable (defaults to :6060).
func StartPprof() {
	enabled := strings.ToLower(os.Getenv("PPROF_ENABLED"))
	if enabled != "true" {
		log.Debug("pprof not enabled; set PPROF_ENABLED=true to enable")
		return
	}
	go func() {
		addr := os.Getenv("PPROF_ADDR")
		if addr == "" {
			addr = ":6060"
		}
		log.Infof("Starting pprof at %s/debug/pprof/", addr)
		if err := http.ListenAndServe(addr, http.DefaultServeMux); err != nil {
			log.Errorf("pprof exited: %v", err)
		}
	}()
}
