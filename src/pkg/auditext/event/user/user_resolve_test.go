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
	"context"
	"net/http"
	"testing"

	beegoorm "github.com/beego/beego/v2/client/orm"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	commonmodels "github.com/goharbor/harbor/src/common/models"
	"github.com/goharbor/harbor/src/common/rbac"
	"github.com/goharbor/harbor/src/controller/event/metadata/commonevent"
	"github.com/goharbor/harbor/src/lib/orm"
	"github.com/goharbor/harbor/src/pkg/auditext/event"
	notifierevent "github.com/goharbor/harbor/src/pkg/notifier/event"
	pkgUser "github.com/goharbor/harbor/src/pkg/user"
	usertesting "github.com/goharbor/harbor/src/testing/pkg/user"
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

// TestPreCheckDeleteReusesRequestContext verifies that resolving the user name
// before a DELETE reuses the ormer from the request context. PreCheck runs inside
// the request transaction, so looking the user up through a separate ormer would
// need a second connection from the same pool while the first one is held.
func TestPreCheckDeleteReusesRequestContext(t *testing.T) {
	origMgr := pkgUser.Mgr
	defer func() { pkgUser.Mgr = origMgr }()
	mgr := &usertesting.Manager{}
	pkgUser.Mgr = mgr

	requestOrmer := &beegoorm.DoNothingOrm{}
	ctx := orm.NewContext(context.Background(), requestOrmer)
	var gotCtx context.Context
	mgr.On("Get", mock.Anything, 7).Run(func(args mock.Arguments) {
		gotCtx = args.Get(0).(context.Context)
	}).Return(&commonmodels.User{UserID: 7, Username: "alice"}, nil)

	r := &userEventResolver{
		Resolver: event.Resolver{
			ResourceType:      rbac.ResourceUser.String(),
			SucceedCodes:      []int{http.StatusCreated, http.StatusOK},
			ShouldResolveName: true,
			IDToNameFunc:      userIDToName,
			ResourceIDPattern: urlPattern,
		},
	}

	capture, name := r.PreCheck(ctx, "/api/v2.0/users/7", http.MethodDelete)
	assert.True(t, capture)
	assert.Equal(t, "alice", name)
	mgr.AssertExpectations(t)
	if assert.NotNil(t, gotCtx) {
		o, err := orm.FromContext(gotCtx)
		assert.NoError(t, err)
		assert.Same(t, requestOrmer, o, "user lookup must reuse the ormer of the request context")
	}
}
