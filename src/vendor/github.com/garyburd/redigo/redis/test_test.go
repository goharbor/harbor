// Copyright 2012 Gary Burd
//
// Licensed under the Apache License, Version 2.0 (the "License"): you may
// not use this file except in compliance with the License. You may obtain
// a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS, WITHOUT
// WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the
// License for the specific language governing permissions and limitations
// under the License.

package redis

import (
	"bufio"
	"errors"
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"sync"
	"testing"
	"time"
)

func SetNowFunc(f func() time.Time) {
	nowFunc = f
}

var (
	ErrNegativeInt = errNegativeInt

	serverPath     = flag.String("redis-server", "redis-server", "Path to redis server binary")
	serverAddress  = flag.String("redis-address", "127.0.0.1", "The address of the server")
	serverBasePort = flag.Int("redis-port", 16379, "Beginning of port range for test servers")
	serverLogName  = flag.String("redis-log", "", "Write Redis server logs to `filename`")
	serverLog      = ioutil.Discard

	defaultServerMu  sync.Mutex
	defaultServer    *Server
	defaultServerErr error
)

type Server struct {
	name string
	cmd  *exec.Cmd
	done chan struct{}
}

func NewServer(name string, args ...string) (*Server, error) {
	s := &Server{
		name: name,
		cmd:  exec.Command(*serverPath, args...),
		done: make(chan struct{}),
	}

	r, err := s.cmd.StdoutPipe()
	if err != nil {
		return nil, err
	}

	err = s.cmd.Start()
	if err != nil {
		return nil, err
	}

	ready := make(chan error, 1)
	go s.watch(r, ready)

	select {
	case err = <-ready:
	case <-time.After(time.Second * 10):
		err = errors.New("timeout waiting for server to start")
	}

	if err != nil {
		s.Stop()
		return nil, err
	}

	return s, nil
}

func (s *Server) watch(r io.Reader, ready chan error) {
	fmt.Fprintf(serverLog, "%d START %s \n", s.cmd.Process.Pid, s.name)
	var listening bool
	var text string
	scn := bufio.NewScanner(r)
	for scn.Scan() {
		text = scn.Text()
		fmt.Fprintf(serverLog, "%s\n", text)
		if !listening {
			if strings.Contains(text, " * Ready to accept connections") ||
				strings.Contains(text, " * The server is now ready to accept connections on port") {
				listening = true
				ready <- nil
			}
		}
	}
	if !listening {
		ready <- fmt.Errorf("server exited: %s", text)
	}
	s.cmd.Wait()
	fmt.Fprintf(serverLog, "%d STOP %s \n", s.cmd.Process.Pid, s.name)
	close(s.done)
}

func (s *Server) Stop() {
	s.cmd.Process.Signal(os.Interrupt)
	<-s.done
}

// stopDefaultServer stops the server created by DialDefaultServer.
func stopDefaultServer() {
	defaultServerMu.Lock()
	defer defaultServerMu.Unlock()
	if defaultServer != nil {
		defaultServer.Stop()
		defaultServer = nil
	}
}

// DefaultServerAddr starts the test server if not already started and returns
// the address of that server.
func DefaultServerAddr() (string, error) {
	defaultServerMu.Lock()
	defer defaultServerMu.Unlock()
	addr := fmt.Sprintf("%v:%d", *serverAddress, *serverBasePort)
	if defaultServer != nil || defaultServerErr != nil {
		return addr, defaultServerErr
	}
	defaultServer, defaultServerErr = NewServer(
		"default",
		"--port", strconv.Itoa(*serverBasePort),
		"--bind", *serverAddress,
		"--save", "",
		"--appendonly", "no")
	return addr, defaultServerErr
}

// DialDefaultServer starts the test server if not already started and dials a
// connection to the server.
func DialDefaultServer() (Conn, error) {
	addr, err := DefaultServerAddr()
	if err != nil {
		return nil, err
	}
	c, err := Dial("tcp", addr, DialReadTimeout(1*time.Second), DialWriteTimeout(1*time.Second))
	if err != nil {
		return nil, err
	}
	c.Do("FLUSHDB")
	return c, nil
}

func TestMain(m *testing.M) {
	os.Exit(func() int {
		flag.Parse()

		var f *os.File
		if *serverLogName != "" {
			var err error
			f, err = os.OpenFile(*serverLogName, os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0600)
			if err != nil {
				fmt.Fprintf(os.Stderr, "Error opening redis-log: %v\n", err)
				return 1
			}
			defer f.Close()
			serverLog = f
		}

		defer stopDefaultServer()

		return m.Run()
	}())
}
