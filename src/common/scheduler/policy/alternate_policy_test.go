package policy

import (
	"testing"
	"time"
)

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
		Duration:       1 * time.Second,
		OffsetTime:     utcOffset + 1,
		StartTimestamp: -1,
		EndTimestamp:   now.Add(3 * time.Second).Unix(),
	})
	err := tp.AttachTasks(&fakeTask{number: 100})
	if err != nil {
		t.Fail()
	}
	ch := tp.Evaluate()
	done := tp.Done()
	counter := 0
READ_SIGNAL:
	for {
		select {
		case <-ch:
			counter++
		case <-done:
			break READ_SIGNAL
		case <-time.After(5 * time.Second):
			t.Fail()
			return
		}
	}

	if counter != 2 {
		t.Fail()
	}
}

func TestDisablePolicy(t *testing.T) {
	now := time.Now().UTC()
	utcOffset := (int64)(now.Hour()*3600 + now.Minute()*60)
	tp := NewAlternatePolicy(&AlternatePolicyConfiguration{
		Duration:       1 * time.Second,
		OffsetTime:     utcOffset + 1,
		StartTimestamp: -1,
		EndTimestamp:   -1,
	})
	err := tp.AttachTasks(&fakeTask{number: 100})
	if err != nil {
		t.Fail()
	}
	ch := tp.Evaluate()
	counter := 0
	terminate := make(chan bool)
	defer func() {
		terminate <- true
	}()
	go func() {
		for {
			select {
			case <-ch:
				counter++
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
	//Waiting for everything is stabel
	<-time.After(1 * time.Second)
	//Copy value
	copiedCounter := counter
	time.Sleep(2 * time.Second)
	if counter != copiedCounter {
		t.Fatalf("Policy is still running after calling Disable() %d=%d", copiedCounter, counter)
	}
}
