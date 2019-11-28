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
	"errors"
	"fmt"
	"math/rand"
	"regexp"
	"strings"
	"sync"
	"time"

	"github.com/casbin/casbin"
	"github.com/casbin/casbin/model"
	"github.com/casbin/casbin/persist"
	"github.com/goharbor/harbor/src/common/utils/log"
)

var (
	errNotImplemented = errors.New("not implemented")
)

// Syntax for models see https://casbin.org/docs/en/syntax-for-models
const modelText = `
# Request definition
[request_definition]
r = sub, obj, act

# Policy definition
[policy_definition]
p = sub, obj, act, eft

# Role definition
[role_definition]
g = _, _

# Policy effect
[policy_effect]
e = some(where (p.eft == allow)) && !some(where (p.eft == deny))

# Matchers
[matchers]
m = g(r.sub, p.sub) && keyMatch2(r.obj, p.obj) && (r.act == p.act || p.act == '*')
`

type regexpStore struct {
	entries sync.Map
}

func (s *regexpStore) Get(key string, build func(string) *regexp.Regexp) *regexp.Regexp {
	value, ok := s.entries.Load(key)
	if !ok {
		value = build(key)
		s.entries.Store(key, value)
	}

	return value.(*regexp.Regexp)
}

func (s *regexpStore) Purge() {
	var keys []interface{}
	s.entries.Range(func(key, value interface{}) bool {
		keys = append(keys, key)
		return true
	})

	for _, key := range keys {
		s.entries.Delete(key)
	}
}

var (
	store = &regexpStore{}
)

func init() {
	startRegexpStorePurging(store, time.Hour*24)
}

func startRegexpStorePurging(s *regexpStore, intervalDuration time.Duration) {
	go func() {
		rand.Seed(time.Now().Unix())
		jitter := time.Duration(rand.Int()%60) * time.Minute
		log.Debugf("Starting regexp store purge in %s", jitter)
		time.Sleep(jitter)

		for {
			s.Purge()
			log.Debugf("Starting regexp store purge in %s", intervalDuration)
			time.Sleep(intervalDuration)
		}
	}()
}

func keyMatch2Build(key2 string) *regexp.Regexp {
	re := regexp.MustCompile(`(.*):[^/]+(.*)`)

	key2 = strings.Replace(key2, "/*", "/.*", -1)
	for {
		if !strings.Contains(key2, "/:") {
			break
		}

		key2 = re.ReplaceAllString(key2, "$1[^/]+$2")
	}

	return regexp.MustCompile("^" + key2 + "$")
}

// keyMatch2 determines whether key1 matches the pattern of key2, its behavior most likely the builtin KeyMatch2
// except that the match of ("/project/1/robot", "/project/1") will return false
func keyMatch2(key1 string, key2 string) bool {
	return store.Get(key2, keyMatch2Build).MatchString(key1)
}

func keyMatch2Func(args ...interface{}) (interface{}, error) {
	name1 := args[0].(string)
	name2 := args[1].(string)

	return bool(keyMatch2(name1, name2)), nil
}

type userAdapter struct {
	User
}

func (a *userAdapter) getRolePolicyLines(role Role) []string {
	lines := []string{}

	roleName := role.GetRoleName()
	// returns empty policy lines if role name is empty
	if roleName == "" {
		return lines
	}

	for _, policy := range role.GetPolicies() {
		line := fmt.Sprintf("p, %s, %s, %s, %s", roleName, policy.Resource, policy.Action, policy.GetEffect())
		lines = append(lines, line)
	}

	return lines
}

func (a *userAdapter) getUserPolicyLines() []string {
	lines := []string{}

	username := a.GetUserName()
	// returns empty policy lines if username is empty
	if username == "" {
		return lines
	}

	for _, policy := range a.GetPolicies() {
		line := fmt.Sprintf("p, %s, %s, %s, %s", username, policy.Resource, policy.Action, policy.GetEffect())
		lines = append(lines, line)
	}

	return lines
}

func (a *userAdapter) getUserAllPolicyLines() []string {
	lines := []string{}

	username := a.GetUserName()
	// returns empty policy lines if username is empty
	if username == "" {
		return lines
	}

	lines = append(lines, a.getUserPolicyLines()...)

	for _, role := range a.GetRoles() {
		lines = append(lines, a.getRolePolicyLines(role)...)
		lines = append(lines, fmt.Sprintf("g, %s, %s", username, role.GetRoleName()))
	}

	return lines
}

func (a *userAdapter) LoadPolicy(model model.Model) error {
	for _, line := range a.getUserAllPolicyLines() {
		persist.LoadPolicyLine(line, model)
	}

	return nil
}

func (a *userAdapter) SavePolicy(model model.Model) error {
	return errNotImplemented
}

func (a *userAdapter) AddPolicy(sec string, ptype string, rule []string) error {
	return errNotImplemented
}

func (a *userAdapter) RemovePolicy(sec string, ptype string, rule []string) error {
	return errNotImplemented
}

func (a *userAdapter) RemoveFilteredPolicy(sec string, ptype string, fieldIndex int, fieldValues ...string) error {
	return errNotImplemented
}

func enforcerForUser(user User) *casbin.Enforcer {
	m := model.Model{}
	m.LoadModelFromText(modelText)

	e := casbin.NewEnforcer(m, &userAdapter{User: user})
	e.AddFunction("keyMatch2", keyMatch2Func)
	return e
}
