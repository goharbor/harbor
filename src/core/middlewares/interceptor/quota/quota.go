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
	"net/http"

	"github.com/goharbor/harbor/src/common/utils/log"
	"github.com/goharbor/harbor/src/common/utils/redis"
	"github.com/goharbor/harbor/src/core/middlewares/interceptor"
	"github.com/goharbor/harbor/src/pkg/types"
)

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

	opts := qi.opts

	for _, key := range opts.MutexKeys {
		m, err := redis.RequireLock(key)
		if err != nil {
			return err
		}
		qi.mutexes = append(qi.mutexes, m)
	}

	resources := opts.Resources
	if len(resources) == 0 && opts.OnResources != nil {
		resources = opts.OnResources(req)
		log.Debugf("Compute the resources for quota, got: %v", resources)
	}
	qi.resources = resources

	err = qi.reserve()
	if err != nil {
		log.Errorf("Failed to %s resources, error: %v", opts.Action, err)
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
		if opts.OnFulfilled != nil {
			opts.OnFulfilled(w, req)
		}
	default:
		if err := qi.unreserve(); err != nil {
			log.Errorf("Failed to %s resources, error: %v", opts.Action, err)
		}

		if opts.OnRejected != nil {
			opts.OnRejected(w, req)
		}
	}

	if opts.OnFinally != nil {
		opts.OnFinally(w, req)
	}
}

func (qi *quotaInterceptor) freeMutexes() {
	for i := len(qi.mutexes) - 1; i >= 0; i-- {
		if err := redis.FreeLock(qi.mutexes[i]); err != nil {
			log.Error(err)
		}
	}
}

func (qi *quotaInterceptor) reserve() error {
	log.Debugf("Reserve %s resources, %v", qi.opts.Action, qi.resources)

	if len(qi.resources) == 0 {
		return nil
	}

	switch qi.opts.Action {
	case AddAction:
		return qi.opts.Manager.AddResources(qi.resources)
	case SubtractAction:
		return qi.opts.Manager.SubtractResources(qi.resources)
	}

	return nil
}

func (qi *quotaInterceptor) unreserve() error {
	log.Debugf("Unreserve %s resources, %v", qi.opts.Action, qi.resources)

	if len(qi.resources) == 0 {
		return nil
	}

	switch qi.opts.Action {
	case AddAction:
		return qi.opts.Manager.SubtractResources(qi.resources)
	case SubtractAction:
		return qi.opts.Manager.AddResources(qi.resources)
	}

	return nil
}
