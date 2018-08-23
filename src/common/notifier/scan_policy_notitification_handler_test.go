package notifier

import (
	"testing"
	"time"

	"github.com/goharbor/harbor/src/common/scheduler"
	"github.com/goharbor/harbor/src/common/scheduler/policy"
)

var testingScheduler = scheduler.DefaultScheduler

func TestScanPolicyNotificationHandler(t *testing.T) {
	//Scheduler should be running.
	testingScheduler.Start()
	if !testingScheduler.IsRunning() {
		t.Fatal("scheduler should be running")
	}

	handler := &ScanPolicyNotificationHandler{}
	if !handler.IsStateful() {
		t.Fail()
	}

	utcTime := time.Now().UTC().Unix()
	notification := ScanPolicyNotification{"daily", utcTime + 3600}
	if err := handler.Handle(notification); err != nil {
		t.Fatal(err)
	}

	if !testingScheduler.HasScheduled("Alternate Policy") {
		t.Fatal("Handler does not work")
	}

	//Policy parameter changed.
	notification2 := ScanPolicyNotification{"daily", utcTime + 7200}
	if err := handler.Handle(notification2); err != nil {
		t.Fatal(err)
	}

	if !testingScheduler.HasScheduled("Alternate Policy") {
		t.Fatal("Handler does not work [2]")
	}
	pl := testingScheduler.GetPolicy("Alternate Policy")
	if pl == nil {
		t.Fail()
	}
	spl := pl.(*policy.AlternatePolicy)
	cfg := spl.GetConfig()
	if cfg == nil {
		t.Fail()
	}
	if cfg.OffsetTime != utcTime+7200 {
		t.Fatal("Policy is not updated")
	}

	notification3 := ScanPolicyNotification{"none", 0}
	if err := handler.Handle(notification3); err != nil {
		t.Fatal(err)
	}

	if testingScheduler.HasScheduled("Alternate Policy") {
		t.Fail()
	}

	//Clear
	testingScheduler.Stop()
	//Waiting for everything is ready.
	<-time.After(1 * time.Second)
	if testingScheduler.IsRunning() {
		t.Fatal("scheduler should be stopped")
	}
}
