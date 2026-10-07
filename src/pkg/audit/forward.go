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
	"context"
	"fmt"
	"io"
	"net"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/goharbor/harbor/src/lib/config"
	"github.com/goharbor/harbor/src/lib/log"
)

const (
	syslogDialTimeout  = 5 * time.Second
	syslogWriteTimeout = time.Second
	syslogInfoPriority = 6
)

// LogMgr manage the audit log forward operations
var LogMgr = &LoggerManager{}

// LoggerManager manage the operations related to the audit log
type LoggerManager struct {
	mu           sync.Mutex
	endpoint     string
	initialized  bool
	remoteLogger *log.Logger
	closer       io.Closer
}

// Init redirect the audit log to the forward endpoint
func (a *LoggerManager) Init(ctx context.Context, logEndpoint string) error {
	a.mu.Lock()
	defer a.mu.Unlock()

	return a.initLocked(ctx, logEndpoint)
}

func (a *LoggerManager) initLocked(ctx context.Context, logEndpoint string) error {
	var (
		w   io.Writer = os.Stdout
		err error
	)
	if logEndpoint != "" {
		var writer *syslogWriter
		writer, err = newSyslogWriter(ctx, logEndpoint)
		if err == nil {
			w = writer
		} else {
			log.Errorf("failed to create audit log, error %v", err)
		}
	}

	if a.closer != nil {
		_ = a.closer.Close()
	}
	a.endpoint = logEndpoint
	a.closer = nil
	a.initialized = logEndpoint != "" && err == nil
	if closer, ok := w.(io.Closer); ok {
		a.closer = closer
	}
	a.remoteLogger = log.New(w, log.NewTextFormatter(), log.InfoLevel, 3)
	a.remoteLogger.SetFallback(log.DefaultLogger())
	return err
}

// DefaultLogger ...
func (a *LoggerManager) DefaultLogger(ctx context.Context) *log.Logger {
	endpoint := config.AuditLogForwardEndpoint(ctx)
	a.mu.Lock()
	defer a.mu.Unlock()
	if a.remoteLogger == nil || !strings.EqualFold(a.endpoint, endpoint) {
		_ = a.initLocked(ctx, endpoint)
	}
	return a.remoteLogger
}

// CheckEndpointActive check the liveliness of the endpoint
func CheckEndpointActive(address string) bool {
	al, err := newSyslogWriter(context.Background(), address)
	if al != nil {
		defer al.Close()
	}
	if err != nil {
		log.Errorf("failed to connect to audit log endpoint, error %v", err)
		return false
	}
	return true
}

type syslogWriter struct {
	mu       sync.Mutex
	address  string
	hostname string
	conn     net.Conn
}

func newSyslogWriter(ctx context.Context, address string) (*syslogWriter, error) {
	if ctx == nil {
		ctx = context.Background()
	}
	hostname, _ := os.Hostname()
	w := &syslogWriter{address: address, hostname: hostname}
	if err := w.connect(ctx); err != nil {
		return nil, err
	}
	return w, nil
}

func (w *syslogWriter) connect(ctx context.Context) error {
	dialCtx, cancel := context.WithTimeout(ctx, syslogDialTimeout)
	defer cancel()
	conn, err := (&net.Dialer{Timeout: syslogDialTimeout}).DialContext(dialCtx, "tcp", w.address)
	if err != nil {
		return err
	}
	w.conn = conn
	if w.hostname == "" {
		w.hostname = conn.LocalAddr().String()
	}
	return nil
}

func (w *syslogWriter) Write(b []byte) (int, error) {
	w.mu.Lock()
	defer w.mu.Unlock()

	if err := w.write(b); err != nil {
		return 0, err
	}
	return len(b), nil
}

func (w *syslogWriter) write(b []byte) error {
	if w.conn == nil {
		return fmt.Errorf("syslog connection is not initialized")
	}
	if err := w.conn.SetWriteDeadline(time.Now().Add(syslogWriteTimeout)); err != nil {
		return err
	}
	message := string(b)
	if !strings.HasSuffix(message, "\n") {
		message += "\n"
	}
	_, err := fmt.Fprintf(w.conn, "<%d>%s %s audit[%d]: %s",
		syslogInfoPriority, time.Now().Format(time.RFC3339), w.hostname, os.Getpid(), message)
	return err
}

func (w *syslogWriter) Close() error {
	w.mu.Lock()
	defer w.mu.Unlock()
	if w.conn == nil {
		return nil
	}
	err := w.conn.Close()
	w.conn = nil
	return err
}
