package policy

import (
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/vmware/harbor/src/common/scheduler/task"
	"github.com/vmware/harbor/src/common/utils/log"
)

const (
	oneDay = 24 * 3600
)

//AlternatePolicyConfiguration store the related configurations for alternate policy.
type AlternatePolicyConfiguration struct {
	//Duration is the interval of executing attached tasks.
	//E.g: 24*3600 for daily
	//     7*24*3600 for weekly
	Duration time.Duration

	//An integer to indicate the the weekday of the week. Please be noted that Sunday is 7.
	//Use default value 0 to indicate weekday is not set.
	//To support by weekly function.
	Weekday int8

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

	//Unique name of this policy to support multiple instances
	name string
}

//NewAlternatePolicy is constructor of creating AlternatePolicy.
//Accept name and configuration as parameters.
func NewAlternatePolicy(name string, config *AlternatePolicyConfiguration) *AlternatePolicy {
	return &AlternatePolicy{
		RWMutex:    new(sync.RWMutex),
		tasks:      task.NewDefaultStore(),
		config:     config,
		isEnabled:  false,
		terminator: make(chan bool),
		name:       name,
	}
}

//GetConfig returns the current configuration options of this policy.
func (alp *AlternatePolicy) GetConfig() *AlternatePolicyConfiguration {
	return alp.config
}

//Name is an implementation of same method in policy interface.
func (alp *AlternatePolicy) Name() string {
	return alp.name
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

	//Check if configuration is valid
	if !alp.isValidConfig() {
		return nil, errors.New("Policy configuration is not valid")
	}

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
		var (
			waitingTime int64
		)
		timeNow := time.Now().UTC()

		//Reach the execution time point?
		//Weekday is set
		if alp.config.Weekday > 0 {
			targetWeekday := (alp.config.Weekday + 7) % 7
			currentWeekday := timeNow.Weekday()
			weekdayDiff := (int)(targetWeekday - (int8)(currentWeekday))
			if weekdayDiff < 0 {
				weekdayDiff += 7
			}
			waitingTime = (int64)(weekdayDiff * oneDay)
		}

		//Time
		utcTime := (int64)(timeNow.Hour()*3600 + timeNow.Minute()*60)
		diff := alp.config.OffsetTime - utcTime
		if waitingTime > 0 {
			waitingTime += diff
		} else {
			waitingTime = diff
			if waitingTime < 0 {
				waitingTime += oneDay
			}
		}

		//Let's wait for a while
		if waitingTime > 0 {
			//Wait for a while.
			log.Infof("Waiting for %d seconds after comparing offset %d and utc time %d\n", diff, alp.config.OffsetTime, utcTime)
			select {
			case <-time.After(time.Duration(waitingTime) * time.Second):
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

	return cfg == nil ||
		(cfg.Duration == cfg2.Duration &&
			cfg.OffsetTime == cfg2.OffsetTime &&
			cfg.Weekday == cfg2.Weekday)
}

//IsEnabled is an implementation of same method in policy interface.
func (alp *AlternatePolicy) IsEnabled() bool {
	defer alp.RUnlock()
	alp.RLock()

	return alp.isEnabled
}

//Check if the config is valid. At least it should have the configurations for supporting daily policy.
func (alp *AlternatePolicy) isValidConfig() bool {
	return alp.config != nil && alp.config.Duration > 0 && alp.config.OffsetTime >= 0
}
