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

package audit

import (
	"bufio"
	"context"
	"net"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestSyslogWriterWrite(t *testing.T) {
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	require.NoError(t, err)
	defer listener.Close()

	messageCh := make(chan string, 1)
	go func() {
		conn, acceptErr := listener.Accept()
		if acceptErr != nil {
			return
		}
		defer conn.Close()
		message, readErr := bufio.NewReader(conn).ReadString('\n')
		if readErr == nil {
			messageCh <- message
		}
	}()

	writer, err := newSyslogWriter(context.Background(), listener.Addr().String())
	require.NoError(t, err)
	defer writer.Close()

	n, err := writer.Write([]byte("audit event"))
	require.NoError(t, err)
	require.Equal(t, len("audit event"), n)

	var message string
	select {
	case message = <-messageCh:
	case <-time.After(time.Second):
		t.Fatal("timed out waiting for syslog message")
	}
	require.True(t, strings.HasPrefix(message, "<6>"))
	require.Contains(t, message, "audit event")
}

func TestLoggerManagerInitRecordsFailedEndpoint(t *testing.T) {
	manager := &LoggerManager{}
	endpoint := "127.0.0.1:1"

	err := manager.Init(context.Background(), endpoint)
	require.Error(t, err)
	require.Equal(t, endpoint, manager.endpoint)
	require.False(t, manager.initialized)
	require.NotNil(t, manager.remoteLogger)
}

func TestNewSyslogWriterHonorsCancelledContext(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	_, err := newSyslogWriter(ctx, "127.0.0.1:1")
	require.Error(t, err)
}
