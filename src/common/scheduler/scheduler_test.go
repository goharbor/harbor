package scheduler

import (
	"sync/atomic"
	"testing"
	"time"

	"github.com/vmware/harbor/src/common/scheduler/policy"
	"github.com/vmware/harbor/src/common/scheduler/task"
)

type fakePolicy struct {
	tasks      []task.Task
	done       chan bool
	evaluation chan bool
	terminate  chan bool
	ticker     *time.Ticker
}

func (fp *fakePolicy) Name() string {
	return "testing policy"
}

func (fp *fakePolicy) Tasks() []task.Task {
	return fp.tasks
}

func (fp *fakePolicy) AttachTasks(tasks ...task.Task) error {
	fp.tasks = append(fp.tasks, tasks...)
	return nil
}

func (fp *fakePolicy) Done() <-chan bool {
	return fp.done
}

func (fp *fakePolicy) Evaluate() (<-chan bool, error) {
	fp.evaluation = make(chan bool, 1)
	fp.done = make(chan bool)
	fp.terminate = make(chan bool)

	fp.evaluation <- true
	go func() {
		fp.ticker = time.NewTicker(1 * time.Second)
		for {
			select {
			case <-fp.terminate:
				return
			case <-fp.ticker.C:
				fp.evaluation <- true
			}
		}
	}()
	return fp.evaluation, nil
}

func (fp *fakePolicy) Disable() error {
	if fp.ticker != nil {
		fp.ticker.Stop()
	}

	fp.terminate <- true
	return nil
}

func (fp *fakePolicy) Equal(policy.Policy) bool {
	return false
}

func (fp *fakePolicy) IsEnabled() bool {
	return true
}

type fakeTask struct {
	number int32
}

func (ft *fakeTask) Name() string {
	return "for testing"
}

func (ft *fakeTask) Run() error {
	atomic.AddInt32(&(ft.number), 1)
	return nil
}

func (ft *fakeTask) Number() int32 {
	return atomic.LoadInt32(&(ft.number))
}

//Wacher will be tested together with scheduler.
func TestScheduler(t *testing.T) {
	DefaultScheduler.Start()
	if DefaultScheduler.policies.Size() != 0 {
		t.Fail()
	}

	if DefaultScheduler.stats.PolicyCount != 0 {
		t.Fail()
	}

	if !DefaultScheduler.IsRunning() {
		t.Fatal("Scheduler is not started")
	}

	fp := &fakePolicy{
		tasks: []task.Task{},
	}
	fk := &fakeTask{number: 100}
	fp.AttachTasks(fk)

	if DefaultScheduler.Schedule(fp) != nil {
		t.Fatal("Schedule policy failed")
	}
	if DefaultScheduler.policies.Size() == 0 {
		t.Fatal("No policy in the store after calling Schedule()")
	}
	if DefaultScheduler.GetPolicy(fp.Name()) == nil {
		t.Fatal("Failed to get poicy by name")
	}

	time.Sleep(2 * time.Second)
	if fk.Number() == 100 {
		t.Fatal("Task is not triggered")
	}

	if DefaultScheduler.UnSchedule(fp.Name()) != nil {
		t.Fatal("Unschedule policy failed")
	}

	if DefaultScheduler.PolicyCount() != 0 {
		t.Fatal("Policy count does not match after calling UnSchedule()")
	}

	var copiedValue int32
	<-time.After(1 * time.Second)
	atomic.StoreInt32(&copiedValue, fk.Number())
	<-time.After(2 * time.Second)

	if atomic.LoadInt32(&copiedValue) != fk.Number() {
		t.Fatalf("Policy is still enabled after calling UnSchedule(),%d=%d", atomic.LoadInt32(&copiedValue), fk.Number())
	}

	DefaultScheduler.Stop()
	<-time.After(1 * time.Second)
	if DefaultScheduler.PolicyCount() != 0 || DefaultScheduler.IsRunning() {
		t.Fatal("Scheduler is still running after stopping")
	}
}
