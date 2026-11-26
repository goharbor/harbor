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
		mu.Lock()
		defer mu.Unlock()
		delete(urlResolvers, valid)
		delete(urlPatterns, valid)
	})

	assert.Panics(t, func() { RegisterResolver(invalid, fakeResolver{}) })
	assert.NotContains(t, urlResolvers, invalid)
	assert.NotContains(t, urlPatterns, invalid)

	RegisterResolver(valid, fakeResolver{})
	m := &Metadata{Ctx: context.Background(), RequestURL: "/api/v2.0/commonevent-test"}
	matched, name := m.PreCheckMetadata()
	assert.True(t, matched)
	assert.Equal(t, "fake", name)
}
