package clockwork

import (
	"sync/atomic"
	"time"
)

// Timer provides an interface which can be used instead of directly
// using the timer within the time module. The real-time timer t
// provides events through t.C which becomes now t.Chan() to make
// this channel requirement definable in this interface.
type Timer interface {
	Chan() <-chan time.Time
	Reset(d time.Duration) bool
	Stop() bool
}

type realTimer struct {
	*time.Timer
}

func (r realTimer) Chan() <-chan time.Time {
	return r.C
}

type fakeTimer struct {
	c       chan time.Time
	clock   FakeClock
	stop    chan struct{}
	reset   chan reset
	stopped uint32
}

func (f *fakeTimer) Chan() <-chan time.Time {
	return f.c
}

func (f *fakeTimer) Reset(d time.Duration) bool {
	stopped := f.Stop()

	f.reset <- reset{t: f.clock.Now().Add(d), next: f.clock.After(d)}
	if d > 0 {
		atomic.StoreUint32(&f.stopped, 0)
	}

	return stopped
}

func (f *fakeTimer) Stop() bool {
	if atomic.CompareAndSwapUint32(&f.stopped, 0, 1) {
		f.stop <- struct{}{}
		return true
	}
	return false
}

type reset struct {
	t    time.Time
	next <-chan time.Time
}

// run initializes a background goroutine to send the timer event to the timer channel
// after the period. Events are discarded if the underlying ticker channel does not have
// enough capacity.
func (f *fakeTimer) run(initialDuration time.Duration) {
	nextTick := f.clock.Now().Add(initialDuration)
	next := f.clock.After(initialDuration)

	waitForReset := func() (time.Time, <-chan time.Time) {
		for {
			select {
			case <-f.stop:
				continue
			case r := <-f.reset:
				return r.t, r.next
			}
		}
	}

	go func() {
		for {
			select {
			case <-f.stop:
			case <-next:
				atomic.StoreUint32(&f.stopped, 1)
				select {
				case f.c <- nextTick:
				default:
				}

				next = nil
			}

			nextTick, next = waitForReset()
		}
	}()
}
