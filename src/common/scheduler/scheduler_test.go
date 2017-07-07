package scheduler

import (
	"testing"
	"time"

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

type fakeTask struct {
	number int
}

func (ft *fakeTask) Name() string {
	return "for testing"
}

func (ft *fakeTask) Run() error {
	ft.number++
	return nil
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
	//Waiting for everything is stable
	time.Sleep(1 * time.Second)
	if DefaultScheduler.policies.Size() == 0 {
		t.Fatal("No policy in the store after calling Schedule()")
	}
	if DefaultScheduler.stats.PolicyCount != 1 {
		t.Fatal("Policy stats do not match")
	}

	time.Sleep(2 * time.Second)
	if fk.number == 100 {
		t.Fatal("Task is not triggered")
	}
	if DefaultScheduler.stats.Tasks == 0 {
		t.Fail()
	}
	if DefaultScheduler.stats.CompletedTasks == 0 {
		t.Fail()
	}

	if DefaultScheduler.UnSchedule(fp.Name()) != nil {
		t.Fatal("Unschedule policy failed")
	}
	//Waiting for everything is stable
	time.Sleep(1 * time.Second)

	if DefaultScheduler.stats.PolicyCount != 0 {
		t.Fatal("Policy count does not match after calling UnSchedule()")
	}
	copiedValue := DefaultScheduler.stats.CompletedTasks
	<-time.After(2 * time.Second)

	if copiedValue != DefaultScheduler.stats.CompletedTasks {
		t.Fatalf("Policy is still enabled after calling UnSchedule(),%d=%d", copiedValue, DefaultScheduler.stats.CompletedTasks)
	}

	DefaultScheduler.Stop()
	<-time.After(1 * time.Second)
	if DefaultScheduler.policies.Size() != 0 || DefaultScheduler.IsRunning() {
		t.Fatal("Scheduler is still running after stopping")
	}
}
