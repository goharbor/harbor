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

package rbac

import (
	"fmt"
	"sync"

	"github.com/casbin/casbin"
)

// Evaluator the permissioin evaluator
type Evaluator interface {
	// HasPermission returns true when user has action permission for the resource
	HasPermission(resource Resource, action Action) bool
}

type userEvaluator struct {
	user     User
	enforcer *casbin.Enforcer
	once     sync.Once
}

func (e *userEvaluator) HasPermission(resource Resource, action Action) bool {
	e.once.Do(func() {
		e.enforcer = enforcerForUser(e.user)
	})

	return e.enforcer.Enforce(e.user.GetUserName(), resource.String(), action.String())
}

// NewUserEvaluator returns Evaluator for the rbac user
func NewUserEvaluator(user User) Evaluator {
	return &userEvaluator{
		user: user,
	}
}

type namespaceEvaluator struct {
	factory       func(Namespace) Evaluator
	namespaceKind string
	cache         sync.Map
}

func (e *namespaceEvaluator) HasPermission(resource Resource, action Action) bool {
	ns, err := resource.GetNamespace()
	if err == nil && ns.Kind() == e.namespaceKind {
		var evaluator Evaluator

		key := fmt.Sprintf("%s:%v", ns.Kind(), ns.Identity())
		value, ok := e.cache.Load(key)
		if !ok {
			evaluator = e.factory(ns)
			e.cache.Store(key, evaluator)
		} else {
			evaluator, _ = value.(Evaluator) // maybe value is nil
		}

		return evaluator != nil && evaluator.HasPermission(resource, action)
	}

	return false
}

// NewNamespaceEvaluator returns permission evaluator for which support namespace
func NewNamespaceEvaluator(namespaceKind string, factory func(Namespace) Evaluator) Evaluator {
	return &namespaceEvaluator{
		namespaceKind: namespaceKind,
		factory:       factory,
	}
}
