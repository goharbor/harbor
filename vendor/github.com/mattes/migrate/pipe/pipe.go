// Package pipe has functions for pipe channel handling.
package pipe

import (
	"os"
)

// New creates a new pipe. A pipe is basically a channel.
func New() chan interface{} {
	return make(chan interface{}, 0)
}

// Close closes a pipe and optionally sends an error
func Close(pipe chan interface{}, err error) {
	if err != nil {
		pipe <- err
	}
	close(pipe)
}

// WaitAndRedirect waits for pipe to be closed and
// redirects all messages from pipe to redirectPipe
// while it waits. It also checks if there was an
// interrupt send and will quit gracefully if yes.
func WaitAndRedirect(pipe, redirectPipe chan interface{}, interrupt chan os.Signal) (ok bool) {
	errorReceived := false
	interruptsReceived := 0
	if pipe != nil && redirectPipe != nil {
		for {
			select {

			case <-interrupt:
				interruptsReceived += 1
				if interruptsReceived > 1 {
					os.Exit(5)
				} else {
					// add white space at beginning for ^C splitting
					redirectPipe <- " Aborting after this migration ... Hit again to force quit."
				}

			case item, ok := <-pipe:
				if !ok {
					return !errorReceived && interruptsReceived == 0
				} else {
					redirectPipe <- item
					switch item.(type) {
					case error:
						errorReceived = true
					}
				}
			}
		}
	}
	return !errorReceived && interruptsReceived == 0
}

// ReadErrors selects all received errors and returns them.
// This is helpful for synchronous migration functions.
func ReadErrors(pipe chan interface{}) []error {
	err := make([]error, 0)
	if pipe != nil {
		for {
			select {
			case item, ok := <-pipe:
				if !ok {
					return err
				} else {
					switch item.(type) {
					case error:
						err = append(err, item.(error))
					}
				}
			}
		}
	}
	return err
}
