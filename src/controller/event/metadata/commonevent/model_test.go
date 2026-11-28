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

package commonevent

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/goharbor/harbor/src/pkg/notifier/event"
)

type fakeResolver struct{}

func (fakeResolver) Resolve(*Metadata, *event.Event) error { return nil }

func (fakeResolver) PreCheck(context.Context, string, string) (bool, string) {
	return true, "fake"
}

func TestRegisterResolverInvalidPatternLeavesRegistryUnchanged(t *testing.T) {
	const invalid = `^/api/v2.0/(`
	const valid = `^/api/v2.0/commonevent-test$`
	t.Cleanup(func() {
		delete(registrations, invalid)
		delete(registrations, valid)
	})

	assert.Panics(t, func() { RegisterResolver(invalid, fakeResolver{}) })
	require.True(t, mu.TryLock(), "registry lock still held after panic")
	mu.Unlock()
	assert.NotContains(t, registrations, invalid)

	RegisterResolver(valid, fakeResolver{})
	m := &Metadata{Ctx: context.Background(), RequestURL: "/api/v2.0/commonevent-test"}
	matched, name := m.PreCheckMetadata()
	assert.True(t, matched)
	assert.Equal(t, "fake", name)
}

func TestResolversReturnsCopy(t *testing.T) {
	const probe = `^/api/v2.0/commonevent-copy$`
	t.Cleanup(func() {
		delete(registrations, probe)
	})

	Resolvers()[probe] = fakeResolver{}
	assert.NotContains(t, registrations, probe)

	m := &Metadata{Ctx: context.Background(), RequestURL: "/v2/library/nginx/blobs/sha256:abc"}
	assert.NotPanics(t, func() { m.PreCheckMetadata() })
}
