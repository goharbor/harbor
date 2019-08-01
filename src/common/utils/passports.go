package utils

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
