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

package audit

import (
	"context"
	"io"
	"log/syslog"
	"os"

	"github.com/goharbor/harbor/src/lib/config"
	"github.com/goharbor/harbor/src/lib/log"
)

// LogMgr manage the audit log forward operations
var LogMgr = &LoggerManager{}

// LoggerManager manage the operations related to the audit log
type LoggerManager struct {
	endpoint     string
	initialized  bool
	remoteLogger *log.Logger
}

// Init redirect the audit log to the forward endpoint
func (a *LoggerManager) Init(_ context.Context, logEndpoint string) {
	var w io.Writer
	w, err := syslog.Dial("tcp", logEndpoint,
		syslog.LOG_INFO, "audit")
	a.initialized = true
	if err != nil {
		if len(logEndpoint) > 0 {
			log.Errorf("failed to create audit log, error %v", err)
		}
		w = os.Stdout
		a.initialized = false
	}
	a.remoteLogger = log.New(w, log.NewTextFormatter(), log.InfoLevel, 3)
	a.remoteLogger.SetFallback(log.DefaultLogger())
}

// DefaultLogger ...
func (a *LoggerManager) DefaultLogger(ctx context.Context) *log.Logger {
	endpoint := config.AuditLogForwardEndpoint(ctx)
	if a.endpoint != endpoint {
		a.Init(ctx, endpoint)
		a.initialized = true
	}
	return a.remoteLogger
}

// CheckEndpointActive check the liveliness of the endpoint
func CheckEndpointActive(address string) bool {
	al, err := syslog.Dial("tcp", address,
		syslog.LOG_INFO, "audit")
	if al != nil {
		defer al.Close()
	}
	if err != nil {
		log.Errorf("failed to connect to audit log endpoint, error %v", err)
		return false
	}
	return true
}
