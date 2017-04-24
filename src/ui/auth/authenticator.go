// Copyright (c) 2017 VMware, Inc. All Rights Reserved.
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
	"fmt"
	"time"

	"github.com/vmware/harbor/src/common/models"
	"github.com/vmware/harbor/src/common/utils/log"
	"github.com/vmware/harbor/src/ui/config"
)

var factories = make(map[string]AuthenticatorFactory)

// Authenticator authenticates users according to the principal and credential
type Authenticator interface {
	// Authenticate the user and if success put the user information
	// into the context
	// principal: username/token/session ID/secret
	// credential: password
	Authenticate(ctx context.Context, principal string,
		credential ...string) (context.Context, error)
}

// AuthenticatorFactory is a factory which can create Authenticator
type AuthenticatorFactory interface {
	// Create the Authenticator according to the parameters
	Create(parameters map[string]interface{}) (Authenticator, error)
}

// Register the AuthenticatorFactory to the factories map
func Register(name string, factory AuthenticatorFactory) {
	if _, ok := factories[name]; ok {
		panic(fmt.Sprintf("AuthenticatorFactory of %s already registered", name))
	}

	factories[name] = factory
	log.Infof("AuthenticatorFactory of %s registered", name)
}

// Create an Authenticator accordint to the name and parameters
func Create(name string, parameters map[string]interface{}) (Authenticator, error) {
	factory, ok := factories[name]
	if !ok {
		return nil, fmt.Errorf("AuthenticatorFactory of %s not found", name)
	}
	return factory.Create(parameters)
}

// 1.5 seconds
const frozenTime time.Duration = 1500 * time.Millisecond

var lock = NewUserLock(frozenTime)

// AuthenticatorOld provides interface to authenticate user credentials.
type AuthenticatorOld interface {

	// Authenticate ...
	Authenticate(m models.AuthModel) (*models.User, error)
}

var registry = make(map[string]AuthenticatorOld)

// RegisterOld add different authenticators to registry map.
func RegisterOld(name string, authenticator AuthenticatorOld) {
	if _, dup := registry[name]; dup {
		log.Infof("authenticator: %s has been registered", name)
		return
	}
	registry[name] = authenticator
}

// Login authenticates user credentials based on setting.
func Login(m models.AuthModel) (*models.User, error) {

	authMode, err := config.AuthMode()
	if err != nil {
		return nil, err
	}
	if authMode == "" || m.Principal == "admin" {
		authMode = "db_auth"
	}
	log.Debug("Current AUTH_MODE is ", authMode)

	authenticator, ok := registry[authMode]
	if !ok {
		return nil, fmt.Errorf("Unrecognized auth_mode: %s", authMode)
	}
	if lock.IsLocked(m.Principal) {
		log.Debugf("%s is locked due to login failure, login failed", m.Principal)
		return nil, nil
	}
	user, err := authenticator.Authenticate(m)
	if user == nil && err == nil {
		log.Debugf("Login failed, locking %s, and sleep for %v", m.Principal, frozenTime)
		lock.Lock(m.Principal)
		time.Sleep(frozenTime)
	}
	return user, err
}
