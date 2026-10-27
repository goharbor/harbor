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
	"context"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/goharbor/harbor/src/lib/orm"
	"github.com/goharbor/harbor/src/pkg/quota"
	"github.com/goharbor/harbor/src/pkg/quota/driver"
	"github.com/goharbor/harbor/src/pkg/quota/types"
	ormtesting "github.com/goharbor/harbor/src/testing/lib/orm"
	"github.com/goharbor/harbor/src/testing/mock"
	quotatesting "github.com/goharbor/harbor/src/testing/pkg/quota"
	drivertesting "github.com/goharbor/harbor/src/testing/pkg/quota/driver"
)

func drainDirty() {
	dirty.Lock()
	dirty.keys = map[refKey]struct{}{}
	dirty.Unlock()
}

func dirtyLen() int {
	dirty.Lock()
	defer dirty.Unlock()
	return len(dirty.keys)
}

func TestAsyncRefreshDisabledByDefault(t *testing.T) {
	if os.Getenv(asyncRefreshDurationEnv) != "" {
		// asyncRefreshConfigured is fixed at init time, so an environment
		// that legitimately enables async refresh cannot satisfy this test
		t.Skipf("%s is set in this environment", asyncRefreshDurationEnv)
	}
	assert.False(t, AsyncRefreshEnabled())
}

func TestParseRefreshInterval(t *testing.T) {
	cases := []struct {
		name       string
		env        string
		interval   time.Duration
		configured bool
	}{
		{"unset", "", defaultDeferredRefreshInterval, false},
		{"valid", "60", 60 * time.Second, true},
		{"one second", "1", time.Second, true},
		{"zero", "0", defaultDeferredRefreshInterval, false},
		{"negative", "-5", defaultDeferredRefreshInterval, false},
		{"not a number", "10s", defaultDeferredRefreshInterval, false},
		{"largest representable", "9223372036", 9223372036 * time.Second, true},
		{"overflows time.Duration", "9223372037", defaultDeferredRefreshInterval, false},
		{"overflows int64", "99999999999999999999", defaultDeferredRefreshInterval, false},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			interval, configured := parseRefreshInterval(c.env)
			assert.Equal(t, c.interval, interval)
			assert.Equal(t, c.configured, configured)
		})
	}
}

func TestMarkRefreshCoalesces(t *testing.T) {
	drainDirty()
	defer drainDirty()

	for i := 0; i < 100; i++ {
		MarkRefresh("project", "1")
	}
	MarkRefresh("project", "2")

	assert.Equal(t, 2, dirtyLen(), "repeated marks of the same reference must coalesce into one")
}

func TestRefreshDirtyEmptySetIsNoop(t *testing.T) {
	drainDirty()
	refreshDirty(context.TODO()) // must not panic, no DB access happens
	assert.Equal(t, 0, dirtyLen())
}

func TestRefreshDirtyFlushesOncePerReference(t *testing.T) {
	drainDirty()
	defer drainDirty()

	reference := "mock-async"
	d := &drivertesting.Driver{}
	driver.Register(reference, d)

	quotaMgr := &quotatesting.Manager{}
	origCtl := Ctl
	Ctl = &controller{quotaMgr: quotaMgr}
	defer func() { Ctl = origCtl }()

	hardLimits := types.ResourceList{types.ResourceStorage: types.UNLIMITED}
	// return a fresh object per call: updateUsageByDB mutates the returned
	// quota, and a shared instance would make the second refresh a no-op
	mock.OnAnything(quotaMgr, "GetByRef").Return(func(context.Context, string, string) *quota.Quota {
		return &quota.Quota{Hard: hardLimits.String(), Used: types.ResourceList{types.ResourceStorage: 0}.String()}
	}, nil)
	mock.OnAnything(d, "CalculateUsage").Return(types.ResourceList{types.ResourceStorage: 42}, nil)
	mock.OnAnything(quotaMgr, "Update").Return(nil)

	// many marks for the same reference, one for another
	for i := 0; i < 50; i++ {
		MarkRefresh(reference, "1")
	}
	MarkRefresh(reference, "2")

	refreshDirty(orm.NewContext(context.TODO(), &ormtesting.FakeOrmer{}))

	// one recompute+store per distinct reference, not per mark
	quotaMgr.AssertNumberOfCalls(t, "Update", 2)
	assert.Equal(t, 0, dirtyLen(), "successful flush must clear the dirty set")
}

func TestRefreshDirtyRetriesFailedFlush(t *testing.T) {
	drainDirty()
	defer drainDirty()

	reference := "mock-async-fail"
	d := &drivertesting.Driver{}
	driver.Register(reference, d)

	quotaMgr := &quotatesting.Manager{}
	origCtl := Ctl
	Ctl = &controller{quotaMgr: quotaMgr}
	defer func() { Ctl = origCtl }()

	mock.OnAnything(quotaMgr, "GetByRef").Return(nil, assert.AnError)

	MarkRefresh(reference, "1")
	refreshDirty(orm.NewContext(context.TODO(), &ormtesting.FakeOrmer{}))

	assert.Equal(t, 1, dirtyLen(), "failed flush must re-mark the reference for the next interval")
}
