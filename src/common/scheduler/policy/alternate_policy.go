package policy

import (
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/vmware/harbor/src/common/scheduler/task"
	"github.com/vmware/harbor/src/common/utils/log"
)

//AlternatePolicyConfiguration store the related configurations for alternate policy.
type AlternatePolicyConfiguration struct {
	//Duration is the interval of executing attached tasks.
	Duration time.Duration

	//OffsetTime is the execution time point of each turn
	//It's a number to indicate the seconds offset to the 00:00 of UTC time.
	OffsetTime int64
}

//AlternatePolicy is a policy that repeatedly executing tasks with specified duration during a specified time scope.
type AlternatePolicy struct {
	//To sync the related operations.
	*sync.RWMutex

	//Keep the attached tasks.
	tasks task.Store

	//Policy configurations.
	config *AlternatePolicyConfiguration

	//To indicated whether policy is enabled or not.
	isEnabled bool

	//Channel used to send evaluation result signals.
	evaluation chan bool

	//Channel used to notify policy termination.
	done chan bool

	//Channel used to receive terminate signal.
	terminator chan bool
}

//NewAlternatePolicy is constructor of creating AlternatePolicy.
func NewAlternatePolicy(config *AlternatePolicyConfiguration) *AlternatePolicy {
	return &AlternatePolicy{
		RWMutex:    new(sync.RWMutex),
		tasks:      task.NewDefaultStore(),
		config:     config,
		isEnabled:  false,
		terminator: make(chan bool),
	}
}

//GetConfig returns the current configuration options of this policy.
func (alp *AlternatePolicy) GetConfig() *AlternatePolicyConfiguration {
	return alp.config
}

//Name is an implementation of same method in policy interface.
func (alp *AlternatePolicy) Name() string {
	return "Alternate Policy"
}

//Tasks is an implementation of same method in policy interface.
func (alp *AlternatePolicy) Tasks() []task.Task {
	return alp.tasks.GetTasks()
}

//Done is an implementation of same method in policy interface.
func (alp *AlternatePolicy) Done() <-chan bool {
	return alp.done
}

//AttachTasks is an implementation of same method in policy interface.
func (alp *AlternatePolicy) AttachTasks(tasks ...task.Task) error {
	if len(tasks) == 0 {
		return errors.New("No tasks can be attached")
	}

	alp.tasks.AddTasks(tasks...)

	return nil
}

//Disable is an implementation of same method in policy interface.
func (alp *AlternatePolicy) Disable() error {
	alp.Lock()
	if !alp.isEnabled {
		alp.Unlock()
		return fmt.Errorf("Instance of policy %s is not enabled", alp.Name())
	}

	//Set state to disabled
	alp.isEnabled = false
	alp.Unlock()

	//Stop the evaluation goroutine
	alp.terminator <- true

	return nil
}

//Evaluate is an implementation of same method in policy interface.
func (alp *AlternatePolicy) Evaluate() (<-chan bool, error) {
	//Lock for state changing
	defer alp.Unlock()
	alp.Lock()

	//Check if policy instance is still running
	if alp.isEnabled {
		return nil, fmt.Errorf("Instance of policy %s is still running", alp.Name())
	}

	//Keep idempotent
	if alp.evaluation != nil {
		return alp.evaluation, nil
	}

	alp.done = make(chan bool)
	alp.evaluation = make(chan bool)

	go func() {
		timeNow := time.Now().UTC()

		//Reach the execution time point?
		utcTime := (int64)(timeNow.Hour()*3600 + timeNow.Minute()*60)
		diff := alp.config.OffsetTime - utcTime
		if diff < 0 {
			diff += 24 * 3600
		}
		if diff > 0 {
			//Wait for a while.
			log.Infof("Waiting for %d seconds after comparing offset %d and utc time %d\n", diff, alp.config.OffsetTime, utcTime)
			select {
			case <-time.After(time.Duration(diff) * time.Second):
			case <-alp.terminator:
				return
			}
		}

		//Trigger the first tick.
		alp.evaluation <- true

		//Start the ticker for repeat checking.
		tk := time.NewTicker(alp.config.Duration)
		defer func() {
			if tk != nil {
				tk.Stop()
			}
		}()

		for {
			select {
			case <-tk.C:
				if alp.IsEnabled() {
					alp.evaluation <- true
				}
			case <-alp.terminator:
				return
			}
		}
	}()

	//Enabled
	alp.isEnabled = true

	return alp.evaluation, nil
}

//Equal is an implementation of same method in policy interface.
func (alp *AlternatePolicy) Equal(p Policy) bool {
	if p == nil {
		return false
	}

	pl, ok := p.(*AlternatePolicy)
	if !ok {
		return false
	}

	cfg := pl.GetConfig()
	cfg2 := alp.GetConfig()
	if (cfg == nil && cfg2 != nil) || (cfg != nil && cfg2 == nil) {
		return false
	}

	return cfg == nil || (cfg.Duration == cfg2.Duration && cfg.OffsetTime == cfg2.OffsetTime)
}

//IsEnabled is an implementation of same method in policy interface.
func (alp *AlternatePolicy) IsEnabled() bool {
	defer alp.RUnlock()
	alp.RLock()

	return alp.isEnabled
}
