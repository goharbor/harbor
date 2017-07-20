package scheduler

import (
	"github.com/vmware/harbor/src/common/scheduler/policy"
	"github.com/vmware/harbor/src/common/utils/log"

	"errors"
	"fmt"
	"reflect"
	"strings"
	"time"
)

const (
	defaultQueueSize = 10

	statSchedulePolicy   = "Schedule Policy"
	statUnSchedulePolicy = "Unschedule Policy"
	statTaskRun          = "Task Run"
	statTaskComplete     = "Task Complete"
	statTaskFail         = "Task Fail"
)

//StatItem is defined for the stat metrics.
type StatItem struct {
	//Metrics catalog
	Type string

	//The stat value
	Value uint32

	//Attach some other info
	Attachment interface{}
}

//StatSummary is used to collect some metrics of scheduler.
type StatSummary struct {
	//Count of scheduled policy
	PolicyCount uint32

	//Total count of tasks
	Tasks uint32

	//Count of successfully complete tasks
	CompletedTasks uint32

	//Count of tasks with errors
	TasksWithError uint32
}

//Configuration defines configuration of Scheduler.
type Configuration struct {
	QueueSize uint8
}

//Scheduler is designed for scheduling policies.
type Scheduler struct {
	//Related configuration options for scheduler.
	config *Configuration

	//Store to keep the references of scheduled policies.
	policies Store

	//Queue for receiving policy unschedule request or complete signal.
	unscheduleQueue chan string

	//Channel for receiving stat metrics.
	statChan chan *StatItem

	//Channel for terminate scheduler damon.
	terminateChan chan bool

	//The stat metrics of scheduler.
	stats *StatSummary

	//To indicate whether scheduler is running or not
	isRunning bool
}

//DefaultScheduler is a default scheduler.
var DefaultScheduler = NewScheduler(nil)

//NewScheduler is constructor for creating a scheduler.
func NewScheduler(config *Configuration) *Scheduler {
	var qSize uint8 = defaultQueueSize
	if config != nil && config.QueueSize > 0 {
		qSize = config.QueueSize
	}

	usq := make(chan string, qSize)
	stChan := make(chan *StatItem, 4)
	tc := make(chan bool, 1)

	store := NewConcurrentStore()
	return &Scheduler{
		config:          config,
		policies:        store,
		unscheduleQueue: usq,
		statChan:        stChan,
		terminateChan:   tc,
		stats: &StatSummary{
			PolicyCount:    0,
			Tasks:          0,
			CompletedTasks: 0,
			TasksWithError: 0,
		},
		isRunning: false,
	}
}

//Start the scheduler damon.
func (sch *Scheduler) Start() {
	if sch.isRunning {
		return
	}
	go func() {
		defer func() {
			if r := recover(); r != nil {
				log.Errorf("Runtime error in scheduler:%s\n", r)
			}
		}()
		defer func() {
			//Exit and clear.
			sch.isRunning = false
			//Stop all watchers.
			for _, wt := range sch.policies.GetAll() {
				wt.Stop()
			}
			//Clear resources
			sch.policies.Clear()
			log.Infof("Policy scheduler stop at %s\n", time.Now().UTC().Format(time.RFC3339))
		}()
		for {
			select {
			case <-sch.terminateChan:
				//Exit
				return
			case name := <-sch.unscheduleQueue:
				//Unscheduled when policy is completed.
				if err := sch.UnSchedule(name); err != nil {
					log.Error(err.Error())
				}
			case stat := <-sch.statChan:
				{
					switch stat.Type {
					case statSchedulePolicy:
						sch.stats.PolicyCount += stat.Value
						break
					case statUnSchedulePolicy:
						sch.stats.PolicyCount -= stat.Value
						break
					case statTaskRun:
						sch.stats.Tasks += stat.Value
						break
					case statTaskComplete:
						sch.stats.CompletedTasks += stat.Value
						break
					case statTaskFail:
						sch.stats.TasksWithError += stat.Value
						break
					default:
						break
					}
					log.Infof("Policies:%d, Tasks:%d, CompletedTasks:%d, FailedTasks:%d\n",
						sch.stats.PolicyCount,
						sch.stats.Tasks,
						sch.stats.CompletedTasks,
						sch.stats.TasksWithError)

					if stat.Attachment != nil &&
						reflect.TypeOf(stat.Attachment).String() == "*errors.errorString" {
						log.Errorf("%s: %s\n", stat.Type, stat.Attachment.(error).Error())
					}
				}

			}
		}
	}()

	sch.isRunning = true
	log.Infof("Policy scheduler start at %s\n", time.Now().UTC().Format(time.RFC3339))
}

//Stop the scheduler damon.
func (sch *Scheduler) Stop() {
	if !sch.isRunning {
		return
	}

	//Terminate damon to stop receiving signals.
	sch.terminateChan <- true
}

//Schedule and enable the policy.
func (sch *Scheduler) Schedule(scheduledPolicy policy.Policy) error {
	if scheduledPolicy == nil {
		return errors.New("nil is not Policy object")
	}

	if strings.TrimSpace(scheduledPolicy.Name()) == "" {
		return errors.New("Policy should be assigned a name")
	}

	tasks := scheduledPolicy.Tasks()
	if tasks == nil || len(tasks) == 0 {
		return errors.New("Policy must attach task(s)")
	}

	if sch.policies.Exists(scheduledPolicy.Name()) {
		return fmt.Errorf("Duplicated policy: %s", scheduledPolicy.Name())
	}

	//Schedule the policy.
	watcher := NewWatcher(scheduledPolicy, sch.statChan, sch.unscheduleQueue)
	//Enable it.
	watcher.Start()

	//Keep the policy for future use after it's successfully scheduled.
	sch.policies.Put(scheduledPolicy.Name(), watcher)

	//Update stats and log info.
	log.Infof("Policy %s is scheduled", scheduledPolicy.Name())
	sch.statChan <- &StatItem{statSchedulePolicy, 1, nil}

	return nil
}

//UnSchedule the specified policy from the enabled policies list.
func (sch *Scheduler) UnSchedule(policyName string) error {
	if strings.TrimSpace(policyName) == "" {
		return errors.New("Empty policy name is invalid")
	}

	if !sch.policies.Exists(policyName) {
		return fmt.Errorf("Policy %s is not existing", policyName)
	}

	//Unschedule the policy.
	//Find the watcher.
	watcher := sch.policies.Remove(policyName)
	if watcher != nil && watcher.IsRunning() {
		watcher.Stop()

		//Update stats and log info.
		log.Infof("Policy %s is unscheduled", policyName)
		sch.statChan <- &StatItem{statUnSchedulePolicy, 1, nil}
	} else {
		log.Warningf("Inconsistent worker status for policy '%s'.\n", policyName)
	}

	return nil
}

//IsRunning to indicate whether the scheduler is running.
func (sch *Scheduler) IsRunning() bool {
	return sch.isRunning
}

//HasScheduled is to check whether the given policy has been scheduled or not.
func (sch *Scheduler) HasScheduled(policyName string) bool {
	return sch.policies.Exists(policyName)
}

//GetPolicy is used to get related policy reference by its name.
func (sch *Scheduler) GetPolicy(policyName string) policy.Policy {
	wk := sch.policies.Get(policyName)
	if wk != nil {
		return wk.p
	}

	return nil
}
