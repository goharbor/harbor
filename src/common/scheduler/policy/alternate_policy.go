package policy

import (
	"errors"
	"time"

	"github.com/vmware/harbor/src/common/scheduler/task"
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
	//Keep the attached tasks.
	tasks []task.Task

	//Policy configurations.
	config *AlternatePolicyConfiguration

	//Generate time ticks with specified duration.
	ticker *time.Ticker

	//To indicated whether policy is completed.
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
		tasks:      []task.Task{},
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
	copyList := []task.Task{}
	if alp.tasks != nil && len(alp.tasks) > 0 {
		copyList = append(copyList, alp.tasks...)
	}

	return copyList
}

//Done is an implementation of same method in policy interface.
func (alp *AlternatePolicy) Done() <-chan bool {
	return alp.done
}

//AttachTasks is an implementation of same method in policy interface.
func (alp *AlternatePolicy) AttachTasks(tasks ...task.Task) error {
	if tasks == nil || len(tasks) == 0 {
		return errors.New("No tasks can be attached")
	}

	alp.tasks = append(alp.tasks, tasks...)

	return nil
}

//Disable is an implementation of same method in policy interface.
func (alp *AlternatePolicy) Disable() error {
	//Stop the ticker
	if alp.ticker != nil {
		alp.ticker.Stop()
	}

	//Stop the evaluation goroutine
	alp.terminator <- true
	alp.ticker = nil

	return nil
}

//Evaluate is an implementation of same method in policy interface.
func (alp *AlternatePolicy) Evaluate() (<-chan bool, error) {
	//Keep idempotent
	if alp.isEnabled && alp.evaluation != nil {
		return alp.evaluation, nil
	}

	alp.done = make(chan bool)
	alp.evaluation = make(chan bool)

	go func() {
		defer func() {
			alp.isEnabled = false
		}()
		timeNow := time.Now().UTC()

		//Reach the execution time point?
		utcTime := (int64)(timeNow.Hour()*3600 + timeNow.Minute()*60)
		diff := alp.config.OffsetTime - utcTime
		if diff < 0 {
			diff += 24 * 3600
		}
		if diff > 0 {
			//Wait for a while.
			select {
			case <-time.After(time.Duration(diff) * time.Second):
			case <-alp.terminator:
				return
			}
		}

		//Trigger the first tick.
		alp.evaluation <- true

		//Start the ticker for repeat checking.
		alp.ticker = time.NewTicker(alp.config.Duration)
		for {
			select {
			case <-alp.ticker.C:
				alp.evaluation <- true
			case <-alp.terminator:
				return
			}
		}
	}()

	//Enabled
	alp.isEnabled = true

	return alp.evaluation, nil
}
