package preheat

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/goharbor/harbor/src/lib/log"
	"github.com/goharbor/harbor/src/pkg/p2p/preheat/history"
	"github.com/goharbor/harbor/src/pkg/p2p/preheat/instance"
	"github.com/goharbor/harbor/src/pkg/p2p/preheat/models"
	"github.com/goharbor/harbor/src/pkg/p2p/preheat/provider"
)

const (
	healthCheckLoopInterval = 30 * time.Second
	progressUpdateInterval  = 5 * time.Second
	qSize                   = 1024
)

type progressItem struct {
	instanceID int64
	taskID     string
}

// Monitor the instance health and distribution status.
// Update the related status flag if needed.
type Monitor struct {
	// Cancellable context
	context context.Context

	// For history
	hManager history.Manager

	// For instance
	iManager instance.Manager

	// Queue for history updating
	q chan *progressItem
}

// NewMonitor is constructor of Monitor
func NewMonitor(ctx context.Context, iManager instance.Manager, hManager history.Manager) *Monitor {
	return &Monitor{
		context:  ctx,
		hManager: hManager,
		iManager: iManager,
		q:        make(chan *progressItem, qSize),
	}
}

// Start the loops
func (m *Monitor) Start() {
	// Start instance health check loop
	go func() {
		defer func() {
			log.Info("Monitor health check loop exit")
		}()

		tk := time.NewTicker(healthCheckLoopInterval)
		defer tk.Stop()

		for {
			select {
			case <-tk.C:
				m.healthLoop()
			case <-m.context.Done():
				return
			}
		}
	}()

	log.Info("Health check loop for instances start")

	// Start progress update loop
	go func() {
		defer func() {
			log.Info("Monitor progress check loop exit")
		}()

		for {
			select {
			case item := <-m.q:
				go func() {
					if done, err := m.checkTaskProgress(item.instanceID, item.taskID); err != nil {
						log.Errorf("Update progress error: %s", err)
					} else {
						log.Debugf("Check preheating progress of task %s to instance %d: done=%v", item.taskID, item.instanceID, done)
						if !done {
							// Keep on checking
							// put back
							// non blocking
							go func() {
								<-time.After(progressUpdateInterval)
								m.q <- item
							}()
						}
					}
				}()
			case <-m.context.Done():
				return
			}
		}
	}()

	log.Info("Task progress auto updater start")
}

// WatchProgress watches the preheating task progress
// non blocking
func (m *Monitor) WatchProgress(instanceID int64, taskID string) {
	go func() {
		m.q <- &progressItem{
			instanceID: instanceID,
			taskID:     taskID,
		}
	}()
}

func (m *Monitor) healthLoop() {
	_, all, err := m.iManager.List(nil)
	if err != nil {
		log.Errorf("health loop error: %s", err)
		return
	}

	for _, inst := range all {
		go func(inst *models.Metadata) {
			if err := m.checkInstanceHealth(inst); err != nil {
				log.Errorf("check instance health error: %s", err)
			} else {
				log.Debugf("check health of instance %s succeed", inst.Name)
			}
		}(inst)
	}
}

func (m *Monitor) checkTaskProgress(instID int64, taskID string) (bool, error) {
	meta, err := m.iManager.Get(instID)
	if err != nil {
		return false, err
	}

	p, err := getProvider(meta)
	if err != nil {
		return false, err
	}

	pStatus, err := p.CheckProgress(taskID)
	if err != nil {
		return false, err
	}

	trackStatus := models.TrackStatus(pStatus.Status)
	// Update history record
	if err := m.hManager.UpdateStatus(taskID, trackStatus, pStatus.StartTime, pStatus.FinishTime); err != nil {
		return false, err
	}

	done := trackStatus.Success() || trackStatus.Fail()

	return done, nil
}

func (m *Monitor) checkInstanceHealth(inst *models.Metadata) error {
	p, err := getProvider(inst)
	if err != nil {
		return err
	}

	// Retrieve the checking instance
	meta, err := m.iManager.Get(inst.ID)
	if err != nil {
		return err
	}

	status, err := p.GetHealth()
	if err != nil {
		// Set status to unhealthy
		meta.Status = provider.DriverStatusUnHealthy
	} else {
		meta.Status = status.Status
	}

	log.Debugf("Check health of instance %d: %s", inst.ID, meta.Status)

	return m.iManager.Update(meta)
}

func getProvider(inst *models.Metadata) (provider.Driver, error) {
	if inst == nil {
		return nil, errors.New("nil instance")
	}

	factory, ok := provider.GetProvider(inst.Provider)
	if !ok {
		return nil, fmt.Errorf("no provider with ID %s existing", inst.Provider)
	}

	p, err := factory(inst)
	if err != nil {
		return nil, err
	}

	return p, nil
}
