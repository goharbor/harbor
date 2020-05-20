package notifier

import (
	"reflect"
	"sync/atomic"
	"testing"
	"time"
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

	// Waiting for async is done
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

	// Clear stateful data.
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

	// Publish in a short interval.
	for i := 0; i < 10; i++ {
		Publish("topic1", 100)
	}

	// Waiting for async is done
	<-time.After(1 * time.Second)

	finalData := atomic.LoadInt32(&statefulData)
	if finalData != 1000 {
		t.Fatalf("Expect execution result %d, but got %d", 1000, finalData)
	}

	err = UnSubscribe("topic1", "")
	if err != nil {
		t.Fatal(err)
	}

	// Clear stateful data.
	atomic.StoreInt32(&statefulData, 0)
}
