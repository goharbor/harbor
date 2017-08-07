package scheduler

import (
	"github.com/vmware/harbor/src/common/scheduler/policy"
	"github.com/vmware/harbor/src/common/scheduler/task"
	"github.com/vmware/harbor/src/common/utils/log"

	"fmt"
	"sync"
)

//Watcher is an asynchronous runner to provide an evaluation environment for the policy.
type Watcher struct {
	//Locker to sync related operations.
	*sync.RWMutex

	//The target policy.
	p policy.Policy

	//The channel for receive stop signal.
	cmdChan chan bool

	//Indicate whether the watcher is started and running.
	isRunning bool

	//Report stats to scheduler.
	stats chan *StatItem

	//If policy is automatically completed, report the policy to scheduler.
	doneChan chan *Watcher
}

//NewWatcher is used as a constructor.
func NewWatcher(p policy.Policy, st chan *StatItem, done chan *Watcher) *Watcher {
	return &Watcher{
		RWMutex:   new(sync.RWMutex),
		p:         p,
		cmdChan:   make(chan bool),
		isRunning: false,
		stats:     st,
		doneChan:  done,
	}
}

//Start the running.
func (wc *Watcher) Start() {
	//Lock for state changing
	wc.Lock()
	defer wc.Unlock()

	if wc.isRunning {
		return
	}

	if wc.p == nil {
		return
	}

	go func(pl policy.Policy) {
		defer func() {
			if r := recover(); r != nil {
				log.Errorf("Runtime error in watcher:%s\n", r)
			}
		}()

		evalChan, err := pl.Evaluate()
		if err != nil {
			log.Errorf("Failed to evaluate ploicy %s with error: %s\n", pl.Name(), err.Error())
			return
		}
		done := pl.Done()

		for {
			select {
			case <-evalChan:
				{
					//If worker is not running, should not response any requests.
					if !wc.IsRunning() {
						continue
					}

					log.Infof("Receive evaluation signal from policy '%s'\n", pl.Name())
					//Start to run the attached tasks.
					for _, t := range pl.Tasks() {
						go func(tk task.Task) {
							defer func() {
								if r := recover(); r != nil {
									st := &StatItem{statTaskFail, 1, fmt.Errorf("Runtime error in task execution:%s", r)}
									if wc.stats != nil {
										wc.stats <- st
									}
								}
							}()
							err := tk.Run()

							//Report task execution stats.
							st := &StatItem{statTaskComplete, 1, err}
							if err != nil {
								st.Type = statTaskFail
							}
							if wc.stats != nil {
								wc.stats <- st
							}
						}(t)

						//Report task run stats.
						st := &StatItem{statTaskRun, 1, nil}
						if wc.stats != nil {
							wc.stats <- st
						}
					}
				}
			case <-done:
				{
					//Policy is automatically completed.
					//Report policy change stats.
					if wc.doneChan != nil {
						wc.doneChan <- wc
					}

					return
				}
			case <-wc.cmdChan:
				//Exit goroutine.
				return
			}
		}
	}(wc.p)

	wc.isRunning = true
}

//Stop the running.
func (wc *Watcher) Stop() {
	//Lock for state changing
	wc.Lock()
	if !wc.isRunning {
		wc.Unlock()
		return
	}

	wc.isRunning = false
	wc.Unlock()

	//Disable policy.
	if wc.p != nil {
		wc.p.Disable()
	}

	//Stop watcher.
	wc.cmdChan <- true

	log.Infof("Worker for policy %s is stopped.\n", wc.p.Name())
}

//IsRunning to indicate if the watcher is still running.
func (wc *Watcher) IsRunning() bool {
	wc.RLock()
	defer wc.RUnlock()

	return wc.isRunning
}
