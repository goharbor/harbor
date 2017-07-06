package notifier

import (
	"testing"
	"time"

	"github.com/vmware/harbor/src/common/scheduler"
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

	//Waiting for everything is ready.
	<-time.After(1 * time.Second)
	if !testingScheduler.HasScheduled("Alternate Policy") {
		t.Fatal("Handler does not work")
	}

	notification2 := ScanPolicyNotification{"none", 0}
	if err := handler.Handle(notification2); err != nil {
		t.Fatal(err)
	}

	//Waiting for everything is ready.
	<-time.After(1 * time.Second)
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
