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

package user

import (
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/goharbor/harbor/src/common/rbac"
	"github.com/goharbor/harbor/src/controller/event/metadata/commonevent"
	"github.com/goharbor/harbor/src/pkg/auditext/event"
	notifierevent "github.com/goharbor/harbor/src/pkg/notifier/event"
)

// TestResolvePutUsersCollectionNilData verifies Resolve handles a PUT whose URL
// does not match a specific user id (the /api/v2.0/users collection). In that
// case the base resolver returns without populating event.Data, so Resolve must
// not dereference a nil event.Data.
func TestResolvePutUsersCollectionNilData(t *testing.T) {
	r := &userEventResolver{
		Resolver: event.Resolver{
			ResourceType:      rbac.ResourceUser.String(),
			SucceedCodes:      []int{http.StatusCreated, http.StatusOK},
			ResourceIDPattern: urlPattern,
		},
	}

	ce := &commonevent.Metadata{
		RequestMethod: http.MethodPut,
		RequestURL:    "/api/v2.0/users", // collection path, no /{id}
		ResponseCode:  http.StatusMethodNotAllowed,
	}
	evt := &notifierevent.Event{}

	assert.NotPanics(t, func() {
		err := r.Resolve(ce, evt)
		assert.NoError(t, err)
	})
}
