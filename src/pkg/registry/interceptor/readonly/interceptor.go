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

package readonly

import (
	"context"
	"errors"
	"net/http"
	"time"

	"github.com/goharbor/harbor/src/lib/cache"
	"github.com/goharbor/harbor/src/lib/cache/memory"
	"github.com/goharbor/harbor/src/lib/config"
	itcp "github.com/goharbor/harbor/src/pkg/registry/interceptor"
)

// Err indicates the system is in read only mode
var (
	Err = errors.New("the system is in read only mode, cancel the request")
	key = "read-only"
)

// NewInterceptor creates an interceptor that intercepts any requests if the system is set to read-only
func NewInterceptor() itcp.Interceptor {
	// ignore the error as the New return nil error
	cache, _ := memory.New(cache.Options{
		Expiration: 5 * time.Second,
		Codec:      cache.DefaultCodec(),
	})
	return &interceptor{cache: cache}
}

type interceptor struct {
	cache cache.Cache
}

func (i *interceptor) Intercept(req *http.Request) error {
	switch req.Method {
	case http.MethodGet, http.MethodHead, http.MethodOptions:
		return nil
	}

	isReadOnly, err := i.isReadOnly(req.Context())
	if err != nil {
		return err
	}
	if isReadOnly {
		return Err
	}

	return nil
}

func (i *interceptor) isReadOnly(ctx context.Context) (bool, error) {
	var (
		isReadOnly bool
		err        error
	)
	// return the cached value if exists
	if err = i.cache.Fetch(ctx, key, &isReadOnly); err == nil {
		return isReadOnly, nil
	}
	if err != cache.ErrNotFound {
		return false, err
	}

	// no cache, get the config via API
	if err := config.Load(ctx); err != nil {
		return false, err
	}
	isReadOnly = config.ReadOnly(ctx)
	if err := i.cache.Save(ctx, key, &isReadOnly); err != nil {
		return false, err
	}
	return isReadOnly, nil
}
