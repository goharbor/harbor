// Copyright 2018 The Harbor Authors. All rights reserved.
package pool

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/vmware/harbor/src/jobservice/opm"

	"github.com/vmware/harbor/src/jobservice/models"
	"github.com/vmware/harbor/src/jobservice/period"
	"github.com/vmware/harbor/src/jobservice/utils"

	"github.com/vmware/harbor/src/jobservice/tests"
)

var redisPool = tests.GiveMeRedisPool()

func TestPublishPolicy(t *testing.T) {
	ms, cancel := createMessageServer()
	err := ms.Subscribe(period.EventSchedulePeriodicPolicy, func(data interface{}) error {
		if _, ok := data.(*period.PeriodicJobPolicy); !ok {
			t.Fatal("expect PeriodicJobPolicy but got other thing")
			return errors.New("expect PeriodicJobPolicy but got other thing")
		}
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}

	err = ms.Subscribe(period.EventUnSchedulePeriodicPolicy, func(data interface{}) error {
		if _, ok := data.(*period.PeriodicJobPolicy); !ok {
			t.Fatal("expect PeriodicJobPolicy but got other thing")
			return errors.New("expect PeriodicJobPolicy but got other thing")
		}

		return nil
	})
	if err != nil {
		t.Fatal(err)
	}

	go func() {
		defer cancel()
		//wait and then publish
		<-time.After(200 * time.Millisecond)

		p := &period.PeriodicJobPolicy{
			PolicyID: "fake_ID",
			JobName:  "fake_job",
			CronSpec: "5 * * * *",
		}
		notification := &models.Message{
			Event: period.EventSchedulePeriodicPolicy,
			Data:  p,
		}

		rawJSON, err := json.Marshal(notification)
		if err != nil {
			t.Fatal(err)
		}

		conn := redisPool.Get()
		defer conn.Close()
		err = conn.Send("PUBLISH", utils.KeyPeriodicNotification(tests.GiveMeTestNamespace()), rawJSON)
		if err != nil {
			t.Fatal(err)
		}

		notification.Event = period.EventUnSchedulePeriodicPolicy
		rawJSON, err = json.Marshal(notification)
		if err != nil {
			t.Fatal(err)
		}
		err = conn.Send("PUBLISH", utils.KeyPeriodicNotification(tests.GiveMeTestNamespace()), rawJSON)
		if err != nil {
			t.Fatal(err)
		}

		//send quit signal
		<-time.After(200 * time.Millisecond)
		err = tests.Clear(utils.KeyPeriodicNotification(tests.GiveMeTestNamespace()), conn)
		if err != nil {
			t.Fatal(err)
		}
	}()

	ms.Start()
}

func TestPublishHook(t *testing.T) {
	ms, cancel := createMessageServer()
	err := ms.Subscribe(opm.EventRegisterStatusHook, func(data interface{}) error {
		if _, ok := data.(*opm.HookData); !ok {
			t.Fatal("expect HookData but got other thing")
			return errors.New("expect HookData but got other thing")
		}
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}

	go func() {
		defer cancel()

		<-time.After(200 * time.Millisecond)
		hook := &opm.HookData{
			JobID:   "fake_job_ID",
			HookURL: "http://localhost:9999/hook",
		}
		notification := &models.Message{
			Event: opm.EventRegisterStatusHook,
			Data:  hook,
		}

		rawJSON, err := json.Marshal(notification)
		if err != nil {
			t.Fatal(err)
		}

		conn := redisPool.Get()
		defer conn.Close()
		err = conn.Send("PUBLISH", utils.KeyPeriodicNotification(tests.GiveMeTestNamespace()), rawJSON)
		if err != nil {
			t.Fatal(err)
		}

		//send quit signal
		<-time.After(200 * time.Millisecond)
		err = tests.Clear(utils.KeyPeriodicNotification(tests.GiveMeTestNamespace()), conn)
		if err != nil {
			t.Fatal(err)
		}
	}()

	ms.Start()
}

func TestPublishCommands(t *testing.T) {
	ms, cancel := createMessageServer()
	err := ms.Subscribe(opm.EventFireCommand, func(data interface{}) error {
		cmds, ok := data.([]string)
		if !ok {
			t.Fatal("expect fired command but got other thing")
			return errors.New("expect fired command but got other thing")
		}
		if len(cmds) != 2 {
			t.Fatalf("expect a array with 2 items but only got '%d' items", len(cmds))
			return fmt.Errorf("expect a array with 2 items but only got '%d' items", len(cmds))
		}
		if cmds[1] != "stop" {
			t.Fatalf("expect command 'stop' but got '%s'", cmds[1])
			return fmt.Errorf("expect command 'stop' but got '%s'", cmds[1])
		}
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}

	go func() {
		defer cancel()
		<-time.After(200 * time.Millisecond)

		notification := &models.Message{
			Event: opm.EventRegisterStatusHook,
			Data:  []string{"fake_job_ID", "stop"},
		}

		rawJSON, err := json.Marshal(notification)
		if err != nil {
			t.Fatal(err)
		}

		conn := redisPool.Get()
		defer conn.Close()
		err = conn.Send("PUBLISH", utils.KeyPeriodicNotification(tests.GiveMeTestNamespace()), rawJSON)
		if err != nil {
			t.Fatal(err)
		}

		//hold for a while
		<-time.After(200 * time.Millisecond)
	}()

	ms.Start()
}

func createMessageServer() (*MessageServer, context.CancelFunc) {
	ns := tests.GiveMeTestNamespace()
	ctx, cancel := context.WithCancel(context.Background())
	return NewMessageServer(ctx, ns, redisPool), cancel
}
