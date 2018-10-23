// Copyright Project Harbor Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//    http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package opm

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/goharbor/harbor/src/jobservice/logger"
	"github.com/goharbor/harbor/src/jobservice/models"
	"github.com/goharbor/harbor/src/jobservice/utils"
	"github.com/gomodule/redigo/redis"
)

const (
	commandValidTime       = 5 * time.Minute
	commandSweepTickerTime = 1 * time.Hour
	// EventFireCommand for firing command event
	EventFireCommand = "fire_command"
)

type oPCommand struct {
	command  string
	fireTime int64
}

// oPCommands maintain commands list
type oPCommands struct {
	lock      *sync.RWMutex
	commands  map[string]*oPCommand
	context   context.Context
	redisPool *redis.Pool
	namespace string
	stopChan  chan struct{}
	doneChan  chan struct{}
}

// newOPCommands is constructor of OPCommands
func newOPCommands(ctx context.Context, ns string, redisPool *redis.Pool) *oPCommands {
	return &oPCommands{
		lock:      new(sync.RWMutex),
		commands:  make(map[string]*oPCommand),
		context:   ctx,
		redisPool: redisPool,
		namespace: ns,
		stopChan:  make(chan struct{}, 1),
		doneChan:  make(chan struct{}, 1),
	}
}

// Start the command sweeper
func (opc *oPCommands) Start() {
	go opc.loop()
	logger.Info("OP commands sweeper is started")
}

// Stop the command sweeper
func (opc *oPCommands) Stop() {
	opc.stopChan <- struct{}{}
	<-opc.doneChan
}

// Fire command
func (opc *oPCommands) Fire(jobID string, command string) error {
	if utils.IsEmptyStr(jobID) {
		return errors.New("empty job ID")
	}

	if command != CtlCommandStop && command != CtlCommandCancel {
		return fmt.Errorf("Unsupported command %s", command)
	}

	notification := &models.Message{
		Event: EventFireCommand,
		Data:  []string{jobID, command},
	}

	rawJSON, err := json.Marshal(notification)
	if err != nil {
		return err
	}

	conn := opc.redisPool.Get()
	defer conn.Close()

	_, err = conn.Do("PUBLISH", utils.KeyPeriodicNotification(opc.namespace), rawJSON)

	return err
}

// Push command into the list
func (opc *oPCommands) Push(jobID string, command string) error {
	if utils.IsEmptyStr(jobID) {
		return errors.New("empty job ID")
	}

	if command != CtlCommandStop && command != CtlCommandCancel {
		return fmt.Errorf("Unsupported command %s", command)
	}

	opc.lock.Lock()
	defer opc.lock.Unlock()

	opc.commands[jobID] = &oPCommand{
		command:  command,
		fireTime: time.Now().Unix(),
	}

	return nil
}

// Pop out the command if existing
func (opc *oPCommands) Pop(jobID string) (string, bool) {
	if utils.IsEmptyStr(jobID) {
		return "", false
	}

	opc.lock.RLock()
	defer opc.lock.RUnlock()

	c, ok := opc.commands[jobID]
	if ok {
		if time.Unix(c.fireTime, 0).Add(commandValidTime).After(time.Now()) {
			delete(opc.commands, jobID)
			return c.command, true
		}
	}

	return "", false
}

func (opc *oPCommands) loop() {
	defer func() {
		logger.Info("OP commands is stopped")
		opc.doneChan <- struct{}{}
	}()

	tk := time.NewTicker(commandSweepTickerTime)
	defer tk.Stop()

	for {
		select {
		case <-tk.C:
			opc.sweepCommands()
		case <-opc.context.Done():
			return
		case <-opc.stopChan:
			return
		}
	}
}

func (opc *oPCommands) sweepCommands() {
	opc.lock.Lock()
	defer opc.lock.Unlock()

	for k, v := range opc.commands {
		if time.Unix(v.fireTime, 0).Add(commandValidTime).After(time.Now()) {
			delete(opc.commands, k)
		}
	}
}
