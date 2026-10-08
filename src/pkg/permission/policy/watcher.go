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

package policy

import (
	"context"
	"errors"
	"strings"
	"sync"
	"time"

	"github.com/jackc/pgx/v5"

	"github.com/goharbor/harbor/src/lib/log"
)

// Channel is the Postgres channel the casbin_rule and role triggers notify on.
const Channel = "harbor_policy"

const (
	// pingEvery bounds how long the watcher sits on a socket without checking
	// it is still alive. A connection through a load balancer can be gone long
	// before a blocking read notices.
	pingEvery = 30 * time.Second

	// debounceWindow coalesces a burst into one reload. A role update is a
	// delete and an insert, and the fleet does not need to rebuild twice for
	// one API call.
	debounceWindow = 40 * time.Millisecond

	maxBackoff = 5 * time.Second

	// callbackTimeout bounds one reload or version check. Close cancels the
	// watcher's context and waits for the dispatch goroutine, so a database
	// call that can neither finish nor be cancelled would hold shutdown open.
	callbackTimeout = 60 * time.Second

	// resyncEvery bounds how long this replica can hold a stale policy when no
	// notification ever arrives. Harbor runs several cores against one
	// database, and a notification is a best effort: the connection can drop
	// between the write and the reconnect, Postgres can restart, this process
	// can be paused past the message. The watcher is the fast path and this is
	// the one that guarantees convergence.
	resyncEvery = pingEvery
)

// Watcher tells this replica when the policy changed somewhere else.
//
// It does not carry the change. The notification says only "transaction 748,
// from core-7" and the replica goes back to the table, so a lost notification
// self-corrects on the next one and a replica can never diverge from what is
// stored. Watchers that apply the rules carried in the payload cannot say that:
// one missed message and they are wrong until something restarts them.
//
// The NOTIFY is not sent from here. An AFTER trigger on casbin_rule sends it
// from inside the writing transaction, so it cannot survive a rollback and it
// fires for policy changes that never went through Harbor at all.
type Watcher struct {
	connString string
	channel    string
	origin     string
	onChange   func(context.Context, string) error
	onTick     func(context.Context) error

	mu    sync.Mutex
	stats WatcherStats

	events chan string
	ticks  chan struct{}
	ctx    context.Context
	cancel context.CancelFunc
	wg     sync.WaitGroup
}

// WatcherStats is what the watcher has seen since this replica started.
type WatcherStats struct {
	Notifications  int       `json:"notifications"`
	SelfOriginated int       `json:"self_originated"`
	Reloads        int       `json:"reloads"`
	Reconnects     int       `json:"reconnects"`
	Resyncs        int       `json:"resyncs"`
	Retries        int       `json:"retries"`
	Errors         int       `json:"errors"`
	LastPayload    string    `json:"last_payload"`
	LastNotifyAt   time.Time `json:"last_notify_at"`
}

// NewWatcher opens the LISTEN connection and starts watching. It connects once
// synchronously, so a bad DSN fails the replica's boot instead of becoming a
// goroutine that logs into the void.
func NewWatcher(connString string, channel string, origin string, onChange func(context.Context, string) error, onTick func(context.Context) error) (*Watcher, error) {
	ctx, cancel := context.WithCancel(context.Background())
	w := &Watcher{
		connString: connString,
		channel:    channel,
		origin:     origin,
		onChange:   onChange,
		onTick:     onTick,
		events:     make(chan string, 64),
		ticks:      make(chan struct{}, 1),
		ctx:        ctx,
		cancel:     cancel,
	}

	conn, err := w.dial(ctx)
	if err != nil {
		cancel()
		return nil, err
	}

	w.wg.Add(2)
	go w.listen(conn)
	go w.dispatch()
	return w, nil
}

// Close stops the watcher and waits for its goroutines.
func (w *Watcher) Close() {
	w.cancel()
	w.wg.Wait()
}

// Stats snapshots the counters.
func (w *Watcher) Stats() WatcherStats {
	w.mu.Lock()
	defer w.mu.Unlock()
	return w.stats
}

func (w *Watcher) dial(ctx context.Context) (*pgx.Conn, error) {
	conn, err := pgx.Connect(ctx, w.connString)
	if err != nil {
		return nil, err
	}
	// Sanitize quotes the channel name rather than pasting it into the statement.
	if _, err := conn.Exec(ctx, "LISTEN "+pgx.Identifier{w.channel}.Sanitize()); err != nil {
		_ = conn.Close(ctx)
		return nil, err
	}
	return conn, nil
}

// listen owns the LISTEN connection for the life of the replica, reconnecting
// with backoff whenever it drops.
func (w *Watcher) listen(conn *pgx.Conn) {
	defer w.wg.Done()

	backoff := 250 * time.Millisecond
	for {
		w.readUntilBroken(conn)
		_ = conn.Close(context.Background())
		if w.ctx.Err() != nil {
			return
		}

		w.bump(func(s *WatcherStats) { s.Reconnects++ })
		log.Warningf("policy watcher lost its connection, reconnecting")

		for {
			select {
			case <-time.After(backoff):
			case <-w.ctx.Done():
				return
			}
			var err error
			if conn, err = w.dial(w.ctx); err == nil {
				break
			}
			w.bump(func(s *WatcherStats) { s.Errors++ })
			if backoff *= 2; backoff > maxBackoff {
				backoff = maxBackoff
			}
		}
		backoff = 250 * time.Millisecond

		// We were deaf for an unknown interval, so reload rather than wait for
		// the next write. A blip must not strand a replica on a stale policy.
		w.record("reconnect")
		w.emit("reconnect")
	}
}

// readUntilBroken returns when the connection is no longer usable, or when the
// watcher is closing.
func (w *Watcher) readUntilBroken(conn *pgx.Conn) {
	for {
		ctx, cancel := context.WithTimeout(w.ctx, pingEvery)
		n, err := conn.WaitForNotification(ctx)
		cancel()

		switch {
		case err == nil:
			w.record(n.Payload)
			// Including our own writes. Writing the row does not update this
			// replica's enforcer, and nothing else does either, so a replica
			// that skipped its own notification would be the last one serving
			// the change an operator just made on it.
			w.emit(n.Payload)

		case w.ctx.Err() != nil:
			return // closing

		case errors.Is(err, context.DeadlineExceeded):
			ping, pcancel := context.WithTimeout(w.ctx, 5*time.Second)
			err := conn.Ping(ping)
			pcancel()
			if err != nil {
				return
			}
			// Quiet for a while. Ask the database whether that is because
			// nothing changed or because we stopped hearing about it.
			select {
			case w.ticks <- struct{}{}:
			default: // one pending tick is enough
			}

		default:
			return
		}
	}
}

// dispatch coalesces a burst of notifications into one reload.
func (w *Watcher) dispatch() {
	defer w.wg.Done()

	var pending []string
	var timer <-chan time.Time
	retry := debounceWindow

	for {
		select {
		case <-w.ctx.Done():
			return

		case p := <-w.events:
			pending = append(pending, p)
			if timer == nil {
				timer = time.After(debounceWindow)
			}

		case <-timer:
			payload := strings.Join(pending, ",")
			timer = nil
			// A reload that failed leaves this replica on the old policy, so
			// hold the work and come back for it. Dropping it here is how a
			// cache goes quietly stale.
			if err := w.call(func(ctx context.Context) error { return w.onChange(ctx, payload) }); err != nil {
				log.Errorf("failed to reload the policy after %q, retrying in %s: %v", payload, retry, err)
				w.bump(func(s *WatcherStats) { s.Errors++; s.Retries++ })
				timer = time.After(retry)
				if retry *= 2; retry > maxBackoff {
					retry = maxBackoff
				}
				continue
			}
			pending, retry = nil, debounceWindow
			w.bump(func(s *WatcherStats) { s.Reloads++ })

		case <-w.ticks:
			w.bump(func(s *WatcherStats) { s.Resyncs++ })
			if w.onTick == nil {
				continue
			}
			if err := w.call(w.onTick); err != nil {
				// Same treatment as a failed reload: hold the work and come
				// back for it with backoff. Waiting for the next tick would
				// leave this replica on the old policy for another full
				// resync interval, and again on every failure after that.
				log.Errorf("failed to check the policy version, retrying in %s: %v", retry, err)
				w.bump(func(s *WatcherStats) { s.Errors++; s.Retries++ })
				pending = append(pending, "resync")
				if timer == nil {
					timer = time.After(retry)
				}
			}
		}
	}
}

// call runs a callback under the watcher's own context, so Close cancels the
// database work it is doing instead of waiting for it. The bound is there for
// the case the context cannot help with: a connection that is neither refused
// nor answered.
func (w *Watcher) call(f func(context.Context) error) error {
	ctx, cancel := context.WithTimeout(w.ctx, callbackTimeout)
	defer cancel()
	return f(ctx)
}

func (w *Watcher) record(payload string) {
	w.bump(func(s *WatcherStats) {
		s.Notifications++
		s.LastPayload = payload
		s.LastNotifyAt = time.Now()
		if isSelf(payload, w.origin) {
			s.SelfOriginated++
		}
	})
}

func (w *Watcher) emit(payload string) {
	select {
	case w.events <- payload:
	case <-w.ctx.Done():
	}
}

func (w *Watcher) bump(f func(*WatcherStats)) {
	w.mu.Lock()
	f(&w.stats)
	w.mu.Unlock()
}

// isSelf reports whether this replica caused the notification. The payload is
// "<txid>:<origin>", where origin is the harbor.origin setting each replica
// puts on its own connections.
func isSelf(payload string, origin string) bool {
	_, got, ok := strings.Cut(payload, ":")
	return ok && origin != "" && got == origin
}
