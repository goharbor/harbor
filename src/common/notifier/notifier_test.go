package notifier

import (
	"reflect"
	"sync/atomic"
	"testing"
	"time"

	"github.com/vmware/harbor/src/common/scheduler"
)

var statefulData int32

type fakeStatefulHandler struct {
	number int
}

func (fsh *fakeStatefulHandler) IsStateful() bool {
	return true
}

func (fsh *fakeStatefulHandler) Handle(v interface{}) error {
	increment := 0
	if v != nil && reflect.TypeOf(v).Kind() == reflect.Int {
		increment = v.(int)
	}
	atomic.AddInt32(&statefulData, (int32)(increment))
	return nil
}

type fakeStatelessHandler struct{}

func (fsh *fakeStatelessHandler) IsStateful() bool {
	return false
}

func (fsh *fakeStatelessHandler) Handle(v interface{}) error {
	return nil
}

func TestSubscribeAndUnSubscribe(t *testing.T) {
	count := len(notificationWatcher.handlers)
	err := Subscribe("topic1", &fakeStatefulHandler{0})
	if err != nil {
		t.Fatal(err)
	}

	err = Subscribe("topic1", &fakeStatelessHandler{})
	if err != nil {
		t.Fatal(err)
	}

	err = Subscribe("topic2", &fakeStatefulHandler{0})
	if err != nil {
		t.Fatal(err)
	}

	err = Subscribe("topic2", &fakeStatelessHandler{})
	if err != nil {
		t.Fatal(err)
	}

	if len(notificationWatcher.handlers) != (count + 2) {
		t.Fail()
	}

	if indexer, ok := notificationWatcher.handlers["topic1"]; !ok {
		t.Fail()
	} else {
		if len(indexer) != 2 {
			t.Fail()
		}
	}

	if len(notificationWatcher.handlerChannels) != 1 {
		t.Fail()
	}

	err = UnSubscribe("topic1", "*notifier.fakeStatefulHandler")
	if err != nil {
		t.Fatal(err)
	}

	err = UnSubscribe("topic2", "*notifier.fakeStatefulHandler")
	if err != nil {
		t.Fatal(err)
	}

	if len(notificationWatcher.handlerChannels) != 0 {
		t.Fail()
	}

	err = UnSubscribe("topic1", "")
	if err != nil {
		t.Fatal(err)
	}

	if len(notificationWatcher.handlers) != (count + 1) {
		t.Fail()
	}

	err = UnSubscribe("topic2", "")
	if err != nil {
		t.Fatal(err)
	}

	if len(notificationWatcher.handlers) != count {
		t.Fail()
	}
}

func TestPublish(t *testing.T) {
	count := len(notificationWatcher.handlers)
	err := Subscribe("topic1", &fakeStatefulHandler{0})
	if err != nil {
		t.Fatal(err)
	}

	err = Subscribe("topic2", &fakeStatefulHandler{0})
	if err != nil {
		t.Fatal(err)
	}

	if len(notificationWatcher.handlers) != (count + 2) {
		t.Fail()
	}

	Publish("topic1", 100)
	Publish("topic2", 50)

	//Waiting for async is done
	<-time.After(1 * time.Second)

	finalData := atomic.LoadInt32(&statefulData)
	if finalData != 150 {
		t.Fatalf("Expect execution result %d, but got %d", 150, finalData)
	}

	err = UnSubscribe("topic1", "")
	if err != nil {
		t.Fatal(err)
	}

	err = UnSubscribe("topic2", "*notifier.fakeStatefulHandler")
	if err != nil {
		t.Fatal(err)
	}

	//Clear stateful data.
	atomic.StoreInt32(&statefulData, 0)
}

func TestConcurrentPublish(t *testing.T) {
	count := len(notificationWatcher.handlers)
	err := Subscribe("topic1", &fakeStatefulHandler{0})
	if err != nil {
		t.Fatal(err)
	}

	if len(notificationWatcher.handlers) != (count + 1) {
		t.Fail()
	}

	//Publish in a short interval.
	for i := 0; i < 10; i++ {
		Publish("topic1", 100)
	}

	//Waiting for async is done
	<-time.After(1 * time.Second)

	finalData := atomic.LoadInt32(&statefulData)
	if finalData != 1000 {
		t.Fatalf("Expect execution result %d, but got %d", 1000, finalData)
	}

	err = UnSubscribe("topic1", "")
	if err != nil {
		t.Fatal(err)
	}

	//Clear stateful data.
	atomic.StoreInt32(&statefulData, 0)
}

func TestConcurrentPublishWithScanPolicyHandler(t *testing.T) {
	scheduler.DefaultScheduler.Start()
	if !scheduler.DefaultScheduler.IsRunning() {
		t.Fatal("Policy scheduler is not started")
	}

	count := len(notificationWatcher.handlers)
	if err := Subscribe("testing_topic", &ScanPolicyNotificationHandler{}); err != nil {
		t.Fatal(err.Error())
	}
	if len(notificationWatcher.handlers) != (count + 1) {
		t.Fatalf("Handler is not registered")
	}

	utcTime := time.Now().UTC().Unix()
	notification := ScanPolicyNotification{"daily", utcTime + 3600}
	for i := 1; i <= 10; i++ {
		notification.DailyTime += (int64)(i)
		if err := Publish("testing_topic", notification); err != nil {
			t.Fatalf("index=%d, error=%s", i, err.Error())
		}
	}

	//Wating for everything is ready.
	<-time.After(2 * time.Second)

	if err := UnSubscribe("testing_topic", ""); err != nil {
		t.Fatal(err.Error())
	}

	if len(notificationWatcher.handlers) != count {
		t.Fatal("Handler is not unregistered")
	}

	scheduler.DefaultScheduler.Stop()
	//Wating for everything is ready.
	<-time.After(1 * time.Second)
	if scheduler.DefaultScheduler.IsRunning() {
		t.Fatal("Policy scheduler is not stopped")
	}

}
