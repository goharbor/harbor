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

package logger

import (
	"context"
	"fmt"
	"reflect"
	"time"

	"github.com/goharbor/harbor/src/jobservice/logger/sweeper"
	"github.com/goharbor/harbor/src/lib/config"
)

const (
	oneDay = 24 * time.Hour
)

// SweeperController is an unified sweeper entry and built on top of multiple sweepers.
// It's responsible for starting the configured sweepers.
type SweeperController struct {
	context       context.Context
	sweepers      []sweeper.Interface
	errChan       chan error
	sweepInterval time.Duration
}

// NewSweeperController is constructor of controller.
func NewSweeperController(ctx context.Context, sweepers []sweeper.Interface) *SweeperController {
	return &SweeperController{
		context:       ctx,
		sweepers:      sweepers,
		errChan:       make(chan error, 1),
		sweepInterval: oneDay,
	}
}

// Sweep logs
func (c *SweeperController) Sweep() (int, error) {
	// Start to process errors
	go func() {
		for {
			select {
			case err := <-c.errChan:
				Error(err)
			case <-c.context.Done():
				return
			}
		}
	}()

	for _, s := range c.sweepers {
		go func(sw sweeper.Interface) {
			c.startSweeper(sw)
		}(s)
	}

	return 0, nil
}

// Duration = -1 for controller
func (c *SweeperController) Duration() int {
	return -1
}

func (c *SweeperController) startSweeper(s sweeper.Interface) {
	interval := c.sweepInterval
	if interval <= 0 {
		interval = oneDay
	}

	// Use the type name as a simple ID
	sid := reflect.TypeOf(s).String()
	defer Infof("sweeper %s exit", sid)

	// First run
	c.doSweeping(sid, s)

	// Loop
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			c.doSweeping(sid, s)
		case <-c.context.Done():
			return
		}
	}
}

func (c *SweeperController) doSweeping(sid string, s sweeper.Interface) {
	select {
	case <-c.context.Done():
		return
	default:
	}

	Debugf("Sweeper %s is under working", sid)

	if err := config.Load(context.Background()); err != nil {
		c.sendError(fmt.Errorf("failed to load configurations: %v", err))
		return
	}
	if config.ReadOnly(context.Background()) {
		c.sendError(fmt.Errorf("the system is in read only mode, cancel the sweeping"))
		return
	}

	count, err := s.Sweep()
	if err != nil {
		c.sendError(fmt.Errorf("sweep logs error in %s at %d: %s", sid, time.Now().Unix(), err))
		return
	}

	Infof("%d outdated log entries are sweepped by sweeper %s", count, sid)
}

func (c *SweeperController) sendError(err error) {
	select {
	case c.errChan <- err:
	case <-c.context.Done():
	}
}
