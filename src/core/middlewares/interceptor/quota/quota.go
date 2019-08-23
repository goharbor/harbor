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

package quota

import (
	"fmt"
	"math/rand"
	"net/http"
	"time"

	"github.com/goharbor/harbor/src/common/utils/log"
	"github.com/goharbor/harbor/src/common/utils/redis"
	"github.com/goharbor/harbor/src/core/middlewares/interceptor"
	"github.com/goharbor/harbor/src/pkg/types"
)

func init() {
	rand.Seed(time.Now().UnixNano())
}

// New ....
func New(opts ...Option) interceptor.Interceptor {
	options := newOptions(opts...)

	return &quotaInterceptor{opts: &options}
}

type statusRecorder interface {
	Status() int
}

type quotaInterceptor struct {
	opts      *Options
	resources types.ResourceList
	mutexes   []*redis.Mutex
}

// HandleRequest ...
func (qi *quotaInterceptor) HandleRequest(req *http.Request) (err error) {
	defer func() {
		if err != nil {
			qi.freeMutexes()
		}
	}()

	err = qi.requireMutexes()
	if err != nil {
		return
	}

	err = qi.computeResources(req)
	if err != nil {
		return
	}

	err = qi.doTry()
	if err != nil {
		log.Errorf("Failed to %s resources, error: %v", qi.opts.Action, err)
	}

	return
}

// HandleResponse ...
func (qi *quotaInterceptor) HandleResponse(w http.ResponseWriter, req *http.Request) {
	defer qi.freeMutexes()

	sr, ok := w.(statusRecorder)
	if !ok {
		return
	}

	opts := qi.opts

	switch sr.Status() {
	case opts.StatusCode:
		if err := qi.doConfirm(); err != nil {
			log.Errorf("Failed to confirm for resource, error: %v", err)
		}

		if opts.OnFulfilled != nil {
			if err := opts.OnFulfilled(w, req); err != nil {
				log.Errorf("Failed to handle on fulfilled, error: %v", err)
			}
		}
	default:
		if err := qi.doCancel(); err != nil {
			log.Errorf("Failed to cancel for resource, error: %v", err)
		}

		if opts.OnRejected != nil {
			if err := opts.OnRejected(w, req); err != nil {
				log.Errorf("Failed to handle on rejected, error: %v", err)
			}
		}
	}

	if opts.OnFinally != nil {
		if err := opts.OnFinally(w, req); err != nil {
			log.Errorf("Failed to handle on finally, error: %v", err)
		}
	}
}

func (qi *quotaInterceptor) requireMutexes() error {
	if !qi.opts.EnforceResources() {
		// Do nothing for locks when quota interceptor not enforce resources
		return nil
	}

	for _, key := range qi.opts.MutexKeys {
		m, err := redis.RequireLock(key)
		if err != nil {
			return err
		}
		qi.mutexes = append(qi.mutexes, m)
	}

	return nil
}

func (qi *quotaInterceptor) freeMutexes() {
	for i := len(qi.mutexes) - 1; i >= 0; i-- {
		if err := redis.FreeLock(qi.mutexes[i]); err != nil {
			log.Error(err)
		}
	}
}

func (qi *quotaInterceptor) computeResources(req *http.Request) error {
	if !qi.opts.EnforceResources() {
		// Do nothing in compute resources when quota interceptor not enforce resources
		return nil
	}

	qi.resources = qi.opts.Resources
	if len(qi.resources) == 0 && qi.opts.OnResources != nil {
		resources, err := qi.opts.OnResources(req)
		if err != nil {
			return fmt.Errorf("failed to compute the resources for quota, error: %v", err)
		}

		qi.resources = resources
	}

	return nil
}

func (qi *quotaInterceptor) doTry() error {
	if !qi.opts.EnforceResources() {
		// Do nothing in try stage when quota interceptor not enforce resources
		return nil
	}

	// Add resources in try stage when it is add action
	// And do nothing in confirm stage for add action
	if len(qi.resources) != 0 && qi.opts.Action == AddAction {
		return qi.opts.Manager.AddResources(qi.resources)
	}

	return nil
}

func (qi *quotaInterceptor) doConfirm() error {
	if !qi.opts.EnforceResources() {
		// Do nothing in confirm stage when quota interceptor not enforce resources
		return nil
	}

	// Subtract resources in confirm stage when it is subtract action
	// And do nothing in try stage for subtract action
	if len(qi.resources) != 0 && qi.opts.Action == SubtractAction {
		return retry(3, 100*time.Millisecond, func() error {
			return qi.opts.Manager.SubtractResources(qi.resources)
		})
	}

	return nil
}

func (qi *quotaInterceptor) doCancel() error {
	if !qi.opts.EnforceResources() {
		// Do nothing in cancel stage when quota interceptor not enforce resources
		return nil
	}

	// Subtract resources back when process failed for add action
	if len(qi.resources) != 0 && qi.opts.Action == AddAction {
		return retry(3, 100*time.Millisecond, func() error {
			return qi.opts.Manager.SubtractResources(qi.resources)
		})
	}

	return nil
}

func retry(attempts int, sleep time.Duration, f func() error) error {
	if err := f(); err != nil {
		if attempts--; attempts > 0 {
			r := time.Duration(rand.Int63n(int64(sleep)))
			sleep = sleep + r/2

			time.Sleep(sleep)
			return retry(attempts, 2*sleep, f)
		}
		return err
	}

	return nil
}
