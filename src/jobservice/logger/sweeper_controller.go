package logger

import (
	"context"
	"fmt"
	"reflect"
	"time"

	"github.com/goharbor/harbor/src/jobservice/logger/sweeper"
)

const (
	oneDay = 24 * time.Hour
)

// SweeperController is an unified sweeper entry and built on top of multiple sweepers.
// It's responsible for starting the configured sweepers.
type SweeperController struct {
	context  context.Context
	sweepers []sweeper.Interface
	errChan  chan error
}

// NewSweeperController is constructor of controller.
func NewSweeperController(ctx context.Context, sweepers []sweeper.Interface) *SweeperController {
	return &SweeperController{
		context:  ctx,
		sweepers: sweepers,
		errChan:  make(chan error, 1),
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
	d := s.Duration()
	if d <= 0 {
		d = 1
	}

	// Use the type name as a simple ID
	sid := reflect.TypeOf(s).String()
	defer Infof("sweeper %s exit", sid)

	// First run
	go c.doSweeping(sid, s)

	// Loop
	ticker := time.NewTicker(time.Duration(d) * oneDay)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			go c.doSweeping(sid, s)
		case <-c.context.Done():
			return
		}
	}
}

func (c *SweeperController) doSweeping(sid string, s sweeper.Interface) {
	Debugf("Sweeper %s is under working", sid)

	count, err := s.Sweep()
	if err != nil {
		c.errChan <- fmt.Errorf("sweep logs error in %s at %d: %s", sid, time.Now().Unix(), err)
		return
	}

	Infof("%d outdated log entries are swept by sweeper %s", count, sid)
}
