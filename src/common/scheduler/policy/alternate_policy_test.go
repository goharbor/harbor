package policy

import (
	"sync/atomic"
	"testing"
	"time"
)

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
	return atomic.LoadInt32(&ft.number)
}

func TestBasic(t *testing.T) {
	tp := NewAlternatePolicy(&AlternatePolicyConfiguration{})
	err := tp.AttachTasks(&fakeTask{number: 100})
	if err != nil {
		t.Fail()
	}

	if tp.GetConfig() == nil {
		t.Fail()
	}

	if tp.Name() != "Alternate Policy" {
		t.Fail()
	}

	tks := tp.Tasks()
	if tks == nil || len(tks) != 1 {
		t.Fail()
	}

}

func TestEvaluatePolicy(t *testing.T) {
	now := time.Now().UTC()
	utcOffset := (int64)(now.Hour()*3600 + now.Minute()*60)
	tp := NewAlternatePolicy(&AlternatePolicyConfiguration{
		Duration:   1 * time.Second,
		OffsetTime: utcOffset + 1,
	})
	err := tp.AttachTasks(&fakeTask{number: 100})
	if err != nil {
		t.Fail()
	}
	ch, _ := tp.Evaluate()
	var counter int32

	for i := 0; i < 3; i++ {
		select {
		case <-ch:
			atomic.AddInt32(&counter, 1)
		case <-time.After(2 * time.Second):
			continue
		}
	}

	if atomic.LoadInt32(&counter) != 3 {
		t.Fail()
	}

	tp.Disable()
}

func TestDisablePolicy(t *testing.T) {
	now := time.Now().UTC()
	utcOffset := (int64)(now.Hour()*3600 + now.Minute()*60)
	tp := NewAlternatePolicy(&AlternatePolicyConfiguration{
		Duration:   1 * time.Second,
		OffsetTime: utcOffset + 1,
	})
	err := tp.AttachTasks(&fakeTask{number: 100})
	if err != nil {
		t.Fail()
	}
	ch, _ := tp.Evaluate()
	var counter int32
	terminate := make(chan bool)
	defer func() {
		terminate <- true
	}()
	go func() {
		for {
			select {
			case <-ch:
				atomic.AddInt32(&counter, 1)
			case <-terminate:
				return
			case <-time.After(6 * time.Second):
				return
			}
		}
	}()
	time.Sleep(2 * time.Second)
	if tp.Disable() != nil {
		t.Fatal("Failed to disable policy")
	}
	//Waiting for everything is stable
	<-time.After(1 * time.Second)
	//Copy value
	var copiedCounter int32
	atomic.StoreInt32(&copiedCounter, atomic.LoadInt32(&counter))
	time.Sleep(2 * time.Second)
	if atomic.LoadInt32(&counter) != atomic.LoadInt32(&copiedCounter) {
		t.Fatalf("Policy is still running after calling Disable() %d=%d", atomic.LoadInt32(&copiedCounter), atomic.LoadInt32(&counter))
	}
}
