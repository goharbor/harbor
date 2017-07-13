package notifier

import (
	"reflect"
	"testing"
	"time"
)

var statefulData int

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
	statefulData += increment
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

	if len(notificationWatcher.handlers) != 2 {
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

	if len(notificationWatcher.handlers) != 1 {
		t.Fail()
	}

	err = UnSubscribe("topic2", "")
	if err != nil {
		t.Fatal(err)
	}

	if len(notificationWatcher.handlers) != 0 {
		t.Fail()
	}
}

func TestPublish(t *testing.T) {
	err := Subscribe("topic1", &fakeStatefulHandler{0})
	if err != nil {
		t.Fatal(err)
	}

	err = Subscribe("topic2", &fakeStatefulHandler{0})
	if err != nil {
		t.Fatal(err)
	}

	if len(notificationWatcher.handlers) != 2 {
		t.Fail()
	}

	Publish("topic1", 100)
	Publish("topic2", 50)

	//Waiting for async is done
	<-time.After(1 * time.Second)

	if statefulData != 150 {
		t.Fatalf("Expect execution result %d, but got %d", 150, statefulData)
	}

	err = UnSubscribe("topic1", "*notifier.fakeStatefulHandler")
	if err != nil {
		t.Fatal(err)
	}

	err = UnSubscribe("topic2", "*notifier.fakeStatefulHandler")
	if err != nil {
		t.Fatal(err)
	}
}
