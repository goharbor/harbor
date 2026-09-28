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

package role

import (
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/goharbor/harbor/src/common/rbac"
	ctlevent "github.com/goharbor/harbor/src/controller/event"
	"github.com/goharbor/harbor/src/controller/event/metadata/commonevent"
	"github.com/goharbor/harbor/src/controller/event/model"
	"github.com/goharbor/harbor/src/pkg/auditext/event"
	notifierevent "github.com/goharbor/harbor/src/pkg/notifier/event"
)

// newResolver builds a role resolver with a stubbed id->name lookup so the test
// stays hermetic (the real roleIDToName hits the DB).
func newResolver() *event.Resolver {
	return &event.Resolver{
		ResourceType:      rbac.ResourceRole.String(),
		SucceedCodes:      []int{http.StatusCreated, http.StatusOK},
		ShouldResolveName: true,
		IDToNameFunc:      func(string) string { return "myrole" },
		ResourceIDPattern: urlPattern,
	}
}

func resolvedEvent(t *testing.T, ce *commonevent.Metadata) *model.CommonEvent {
	evt := &notifierevent.Event{}
	require.NoError(t, newResolver().Resolve(ce, evt))
	require.Equal(t, ctlevent.TopicCommonEvent, evt.Topic)
	ce2, ok := evt.Data.(*model.CommonEvent)
	require.True(t, ok)
	return ce2
}

// TestRoleResolverRegistered verifies the package registered a resolver for both
// the roles collection and the single-role URL patterns.
func TestRoleResolverRegistered(t *testing.T) {
	resolvers := commonevent.Resolvers()
	assert.NotNil(t, resolvers[`/api/v2.0/roles$`])
	assert.NotNil(t, resolvers[urlPattern])
}

func TestRoleCreateResolved(t *testing.T) {
	e := resolvedEvent(t, &commonevent.Metadata{
		Username:         "admin",
		RequestMethod:    http.MethodPost,
		RequestURL:       "/api/v2.0/roles",
		ResponseLocation: "/api/v2.0/roles/5",
		ResponseCode:     http.StatusCreated,
	})
	assert.Equal(t, "create", e.Operation)
	assert.Equal(t, rbac.ResourceRole.String(), e.ResourceType)
	assert.Equal(t, "myrole", e.ResourceName)
	assert.True(t, e.IsSuccessful)
}

// TestRoleUpdateResolvesCurrentName proves the name is looked up after the
// request, so a rename logs the new name (the bug in the topic-based version,
// which logged the pre-update name).
func TestRoleUpdateResolvesCurrentName(t *testing.T) {
	e := resolvedEvent(t, &commonevent.Metadata{
		Username:      "admin",
		RequestMethod: http.MethodPut,
		RequestURL:    "/api/v2.0/roles/5",
		ResponseCode:  http.StatusOK,
	})
	assert.Equal(t, "update", e.Operation)
	assert.Equal(t, "myrole", e.ResourceName)
	assert.True(t, e.IsSuccessful)
}

// TestRoleFailedRequestRecorded proves a non-success response is still recorded,
// marked unsuccessful.
func TestRoleFailedRequestRecorded(t *testing.T) {
	e := resolvedEvent(t, &commonevent.Metadata{
		Username:      "admin",
		RequestMethod: http.MethodDelete,
		RequestURL:    "/api/v2.0/roles/5",
		ResponseCode:  http.StatusInternalServerError,
	})
	assert.Equal(t, "delete", e.Operation)
	assert.False(t, e.IsSuccessful)
}
