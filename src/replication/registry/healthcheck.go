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

package registry

import (
	"math/rand"
	"time"

	"github.com/goharbor/harbor/src/common/utils/log"
)

// MinInterval defines the minimum interval to check registries' health status.
const MinInterval = time.Minute * 5

// HealthChecker is used to regularly check all registries' health status and update
// check result to database
type HealthChecker struct {
	interval time.Duration
	closing  chan struct{}
	done     chan struct{}
	manager  Manager
}

// NewHealthChecker creates a new health checker
// - interval specifies the time interval to perform health check for registries
// - closing is a channel to stop the health checker
func NewHealthChecker(interval time.Duration, closing, done chan struct{}) *HealthChecker {
	return &HealthChecker{
		interval: interval,
		manager:  NewDefaultManager(),
		closing:  closing,
		done:     done,
	}
}

// Run performs health check for all registries regularly
func (c *HealthChecker) Run() {
	interval := c.interval
	if c.interval < MinInterval {
		interval = MinInterval
	}

	// Wait some random time before starting health checking. If Harbor is deployed in HA mode
	// with multiple instances, this will avoid instances check health in the same time.
	<-time.After(time.Duration(rand.Int63n(int64(interval))))

	ticker := time.NewTicker(interval)
	log.Infof("Start regular health check for registries with interval %v", interval)
	for {
		select {
		case <-ticker.C:
			if err := c.manager.HealthCheck(); err != nil {
				log.Errorf("Health check error: %v", err)
				continue
			}
			log.Debug("Health Check succeeded")
		case <-c.closing:
			log.Info("Stop health checker")
			// No cleanup works to do, signal done directly
			close(c.done)
			return
		}
	}
}
