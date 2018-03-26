// Copyright 2018 The Harbor Authors. All rights reserved.

package pool

import (
	"context"
	"encoding/json"
	"errors"
	"reflect"
	"time"

	"github.com/vmware/harbor/src/jobservice_v2/logger"
	"github.com/vmware/harbor/src/jobservice_v2/opm"
	"github.com/vmware/harbor/src/jobservice_v2/period"

	"github.com/garyburd/redigo/redis"
	"github.com/vmware/harbor/src/jobservice_v2/models"
	"github.com/vmware/harbor/src/jobservice_v2/utils"
)

//MessageServer implements the sub/pub mechanism via redis to do async message exchanging.
type MessageServer struct {
	context   context.Context
	redisPool *redis.Pool
	namespace string
	callbacks map[string]reflect.Value //no need to sync
}

//NewMessageServer creates a new ptr of MessageServer
func NewMessageServer(ctx context.Context, namespace string, redisPool *redis.Pool) *MessageServer {
	return &MessageServer{
		context:   ctx,
		redisPool: redisPool,
		namespace: namespace,
		callbacks: make(map[string]reflect.Value),
	}
}

//Start to serve
func (ms *MessageServer) Start() error {
	defer func() {
		logger.Info("Message server is stopped")
	}()

	//As we get one connection from the pool, don't try to close it.
	conn := ms.redisPool.Get()
	defer conn.Close()

	psc := redis.PubSubConn{
		Conn: conn,
	}

	err := psc.Subscribe(redis.Args{}.AddFlat(utils.KeyPeriodicNotification(ms.namespace))...)
	if err != nil {
		return err
	}

	done := make(chan error, 1)
	go func() {
		for {
			switch res := psc.Receive().(type) {
			case error:
				done <- res
				return
			case redis.Message:
				m := &models.Message{}
				if err := json.Unmarshal(res.Data, m); err != nil {
					//logged
					logger.Warningf("read invalid message: %s\n", res.Data)
				}
				if callback, ok := ms.callbacks[m.Event]; !ok {
					//logged
					logger.Warningf("no handler to handle event %s\n", m.Event)
				} else {
					//Try to recover the concrete type
					var converted interface{}
					switch m.Event {
					case period.EventSchedulePeriodicPolicy,
						period.EventUnSchedulePeriodicPolicy:
						//ignore error, actually error should not be happend because we did not change data
						//after the last unmarshal try.
						policyObject := &period.PeriodicJobPolicy{}
						dt, _ := json.Marshal(m.Data)
						json.Unmarshal(dt, policyObject)
						converted = policyObject
					case opm.EventRegisterStatusHook:
						//ignore error
						hookObject := &opm.HookData{}
						dt, _ := json.Marshal(m.Data)
						json.Unmarshal(dt, hookObject)
						converted = hookObject
					}
					res := callback.Call([]reflect.Value{reflect.ValueOf(converted)})
					e := res[0].Interface()
					if e != nil {
						err := e.(error)
						//logged
						logger.Errorf("failed to fire callback with error: %s\n", err)
					}
				}
			case redis.Subscription:
				switch res.Kind {
				case "subscribe":
					logger.Infof("Subscribe redis channel %s\n", res.Channel)
					break
				case "unsubscribe":
					//Unsubscribe all, means main goroutine is exiting
					logger.Infof("Unsubscribe redis channel %s\n", res.Channel)
					done <- nil
					return
				}
			}
		}
	}()

	logger.Info("Message server is started")

	ticker := time.NewTicker(time.Minute)
	defer ticker.Stop()

	//blocking here
	for err == nil {
		select {
		case <-ticker.C:
			err = psc.Ping("ping!")
		case <-ms.context.Done():
			err = errors.New("context exit")
		case err = <-done:
			return err
		}
	}

	//Unsubscribe all
	psc.Unsubscribe()
	return <-done
}

//Subscribe event with specified callback
func (ms *MessageServer) Subscribe(event string, callback interface{}) error {
	if utils.IsEmptyStr(event) {
		return errors.New("empty event is not allowed")
	}

	handler, err := validateCallbackFunc(callback)
	if err != nil {
		return err
	}

	ms.callbacks[event] = handler
	return nil
}

func validateCallbackFunc(callback interface{}) (reflect.Value, error) {
	if callback == nil {
		return reflect.ValueOf(nil), errors.New("nil callback handler")
	}

	vFn := reflect.ValueOf(callback)
	vFType := vFn.Type()
	if vFType.Kind() != reflect.Func {
		return reflect.ValueOf(nil), errors.New("callback handler must be a generic func")
	}

	inNum := vFType.NumIn()
	outNum := vFType.NumOut()
	if inNum != 1 || outNum != 1 {
		return reflect.ValueOf(nil), errors.New("callback handler can only be func(interface{})error format")
	}

	inType := vFType.In(0)
	var intf *interface{}
	if inType != reflect.TypeOf(intf).Elem() {
		return reflect.ValueOf(nil), errors.New("callback handler can only be func(interface{})error format")
	}

	outType := vFType.Out(0)
	var e *error
	if outType != reflect.TypeOf(e).Elem() {
		return reflect.ValueOf(nil), errors.New("callback handler can only be func(interface{})error format")
	}

	return vFn, nil
}
