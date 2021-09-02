// Copyright 2018 Project Harbor Authors
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
package auth

import (
	"context"
	"testing"
	"time"

	"github.com/goharbor/harbor/src/common/models"
	"github.com/goharbor/harbor/src/pkg/usergroup/model"
	"github.com/stretchr/testify/assert"
)

var l = NewUserLock(2 * time.Second)

func TestLock(t *testing.T) {
	t.Log("Locking john")
	l.Lock("john")
	if !l.IsLocked("john") {
		t.Errorf("John should be locked")
	}
	t.Log("Locking jack")
	l.Lock("jack")
	t.Log("Sleep for 2 seconds and check...")
	time.Sleep(2 * time.Second)
	if l.IsLocked("jack") {
		t.Errorf("After 2 seconds, jack shouldn't be locked")
	}
	if l.IsLocked("daniel") {
		t.Errorf("daniel has never been locked, he should not be locked")
	}
}

func TestDefaultAuthenticate(t *testing.T) {
	authHelper := DefaultAuthenticateHelper{}
	m := models.AuthModel{}
	user, err := authHelper.Authenticate(context.TODO(), m)
	if user != nil || err == nil {
		t.Fatal("Default implementation should return nil")
	}
}

func TestDefaultOnBoardUser(t *testing.T) {
	user := &models.User{}
	authHelper := DefaultAuthenticateHelper{}
	err := authHelper.OnBoardUser(context.TODO(), user)
	if err == nil {
		t.Fatal("Default implementation should return error")
	}
}

func TestDefaultMethods(t *testing.T) {
	authHelper := DefaultAuthenticateHelper{}
	_, err := authHelper.SearchUser(context.TODO(), "sample")
	if err == nil {
		t.Fatal("Default implementation should return error")
	}

	_, err = authHelper.SearchGroup(context.TODO(), "sample")
	if err == nil {
		t.Fatal("Default implementation should return error")
	}

	err = authHelper.OnBoardGroup(context.TODO(), &model.UserGroup{}, "sample")
	if err == nil {
		t.Fatal("Default implementation should return error")
	}
}

func TestErrAuth(t *testing.T) {
	assert := assert.New(t)
	e := NewErrAuth("test")
	expectedStr := "Failed to authenticate user, due to error 'test'"
	assert.Equal(expectedStr, e.Error())
}
