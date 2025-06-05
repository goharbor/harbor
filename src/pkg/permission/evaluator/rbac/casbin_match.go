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
	"math/rand"
	"regexp"
	"strings"
	"sync"
	"time"

	"github.com/goharbor/harbor/src/lib/log"
)

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
	var keys []any
	s.entries.Range(func(key, _ any) bool {
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
		rand.NewSource(time.Now().UnixNano())
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

func keyMatch2Func(args ...any) (any, error) {
	name1 := args[0].(string)
	name2 := args[1].(string)

	return keyMatch2(name1, name2), nil
}
