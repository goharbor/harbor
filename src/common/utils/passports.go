package utils

import (
	"context"
	"sync"
)

// PassportsPool holds a given number of passports, they can be applied or be revoked. PassportsPool
// is used to control the concurrency of tasks, the pool size determine the max concurrency. When users
// want to start a goroutine to perform some task, they must apply a passport firstly, and after finish
// the task, the passport must be revoked.
type PassportsPool interface {
	// Apply applies a passport from the pool.
	Apply() bool
	// Revoke revokes a passport to the pool
	Revoke() bool
}

type passportsPool struct {
	passports chan struct{}
	stopped   <-chan struct{}
}

// NewPassportsPool creates a passports pool with given size
func NewPassportsPool(size int, stopped <-chan struct{}) PassportsPool {
	return &passportsPool{
		passports: make(chan struct{}, size),
		stopped:   stopped,
	}
}

// Apply applies a passport from the pool. Returning value 'true' means passport acquired
// successfully. If no available passports in the pool, 'Apply' will wait for it. If the
// all passports in the pool are turned into invalid by the 'stopped' channel, then false
// is returned, means no more passports will be dispatched.
func (p *passportsPool) Apply() bool {
	select {
	case p.passports <- struct{}{}:
		return true
	case <-p.stopped:
		return false
	}
}

// Revoke revokes a passport to the pool. Returning value 'true' means passport revoked
// successfully, otherwise 'Revoke' will wait. If pool turns into invalid by 'stopped' channel
// false will be returned.
func (p *passportsPool) Revoke() bool {
	select {
	case <-p.passports:
		return true
	case <-p.stopped:
		return false
	}
}

// LimitedConcurrentRunner is used to run tasks, but limit the max concurrency.
type LimitedConcurrentRunner interface {
	// AddTask adds a task to run
	AddTask(task func() error)
	// Wait waits all the tasks to be finished, returns error if the any of the tasks gets error
	Wait() (err error)
	// Cancel cancels all tasks, tasks that already started will continue to run
	Cancel(err error)
}

type limitedConcurrentRunner struct {
	sync.Mutex
	err           error
	wg            *sync.WaitGroup
	ctx           context.Context
	cancel        context.CancelFunc
	passportsPool PassportsPool
}

// NewLimitedConcurrentRunner creates a runner
func NewLimitedConcurrentRunner(limit int) LimitedConcurrentRunner {
	ctx, cancel := context.WithCancel(context.Background())
	return &limitedConcurrentRunner{
		wg:            new(sync.WaitGroup),
		ctx:           ctx,
		cancel:        cancel,
		passportsPool: NewPassportsPool(limit, ctx.Done()),
	}
}

// AddTask adds a task to run
func (r *limitedConcurrentRunner) AddTask(task func() error) {
	r.wg.Add(1)
	go func() {
		defer func() {
			r.wg.Done()
		}()

		// Return false means no passport acquired, and no valid passport will be dispatched any more.
		// For example, some crucial errors happened and all tasks should be cancelled.
		if ok := r.passportsPool.Apply(); !ok {
			return
		}
		defer func() {
			r.passportsPool.Revoke()
		}()

		err := task()
		if err != nil {
			r.Cancel(err)
		}
	}()
}

// Wait waits all the tasks to be finished
func (r *limitedConcurrentRunner) Wait() (err error) {
	r.wg.Wait()
	return r.err
}

// Cancel cancels all tasks, tasks that already started will continue to run
func (r *limitedConcurrentRunner) Cancel(err error) {
	if err != nil {
		r.Lock()
		defer r.Unlock()
		r.err = err
	}
	r.cancel()
}
